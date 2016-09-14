// +build unit

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

package api

import (
	"testing"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	"github.com/docker/docker/api/types"
)

func TestAddHostNamesToAuths_AddOne(t *testing.T) {
	hostName := "http://expected.com"
	auths := make(map[string]types.AuthConfig)

	authsModified := addHostNamesToAuths([]string{hostName}, auths)

	if !authsModified {
		t.Error("Expected addHostNamesToAuths to return true")
	}
	if len(auths) != 1 {
		t.Error("Expected auths to contain 1 element, was %d", len(auths))
	}
	if _, exists := auths[hostName]; !exists {
		t.Errorf("Expected auths to contain an entry for %s", hostName)
	}

	if t.Failed() {
		t.Logf("Auths: %v", auths)
	}
}

func TestAddHostNamesToAuths_AddOne_AlreadyExists(t *testing.T) {
	hostName := "http://expected.com"
	auths := make(map[string]types.AuthConfig)
	auths[hostName] = types.AuthConfig{}

	authsModified := addHostNamesToAuths([]string{hostName}, auths)

	if authsModified {
		t.Error("Expected addHostNamesToAuths to return false")
	}
	if len(auths) != 1 {
		t.Errorf("Expected auths to contain 1 element, was %d", len(auths))
	}
	if _, exists := auths[hostName]; !exists {
		t.Errorf("Expected auths to contain an entry for %s", hostName)
	}

	if t.Failed() {
		t.Logf("Auths: %v", auths)
	}
}

func TestAddHostNamesToAuths_AddOne_NoScheme(t *testing.T) {
	hostName := "expected.com"
	expectedEntry := "https://" + hostName
	auths := make(map[string]types.AuthConfig)

	authsModified := addHostNamesToAuths([]string{hostName}, auths)

	if !authsModified {
		t.Error("Expected addHostNamesToAuths to return true")
	}
	if len(auths) != 1 {
		t.Error("Expected auths to contain 1 element, was %d", len(auths))
	}
	if _, exists := auths[expectedEntry]; !exists {
		t.Errorf("Expected auths to contain an entry for %s", expectedEntry)
	}

	if t.Failed() {
		t.Logf("Auths: %v", auths)
	}
}

func TestAddHostNamesToAuths_AddAllGCR(t *testing.T) {
	auths := make(map[string]types.AuthConfig)
	var newRegistries []string

	for registry := range config.SupportedGCRRegistries {
		newRegistries = append(newRegistries, registry)
	}

	authsModified := addHostNamesToAuths(newRegistries, auths)

	if !authsModified {
		t.Error("Expected addHostNamesToAuths to return true")
	}
	numRegistries := len(config.SupportedGCRRegistries)
	numAuths := len(auths)
	if numAuths != numRegistries {
		t.Errorf("Expected auths to contain %d elements, was %d", numRegistries, numAuths)
	}
	for registry := range config.SupportedGCRRegistries {
		expectedEntry := "https://" + registry
		if _, exists := auths[expectedEntry]; !exists {
			t.Errorf("Expected auths to contain an entry for %s", expectedEntry)
		}
	}

	if t.Failed() {
		t.Logf("Auths: %v", auths)
	}
}
