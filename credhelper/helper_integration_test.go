// +build !unit

// Copyright 2017 Google, Inc.
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

package credhelper

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	"github.com/GoogleCloudPlatform/docker-credential-gcr/store"
	"github.com/docker/docker-credential-helpers/credentials"
)

var expectedGcrHosts = [...]string{
	"gcr.io",
	"us.gcr.io",
	"eu.gcr.io",
	"asia.gcr.io",
	"staging-k8s.gcr.io",
	"marketplace.gcr.io",
}

var testCredStorePath = filepath.Clean("helper_test_cred_store.json")

func TestMain(m *testing.M) {
	err := cleanUp()
	if err != nil {
		panic(fmt.Sprintf("Unable to clean the test environment, test results cannot be trusted: %v", err))
	}
	exitCode := m.Run()

	// be polite and clean up
	cleanUp()
	os.Exit(exitCode)
}

// cleanUp eliminates test artifacts.
func cleanUp() error {
	err := os.Remove(testCredStorePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func TestList_NoCredFile(t *testing.T) {
	err := cleanUp()
	if err != nil {
		t.Fatal("Could not clean the test environment.")
	}
	store := store.NewGCRCredStore(testCredStorePath)
	userCfg, err := config.LoadUserConfig()
	if err != nil {
		t.Fatalf("Could not load user config file: %v", err)
	}

	helper := NewGCRCredentialHelper(store, userCfg)

	creds, err := helper.List()

	if err != nil {
		t.Fatalf("Error listing credentials: %v", err)
	}

	if len(expectedGcrHosts) != len(creds) {
		t.Fatalf("Expected %d credentials, got %d", len(expectedGcrHosts), len(creds))
	}
	for _, host := range expectedGcrHosts {
		if username := creds[host]; username != expectedGCRUsername {
			t.Errorf("Expected username to be %s for host %s, was %s", expectedGCRUsername, host, username)
		}
	}
}

func TestList_CredFileExists(t *testing.T) {
	err := cleanUp()
	if err != nil {
		t.Fatal("Could not clean the test environment.")
	}

	// Store some 3P creds in a credential file.
	store := store.NewGCRCredStore(testCredStorePath)
	expected3PRegistry := "https://coolreg.io"
	expected3PUsername := "fancybear"
	store.SetOtherCreds(&credentials.Credentials{
		ServerURL: expected3PRegistry,
		Username:  expected3PUsername,
		Secret:    "спасиба",
	})

	userCfg, err := config.LoadUserConfig()
	if err != nil {
		t.Fatalf("Could not load user config file: %v", err)
	}

	helper := NewGCRCredentialHelper(store, userCfg)

	creds, err := helper.List()

	if err != nil {
		t.Fatalf("Error listing credentials: %v", err)
	}

	if len(expectedGcrHosts)+1 != len(creds) {
		t.Fatalf("Expected %d credentials, got %d", len(expectedGcrHosts), len(creds))
	}
	for _, host := range expectedGcrHosts {
		if username := creds[host]; username != expectedGCRUsername {
			t.Errorf("Expected username to be %s for host %s, was %s", expectedGCRUsername, host, username)
		}
	}
	if username := creds[expected3PRegistry]; username != expected3PUsername {
		t.Fatal("Failed to include 3P credentials")
	}

}
