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

package cli

import (
	"errors"
	"reflect"
	"testing"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/mock/mock_store"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/cli/config/configfile"
	"github.com/golang/mock/gomock"
)

func expectedAuthConfigs() map[string]types.AuthConfig {
	return map[string]types.AuthConfig{
		"https://gcr.io":            {},
		"https://us.gcr.io":         {},
		"https://eu.gcr.io":         {},
		"https://asia.gcr.io":       {},
		"https://b.gcr.io":          {},
		"https://bucket.gcr.io":     {},
		"https://appengine.gcr.io":  {},
		"https://gcr.kubernetes.io": {},
		"https://beta.gcr.io":       {},
	}
}

func TestSetGCRAuthConfigs_EmptyConfig(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)
	mockStore.EXPECT().AllThirdPartyCreds().Return(nil, nil)

	tested := &configfile.ConfigFile{}

	modified := setAuthConfigs(tested, mockStore)

	if !modified {
		t.Fatal("expected setAuthConfigs to return true")
	}

	if !reflect.DeepEqual(expectedAuthConfigs(), tested.AuthConfigs) {
		t.Fatalf("expected: %v, got: %v", expectedAuthConfigs(), tested.AuthConfigs)
	}
}

func TestSetGCRAuthConfigs_CredStoreSet(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)
	mockStore.EXPECT().AllThirdPartyCreds().Return(nil, nil)

	tested := &configfile.ConfigFile{}
	tested.CredentialsStore = "gcr"

	modified := setAuthConfigs(tested, mockStore)

	if !modified {
		t.Fatal("expected setAuthConfigs to return true")
	}

	if !reflect.DeepEqual(expectedAuthConfigs(), tested.AuthConfigs) {
		t.Errorf("expected: %v, got: %v", expectedAuthConfigs(), tested.AuthConfigs)
	}

	if tested.CredentialsStore != "gcr" {
		t.Errorf("expected credsStore to be: %s, got: %s", "gcr", tested.CredentialsStore)
	}
}

func TestSetGCRAuthConfigs_NoOp(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)
	mockStore.EXPECT().AllThirdPartyCreds().Return(nil, nil)

	tested := &configfile.ConfigFile{}
	tested.AuthConfigs = expectedAuthConfigs()

	modified := setAuthConfigs(tested, mockStore)

	if modified {
		t.Fatal("expected setAuthConfigs to return false")
	}

	if !reflect.DeepEqual(expectedAuthConfigs(), tested.AuthConfigs) {
		t.Errorf("expected: %v, got: %v", expectedAuthConfigs(), tested.AuthConfigs)
	}
}

func TestSetGCRAuthConfigs_OverwriteThirdParty(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)
	mockStore.EXPECT().AllThirdPartyCreds().Return(nil, nil)

	tested := &configfile.ConfigFile{}
	tested.AuthConfigs = map[string]types.AuthConfig{
		"https://www.whatever.com": {},
	}

	modified := setAuthConfigs(tested, mockStore)

	if !modified {
		t.Fatal("expected setAuthConfigs to return true")
	}

	if !reflect.DeepEqual(expectedAuthConfigs(), tested.AuthConfigs) {
		t.Errorf("expected: %v, got: %v", expectedAuthConfigs(), tested.AuthConfigs)
	}
}

func TestSetGCRAuthConfigs_OverwriteMixed(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)
	mockStore.EXPECT().AllThirdPartyCreds().Return(nil, nil)

	tested := &configfile.ConfigFile{}
	tested.AuthConfigs = expectedAuthConfigs()
	tested.AuthConfigs["https://www.whatever.com"] = types.AuthConfig{}

	modified := setAuthConfigs(tested, mockStore)

	if !modified {
		t.Fatal("expected setAuthConfigs to return true")
	}

	if !reflect.DeepEqual(expectedAuthConfigs(), tested.AuthConfigs) {
		t.Errorf("expected: %v, got: %v", expectedAuthConfigs(), tested.AuthConfigs)
	}
}

func TestSetGCRAuthConfigs_Subset(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)
	mockStore.EXPECT().AllThirdPartyCreds().Return(nil, nil)

	tested := &configfile.ConfigFile{}
	tested.AuthConfigs = expectedAuthConfigs()
	delete(tested.AuthConfigs, "https://gcr.io")
	delete(tested.AuthConfigs, "https://beta.gcr.io")

	modified := setAuthConfigs(tested, mockStore)

	if !modified {
		t.Fatal("expected setAuthConfigs to return true")
	}

	if !reflect.DeepEqual(expectedAuthConfigs(), tested.AuthConfigs) {
		t.Errorf("expected: %v, got: %v", expectedAuthConfigs(), tested.AuthConfigs)
	}
}

func TestSetGCRAuthConfigs_Adds3p(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	expectedThirdParty := "https://www.winning.com"
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)
	mockStore.EXPECT().AllThirdPartyCreds().Return(map[string]credentials.Credentials{
		expectedThirdParty: {},
	}, nil)

	tested := &configfile.ConfigFile{}
	tested.AuthConfigs = nil

	modified := setAuthConfigs(tested, mockStore)

	if !modified {
		t.Fatal("expected setAuthConfigs to return true")
	}

	expected := expectedAuthConfigs()
	expected[expectedThirdParty] = types.AuthConfig{}

	if !reflect.DeepEqual(expected, tested.AuthConfigs) {
		t.Errorf("expected: %v, got: %v", expected, tested.AuthConfigs)
	}
}

func TestSetGCRAuthConfigs_InvalidCredStoreOK(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)
	mockStore.EXPECT().AllThirdPartyCreds().Return(nil, errors.New("dead cred store"))

	tested := &configfile.ConfigFile{}

	modified := setAuthConfigs(tested, mockStore)

	if !modified {
		t.Fatal("expected setAuthConfigs to return true")
	}

	if !reflect.DeepEqual(expectedAuthConfigs(), tested.AuthConfigs) {
		t.Errorf("expected: %v, got: %v", expectedAuthConfigs(), tested.AuthConfigs)
	}
}

func TestCredHelpersSupported_Supported(t *testing.T) {
	t.Parallel()

	for minor := 13; minor < 16; minor++ {
		if !credHelpersSupported(1, minor) {
			t.Errorf("credHelperSupported erronously returned false for v1.%d", minor)
		}
	}
}

func TestCredHelpersSupported_Unsupported(t *testing.T) {
	t.Parallel()

	for major := -1; major < 1; major++ {
		if credHelpersSupported(major, 0) {
			t.Errorf("credHelperSupported erronously returned true for v%d.0", major)
		}
	}

	for minor := -1; minor < 13; minor++ {
		if credHelpersSupported(1, minor) {
			t.Errorf("credHelperSupported erronously returned true for v1.%d", minor)
		}
	}
}

func TestCredsStoreSupported_Supported(t *testing.T) {
	t.Parallel()

	for minor := 11; minor < 16; minor++ {
		if !credsStoreSupported(1, minor) {
			t.Errorf("credsStoreSupported erronously returned false for v1.%d", minor)
		}
	}
}

func TestCredsStoreSupported_Unsupported(t *testing.T) {
	t.Parallel()

	for major := -1; major < 1; major++ {
		if credsStoreSupported(major, 0) {
			t.Errorf("credsStoreSupported erronously returned true for v%d.0", major)
		}
	}

	for minor := -1; minor < 11; minor++ {
		if credsStoreSupported(1, minor) {
			t.Errorf("credsStoreSupported erronously returned true for v1.%d", minor)
		}
	}
}
