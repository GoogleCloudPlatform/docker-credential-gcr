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

package config

import (
	"os"
	"strings"
	"testing"
)

const (
	expectedConfigEnvVar = "DOCKER_CREDENTIAL_GCR_CONFIG"
	expectedEnvPath      = "whatever"
	expectedFilename     = "docker_credential_gcr_config.json"
)

var expctedDefaultTokSrcs = []string{"store", "env"}

func assertEqual(t *testing.T, expected, actual []string) {
	if (expected == nil && actual != nil) || (expected != nil && actual == nil) {
		t.Fatalf("Expected: %v, Actual: %v", expected, actual)
	}

	if len(expected) != len(actual) {
		t.Fatalf("Expected: %v, Actual: %v", expected, actual)
	}

	for i := range expected {
		if expected[i] != actual[i] {
			t.Fatalf("Expected: %v, Actual: %v", expected, actual)
		}
	}
}

func TestConfigPath_RespectsEnvVar(t *testing.T) {
	old := os.Getenv(expectedConfigEnvVar)
	os.Setenv(expectedConfigEnvVar, expectedEnvPath)
	defer os.Setenv(expectedConfigEnvVar, old)

	result, err := configPath()

	if err != nil {
		t.Fatalf("Could not retrieve config path: %v", err)
	} else if result != expectedEnvPath {
		t.Fatalf("Expected config path to be: %s, was: %s", expectedEnvPath, result)
	}
}

func TestConfigPath_Sanity(t *testing.T) {
	result, err := configPath()

	if err != nil {
		t.Fatalf("Could not retrieve config path: %v", err)
	} else if !strings.HasSuffix(result, expectedFilename) {
		t.Fatalf("Expected config path to end in: %s, was: %s", expectedFilename, result)
	}
}

func TestTokenSources_ReturnsDefaultWhenUnset(t *testing.T) {
	tested := &configFile{
		TokenSrcs: nil,
	}

	result := tested.TokenSources()

	assertEqual(t, expctedDefaultTokSrcs, result)
}

func TestTokenSources_ReturnsDefaultWhenEmpty(t *testing.T) {
	tested := &configFile{
		TokenSrcs: []string{},
	}

	result := tested.TokenSources()

	assertEqual(t, expctedDefaultTokSrcs, result)
}

func TestTokenSources_UserDefined(t *testing.T) {
	expected := []string{"env", "store", "gcloud"}
	tested := &configFile{
		TokenSrcs: expected,
	}

	actual := tested.TokenSources()

	assertEqual(t, expected, actual)
}

func TestSetTokenSources(t *testing.T) {
	expected := []string{"gcloud"}
	tested := &configFile{
		persist: func(c *configFile) error {
			if !equal(expected, c.TokenSrcs) {
				t.Errorf("Expected: %v, Actual %v", expected, c.TokenSrcs)
			}
			return nil
		},
	}

	tested.SetTokenSources(expected)
}

func TestEqual(t *testing.T) {
	if !equal(nil, nil) {
		t.Error("!equal(nil, nil)")
	}
	if equal(nil, []string{}) {
		t.Error("equal(nil, []string{})")
	}
	if equal([]string{"something"}, nil) {
		t.Error(`equal([]string{"something"}, nil)`)
	}
	if equal([]string{}, []string{"something"}) {
		t.Error(`equal([]string{}, []string{"something"})`)
	}
	if equal([]string{"something"}, []string{"something else"}) {
		t.Error(`equal([]string{"something"}, []string{"something else"})`)
	}
	if !equal([]string{"equal"}, []string{"equal"}) {
		t.Error(`!equal([]string{"equal"}, []string{"equal"})`)
	}
	if equal([]string{"equal"}, []string{"equal", "notreally"}) {
		t.Error(`equal([]string{"equal"}, []string{"equal", "notreally"})`)
	}
	if equal([]string{"equal", "forsure"}, []string{"equal", "notreally"}) {
		t.Error(`equal([]string{"equal","forsure"}, []string{"equal", "notreally"})`)
	}
}

func TestDefaultTokenSources(t *testing.T) {
	// The exact contents and ordering are important, any changes to default
	// ordering require user notification.
	assertEqual(t, expctedDefaultTokSrcs, DefaultTokenSources[:])
}
