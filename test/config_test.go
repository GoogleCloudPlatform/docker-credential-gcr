//go:build !unit && !gazelle && !windows
// +build !unit,!gazelle,!windows

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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
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

// getDockerConfig returns the docker config.
func getDockerConfig() (*configfile.ConfigFile, error) {
	dockerConfig, err := cliconfig.Load(cliconfig.Dir())
	if err != nil {
		return nil, fmt.Errorf("unable to load docker config: %v", err)
	}
	return dockerConfig, nil
}

// deleteDockerConfig deletes the docker config.
func deleteDockerConfig() error {
	filename := filepath.Join(cliconfig.Dir(), cliconfig.ConfigFileName)

	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return os.Remove(filename)
}

func TestConfigureDocker(t *testing.T) {
	if err := deleteDockerConfig(); err != nil {
		if err != nil {
			t.Fatalf("Failed to delete the pre-existing docker config: %v", err)
		}
	}
	dockerConfig, err := getDockerConfig()
	if err != nil {
		t.Fatalf("Failed to get the docker config: %#v", err)
	}

	if len(dockerConfig.CredentialHelpers) > 0 {
		t.Fatal("Failed to clean the docker config.")
	}

	// Configure docker.
	helper := helperCmd([]string{"configure-docker", "--overwrite"})
	var out bytes.Buffer
	helper.Stdout = &out
	helper.Stderr = os.Stderr
	if err = helper.Run(); err != nil {
		t.Fatalf("Failed to execute `configure-docker --overwrite`: %v Stdout: %s", err, string(out.Bytes()))
	}

	dockerConfig, err = getDockerConfig()
	if err != nil {
		t.Fatalf("Failed to get the docker config: %v", err)
	}

	if len(dockerConfig.CredentialHelpers) == 0 {
		t.Fatal("CredentialHelpers was empty.")
	}

	for _, registryHostname := range config.DefaultGCRRegistries {
		if helperSuffix, ok := dockerConfig.CredentialHelpers[registryHostname]; ok {
			if helperSuffix != "gcr" {
				t.Errorf("Wanted value for %s in dockerConfig.CredentialHelpers to be %s, got %s", registryHostname, "gcr", helperSuffix)
			}
		} else {
			t.Errorf("Expected %s to be present in dockerConfig.CredentialHelpers: %v", helperSuffix, dockerConfig.CredentialHelpers)
		}
	}
}

func TestConfigureDocker_NonDefault(t *testing.T) {
	if err := deleteDockerConfig(); err != nil {
		if err != nil {
			t.Fatalf("Failed to delete the pre-existing docker config: %v", err)
		}
	}
	dockerConfig, err := getDockerConfig()
	if err != nil {
		t.Fatalf("Failed to get the docker config: %#v", err)
	}

	if len(dockerConfig.CredentialHelpers) > 0 {
		t.Fatal("Failed to clean the docker config.")
	}

	// Configure docker.
	helper := helperCmd([]string{"configure-docker", "--overwrite", "--registries=foo.gcr.io, bar.gcr.io, baz.gcr.io"})
	var out bytes.Buffer
	helper.Stdout = &out
	helper.Stderr = os.Stderr
	if err = helper.Run(); err != nil {
		t.Fatalf("Failed to execute `configure-docker --overwrite`: %v Stdout: %s", err, string(out.Bytes()))
	}

	dockerConfig, err = getDockerConfig()
	if err != nil {
		t.Fatalf("Failed to get the docker config: %v", err)
	}

	if len(dockerConfig.CredentialHelpers) == 0 {
		t.Fatal("CredentialHelpers was empty.")
	}

	for _, registryHostname := range []string{"foo.gcr.io", "bar.gcr.io", "baz.gcr.io"} {
		if helperSuffix, ok := dockerConfig.CredentialHelpers[registryHostname]; ok {
			if helperSuffix != "gcr" {
				t.Errorf("Wanted value for %s in dockerConfig.CredentialHelpers to be %s, got %s", registryHostname, "gcr", helperSuffix)
			}
		} else {
			t.Errorf("Expected %s to be present in dockerConfig.CredentialHelpers: %v", helperSuffix, dockerConfig.CredentialHelpers)
		}
	}
}
