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

package store

import (
	"os"
	"strings"
	"testing"
)

const (
	expectedStoreEnvVar = "DOCKER_CREDENTIAL_GCR_STORE"
	expectedEnvPath     = "whatever"
	expectedFilename    = "docker_credentials.json"
)

func TestDockerCredentialPath_RespectsEnvVar(t *testing.T) {
	old := os.Getenv(expectedStoreEnvVar)
	os.Setenv(expectedStoreEnvVar, expectedEnvPath)
	defer os.Setenv(expectedStoreEnvVar, old)

	result, err := dockerCredentialPath()

	if err != nil {
		t.Fatalf("Could not retrieve cred store path: %v", err)
	} else if result != expectedEnvPath {
		t.Fatalf("Expected store path to be: %s, was: %s", expectedStoreEnvVar, result)
	}
}

func TestDockerCredentialPath_Sanity(t *testing.T) {
	result, err := dockerCredentialPath()

	if err != nil {
		t.Fatalf("Could not retrieve cred store path: %v", err)
	} else if !strings.HasSuffix(result, expectedFilename) {
		t.Fatalf("Expected store path to end with: %s, was: %s", expectedFilename, result)
	}
}
