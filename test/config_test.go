// +build surface

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
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
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

	// Unset everything.
	helper = helperCmd([]string{"config", "--unset-all"})
	if err := helper.Run(); err != nil {
		t.Fatalf("Failed to un-configure the helper: %v", err)
	}
	if _, err = os.Stat(configPath); err == nil {
		t.Fatal("Config file still present.")
	}
}
