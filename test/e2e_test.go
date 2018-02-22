// +build !unit

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
	"strings"
	"testing"
	"time"

	"github.com/docker/docker-credential-helpers/credentials"
)

const (
	thirdPartyRegistry = "www.wherever.test"
	expectedUsername   = "u$ern4me"
	expectedSecret     = "$ecr3t"

	gcrRegistry     = "us.gcr.io"
	gcrAccessToken  = "gcr.access.token"
	gcrRefreshToken = "gcr.refresh.token"
)

func TestEndToEnd_ThirdPartyCreds(t *testing.T) {
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

	// store some creds
	helper = helperCmd([]string{"store"})
	payload := fmt.Sprintf(`{
		"ServerURL": "%s",
		"Username": "%s",
		"Secret": "%s"
	}`, thirdPartyRegistry, expectedUsername, expectedSecret)
	helper.Stdin = strings.NewReader(payload)
	if err = helper.Run(); err != nil {
		t.Fatalf("store failed: %v", err)
	}

	// retrieve the stored credentials
	helper = helperCmd([]string{"get"})
	var out bytes.Buffer
	helper.Stdout = &out
	helper.Stdin = strings.NewReader(thirdPartyRegistry)
	if err = helper.Run(); err != nil {
		t.Fatalf("get failed: %v", err)
	}

	// Verify the credentials
	var creds credentials.Credentials
	if err := json.NewDecoder(bytes.NewReader(out.Bytes())).Decode(&creds); err != nil {
		t.Fatalf("Unable to decode credentials returned from get: %v", err)
	}
	if creds.Username != expectedUsername {
		t.Errorf("Expected username: %s, was: %s", expectedUsername, creds.Username)
	}
	if creds.Secret != expectedSecret {
		t.Errorf("Expected secret: %s, was: %s", expectedSecret, creds.Secret)
	}

	// erase the credentials
	helper = helperCmd([]string{"erase"})
	helper.Stdin = strings.NewReader(thirdPartyRegistry)
	if err = helper.Run(); err != nil {
		t.Fatalf("erase failed: %v", err)
	}

	// verify erasure
	helper = helperCmd([]string{"get"})
	out.Reset()
	helper.Stdout = &out
	helper.Stdin = strings.NewReader(thirdPartyRegistry)
	if err = helper.Run(); err == nil {
		t.Fatalf("get succeeded, returned: %s", out.String())
	}
}

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
	if creds.Username != "oauth2accesstoken" {
		t.Errorf("Bad GCR username: Wanted: %s, Got: %s", "oauth2accesstoken", creds.Username)
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
