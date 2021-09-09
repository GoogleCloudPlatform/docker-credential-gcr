//go:build !unit && !gazelle
// +build !unit,!gazelle

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
	"regexp"
	"testing"
)

func TestVersion(t *testing.T) {
	helper := helperCmd([]string{"version"})

	var out bytes.Buffer
	helper.Stdout = &out
	if err := helper.Run(); err != nil {
		t.Fatalf("Failed to exec the helper: %v", err)
	}

	// Enforce a particular format so that a regex can extract the version easily.
	actual := out.String()
	expectedProdRegex := "Google Container Registry Docker credential helper [0-9]+\\_[0-9]+\\_[0-9]+\n"
	if match, _ := regexp.MatchString(expectedProdRegex, actual); !match {
		// Fail if not a dev version.
		expectedDevString := "Google Container Registry Docker credential helper (devel)\n"
		if actual != expectedDevString {
			t.Fatalf("Expected version string to match: %s, got: %s", expectedProdRegex, actual)
		}
	}
}
