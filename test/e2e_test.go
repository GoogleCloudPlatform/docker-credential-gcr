//go:build !unit && !gazelle
// +build !unit,!gazelle

// Copyright 2016 Google, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker-credential-helpers/credentials"
)

const (
	gcrRegistry     = "us.gcr.io"
	gcrAccessToken  = "gcr.access.token"
	gcrRefreshToken = "gcr.refresh.token"
)

func writeValidGCRCreds(accessToken, refreshToken string) error {
	oneHourFromNow := time.Now().Add(time.Hour)
	expiryJSON, err := oneHourFromNow.MarshalJSON()
	if err != nil {
		return err
	}
	credsJSON := fmt.Sprintf(`{"gcrCreds":{"access_token":"%s","refresh_token":"%s","token_expiry":%s}}`, accessToken, refreshToken, expiryJSON)
	return setTestCredentialFileContents(credsJSON)
}

func createTestCredentialFile() (*os.File, error) {
	credStorePath, err := testCredStorePath()
	if err != nil {
		return nil, err
	}

	// create the credential path, if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(credStorePath), 0777); err != nil {
		return nil, err
	}

	// create the credential file, or truncate (clear) it if it exists
	f, err := os.Create(credStorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create test credential file: %v", err)
	}
	return f, nil
}

func setTestCredentialFileContents(contents string) error {
	f, err := createTestCredentialFile()
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.WriteString(f, contents)
	return err
}

func TestEndToEnd_GCRCreds(t *testing.T) {
	err := initTestEnvironment()
	if err != nil {
		t.Fatalf("Could not initialize test environment: %v", err)
	}
	// Sanity test to verify that the environment is set up correctly.
	assertTestEnv(t)

	// Configure the helper to only use the private credential store.
	helper := helperCmd([]string{"config", "--token-source=store"})
	if err := helper.Run(); err != nil {
		t.Fatalf("Failed to configure the helper: %v", err)
	}
	// Verify the contents of the config.
	configPath, err := testConfigPath()
	if err != nil {
		t.Fatalf("Unable construct test config path: %v", err)
	}
	configBuf, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Unable to verify config: %v", err)
	} else if configStr := string(configBuf); strings.TrimSpace(configStr) != `{"TokenSources":["store"]}` {
		t.Fatalf("Expected config: %s, was: %s", `{"TokenSources":["store"]}`, configStr)
	}

	if err := writeValidGCRCreds(gcrAccessToken, gcrRefreshToken); err != nil {
		t.Fatalf("Unable to write creds store: %#v", err)
	}

	// retrieve the stored credentials
	helper = helperCmd([]string{"get"})
	var out bytes.Buffer
	helper.Stdout = &out
	helper.Stdin = strings.NewReader(gcrRegistry)
	if err = helper.Run(); err != nil {
		t.Fatalf("`get` failed: %v, Stdout: %s", err, string(out.Bytes()))
	}

	// Verify the credentials
	var creds credentials.Credentials
	if err := json.NewDecoder(bytes.NewReader(out.Bytes())).Decode(&creds); err != nil {
		t.Fatalf("Unable to decode credentials returned from get: %v", err)
	}

	if match, err := regexp.MatchString("_dcgcr_(?:[0-9]+_)*token", creds.Username); !match || err != nil {
		// Fail if not a dev version.
		devUsername := "_dcgcr__token"
		if creds.Username != devUsername {
			t.Errorf("Bad GCR username: Wanted: %q, Got (val, err): %q, %v", "_dcgcr_(?:[0-9]+_)*token", creds.Username, err)
		}
	}
	if creds.Secret != gcrAccessToken {
		t.Errorf("Bad GCR access token. Wanted: %s, Got: %s", gcrAccessToken, creds.Secret)
	}

	// erase the credentials
	helper = helperCmd([]string{"erase"})
	helper.Stdin = strings.NewReader(gcrRegistry)
	if err = helper.Run(); err == nil {
		t.Fatal("Expected erase to fail for GCR hostname.")
	}
}
