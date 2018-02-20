// +build travis

// Copyright 2018 Google LLC
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
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
)

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

	for registryHostname := range config.DefaultGCRRegistries {
		if helperSuffix, ok := dockerConfig.CredentialHelpers[registryHostname]; ok {
			if helperSuffix != "gcr" {
				t.Errorf("Wanted value for %s in dockerConfig.CredentialHelpers to be %s, got %s", registryHostname, "gcr", helperSuffix)
			}
		} else {
			t.Errorf("Expected %s to be present in dockerConfig.CredentialHelpers: %v", helperSuffix, dockerConfig.CredentialHelpers)
		}
	}
}
