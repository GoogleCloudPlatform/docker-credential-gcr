// +build unit

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

package util

import (
	"testing"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/mock/mock_cmd" // mocks must be generated before test execution
	"github.com/GoogleCloudPlatform/docker-credential-gcr/util/cmd"
	"github.com/golang/mock/gomock"
)

func mockGetVersionCmd(ctrl *gomock.Controller, returnVals ...interface{}) cmd.Command {
	mockCmd := mock_cmd.NewMockCommand(ctrl)
	mockCmd.EXPECT().Exec("version", "--format", "'{{.Client.Version}}'").Return(returnVals...)
	return mockCmd
}

func TestDockerClientVersion_Basic(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Mock out the executor.
	docker = mockGetVersionCmd(mockCtrl, []byte("1.2.3"), nil)

	major, minor, patch, patchSuffix, err := DockerClientVersion()

	if err != nil {
		t.Fatalf("DockerClientVersion returned an err: %v", err)
	}

	if major != 1 {
		t.Errorf("Expected major version number to be 1, got %d", major)
	}
	if minor != 2 {
		t.Errorf("Expected minor version number to be 2, got %d", minor)
	}
	if patch != 3 {
		t.Errorf("Expected patch version number to be 3, got %d", patch)
	}
	if patchSuffix != "" {
		t.Errorf("Expected patchSuffix to be blank, got \"%s\"", patchSuffix)
	}
}

func TestDockerClientVersion_WithSuffix(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Mock out the executor.
	docker = mockGetVersionCmd(mockCtrl, []byte("1.2.3-dev"), nil)

	major, minor, patch, patchSuffix, err := DockerClientVersion()

	if err != nil {
		t.Fatalf("DockerClientVersion returned an err: %v", err)
	}

	if major != 1 {
		t.Errorf("Expected major version number to be 1, got %d", major)
	}
	if minor != 2 {
		t.Errorf("Expected minor version number to be 2, got %d", minor)
	}
	if patch != 3 {
		t.Errorf("Expected patch version number to be 3, got %d", patch)
	}
	if patchSuffix != "dev" {
		t.Errorf("Expected patchSuffix to be \"dev\", got \"%s\"", patchSuffix)
	}
}

func TestDockerClientVersion_SingleQuotedWithSuffix(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Mock out the executor.
	docker = mockGetVersionCmd(mockCtrl, []byte("'1.2.3-dev'"), nil)

	major, minor, patch, patchSuffix, err := DockerClientVersion()

	if err != nil {
		t.Fatalf("DockerClientVersion returned an err: %v", err)
	}

	if major != 1 {
		t.Errorf("Expected major version number to be 1, got %d", major)
	}
	if minor != 2 {
		t.Errorf("Expected minor version number to be 2, got %d", minor)
	}
	if patch != 3 {
		t.Errorf("Expected patch version number to be 3, got %d", patch)
	}
	if patchSuffix != "dev" {
		t.Errorf("Expected patchSuffix to be \"dev\", got \"%s\"", patchSuffix)
	}
}

func TestDockerClientVersion_SingleQuotes(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Mock out the executor.
	docker = mockGetVersionCmd(mockCtrl, []byte("'1.2.3'"), nil)

	major, minor, patch, patchSuffix, err := DockerClientVersion()

	if err != nil {
		t.Fatalf("DockerClientVersion returned an err: %v", err)
	}

	if major != 1 {
		t.Errorf("Expected major version number to be 1, got %d", major)
	}
	if minor != 2 {
		t.Errorf("Expected minor version number to be 2, got %d", minor)
	}
	if patch != 3 {
		t.Errorf("Expected patch version number to be 3, got %d", patch)
	}
	if patchSuffix != "" {
		t.Errorf("Expected patchSuffix to be blank, got \"%s\"", patchSuffix)
	}
}

func TestDockerClientVersion_Whitespace(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Mock out the executor.
	docker = mockGetVersionCmd(mockCtrl, []byte("1.2.3\n"), nil)

	major, minor, patch, patchSuffix, err := DockerClientVersion()

	if err != nil {
		t.Fatalf("DockerClientVersion returned an err: %v", err)
	}

	if major != 1 {
		t.Errorf("Expected major version number to be 1, got %d", major)
	}
	if minor != 2 {
		t.Errorf("Expected minor version number to be 2, got %d", minor)
	}
	if patch != 3 {
		t.Errorf("Expected patch version number to be 3, got %d", patch)
	}
	if patchSuffix != "" {
		t.Errorf("Expected patchSuffix to be blank, got \"%s\"", patchSuffix)
	}
}

func TestDockerClientVersion_WhitespaceQuoted(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Mock out the executor.
	docker = mockGetVersionCmd(mockCtrl, []byte("'1.2.3'\n"), nil)

	major, minor, patch, patchSuffix, err := DockerClientVersion()

	if err != nil {
		t.Fatalf("DockerClientVersion returned an err: %v", err)
	}

	if major != 1 {
		t.Errorf("Expected major version number to be 1, got %d", major)
	}
	if minor != 2 {
		t.Errorf("Expected minor version number to be 2, got %d", minor)
	}
	if patch != 3 {
		t.Errorf("Expected patch version number to be 3, got %d", patch)
	}
	if patchSuffix != "" {
		t.Errorf("Expected patchSuffix to be blank, got \"%s\"", patchSuffix)
	}
}
