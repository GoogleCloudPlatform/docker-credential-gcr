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
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initTestEnvironment() error {
	err := removeTestFiles()
	if err != nil {
		return err
	}
	return setTestFileEnvVars()
}

// Ensure that there are no pre-existing test files.
func removeTestFiles() error {
	path, err := testConfigPath()
	if err != nil {
		return err
	}
	if _, err = os.Stat(path); err == nil {
		if err = os.Remove(path); err != nil {
			return err
		}
	}

	path, err = testCredStorePath()
	if err != nil {
		return err
	}
	if _, err = os.Stat(path); err == nil {
		return os.Remove(path)
	}
	return nil
}

func setTestFileEnvVars() error {
	// construct test file paths
	configPath, err := testConfigPath()
	if err != nil {
		return err
	}
	storePath, err := testCredStorePath()
	if err != nil {
		return err
	}

	// set the test files in the environment variables
	os.Setenv(credentialStoreEnvVar, storePath)
	os.Setenv(configFileEnvVar, configPath)

	return nil
}

// helperCmd returns a *exec.Cmd to execute the credential helper binary with
// the given arguments.
func helperCmd(args []string) *exec.Cmd {
	// Execute the compiled helper binary relative to the test directory.
	return exec.Command("../docker-credential-gcr", args...)
}

// testConfigPath returns the absolute path to the test credential helper config.
func testConfigPath() (string, error) {
	return filepath.Abs(testConfigFile)
}

// testCredStorePath returns the absolute path to the test credential store.
func testCredStorePath() (string, error) {
	return filepath.Abs(testStoreFile)
}

// assertTestEnv guarantees that the config file and credential store
// environment variables have been set, point to the expected paths,
// and that those files do not exist yet. Otherwise, calls t.FailNow.
func assertTestEnv(t *testing.T) {
	expectedTestConfigPath, err := testConfigPath()
	if err != nil {
		t.Fatalf("Unable construct test config path: %v", err)
	}
	if _, err = os.Stat(expectedTestConfigPath); err == nil {
		t.Errorf("Test config exists: %s", expectedTestConfigPath)
	}

	expectedTestCredStorePath, err := testCredStorePath()
	if err != nil {
		t.Fatalf("Unable construct test cred store path: %v", err)
	}
	if _, err = os.Stat(expectedTestCredStorePath); err == nil {
		t.Errorf("Test cred store exists: %s", expectedTestCredStorePath)
	}

	actualStorePath := os.Getenv(credentialStoreEnvVar)
	if actualStorePath != expectedTestCredStorePath {
		t.Errorf("Expected credential store env var %s to be set to: %s, got %s", credentialStoreEnvVar, expectedTestCredStorePath, actualStorePath)
	}
	actualConfigPath := os.Getenv(configFileEnvVar)
	if actualConfigPath != expectedTestConfigPath {
		t.Errorf("Expected config file env var %s to be set to: %s, got %s", configFileEnvVar, expectedTestConfigPath, actualConfigPath)
	}

	if t.Failed() {
		t.FailNow()
	}
}
