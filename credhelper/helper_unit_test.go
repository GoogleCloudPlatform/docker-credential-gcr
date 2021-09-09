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

package credhelper

import (
	"errors"
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/mock/mock_cmd"
	"github.com/GoogleCloudPlatform/docker-credential-gcr/mock/mock_config" // mocks must be generated before test execution
	"github.com/GoogleCloudPlatform/docker-credential-gcr/mock/mock_store"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	"github.com/GoogleCloudPlatform/docker-credential-gcr/store"
	"github.com/GoogleCloudPlatform/docker-credential-gcr/util/cmd"
	"github.com/golang/mock/gomock"
)

var expectedGCRUsername = fmt.Sprintf("_dcgcr_%s_token", config.Version)

var testGCRHosts = [...]string{
	"gcr.io",
	"us.gcr.io",
	"eu.gcr.io",
	"asia.gcr.io",
	"staging-k8s.gcr.io",
	"marketplace.gcr.io",
	"appengine.gcr.io",
	"hypothetical-alias.gcr.io",
}

func TestGet_GCRCredentials(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// create a mocks for the helper to use
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)
	mockUserCfg := mock_config.NewMockUserConfig(mockCtrl)

	// mock the helper methods used by getGCRAccessToken
	expectedSecret := "secrets!"
	tested := &gcrCredHelper{
		store:   mockStore,
		userCfg: mockUserCfg,
		envToken: func() (string, error) {
			return expectedSecret, nil
		},
		gcloudSDKToken: func(_ cmd.Command) (string, error) {
			return "", errors.New("no token here")
		},
		credStoreToken: func(_ store.GCRCredStore) (string, error) {
			return "", errors.New("no token here")
		},
	}

	// Verify that all of GCR's hostnames return GCR's access token.
	for _, host := range testGCRHosts {
		mockUserCfg.EXPECT().TokenSources().Return(config.DefaultTokenSources[:])
		username, secret, err := tested.Get("https://" + host)
		if err != nil {
			t.Errorf("get returned an error: %v", err)
		} else if username != expectedGCRUsername {
			t.Errorf("expected GCR username: %s but got: %s", expectedGCRUsername, username)
		} else if secret != expectedSecret {
			t.Errorf("expected secret: %s but got: %s", expectedSecret, secret)
		}
	}
}

/*
	The following tests verify the behavior of getGCRAccessToken. Preference
	is defined by tokenSources
*/

func TestGetGCRAccessToken_Env(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	// create a mock store to use
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)

	mockUserCfg := mock_config.NewMockUserConfig(mockCtrl)
	mockUserCfg.EXPECT().TokenSources().Return(config.DefaultTokenSources[:])

	// mock the helper methods used by getGCRAccessToken
	const expected = "application default creds!"
	tested := &gcrCredHelper{
		store:   mockStore,
		userCfg: mockUserCfg,
		envToken: func() (string, error) {
			return expected, nil
		},
		gcloudSDKToken: func(_ cmd.Command) (string, error) {
			return "", errors.New("no token from gcloud")
		},
		credStoreToken: func(_ store.GCRCredStore) (string, error) {
			return "", errors.New("no token in the cred store")
		},
	}

	token, err := tested.getGCRAccessToken()

	if err != nil {
		t.Fatalf("getGCRAccessToken returned an error: %v", err)
	} else if token != expected {
		t.Fatalf("Expected: %s got: %s", expected, token)
	}
}

func TestGetGCRAccessToken_PrivateStore(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// create a mock store to use
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)

	mockUserCfg := mock_config.NewMockUserConfig(mockCtrl)
	mockUserCfg.EXPECT().TokenSources().Return(config.DefaultTokenSources[:])

	// mock the helper methods used by getGCRAccessToken
	const expected = "private creds!"
	tested := &gcrCredHelper{
		store:   mockStore,
		userCfg: mockUserCfg,
		envToken: func() (string, error) {
			return "creds from `env`", nil
		},
		gcloudSDKToken: func(_ cmd.Command) (string, error) {
			return "creds from `gcloud`", nil

		},
		credStoreToken: func(_ store.GCRCredStore) (string, error) {
			return expected, nil
		},
	}

	token, err := tested.getGCRAccessToken()

	if err != nil {
		t.Fatalf("getGCRAccessToken returned an error: %v", err)
	} else if token != expected {
		t.Fatalf("Expected: %s got: %s", expected, token)
	}
}

func TestGetGCRAccessToken_NoneExist(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// create a mock store to use
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)

	mockUserCfg := mock_config.NewMockUserConfig(mockCtrl)
	mockUserCfg.EXPECT().TokenSources().Return(config.DefaultTokenSources[:])

	// mock the helper methods used by getGCRAccessToken
	tested := &gcrCredHelper{
		store:   mockStore,
		userCfg: mockUserCfg,
		envToken: func() (string, error) {
			return "", errors.New("no token here")
		},
		gcloudSDKToken: func(_ cmd.Command) (string, error) {
			return "", errors.New("still no token here")
		},
		credStoreToken: func(_ store.GCRCredStore) (string, error) {
			return "", errors.New("sad panda")
		},
	}

	token, err := tested.getGCRAccessToken()

	if err == nil {
		t.Fatalf("Expected an error, got token: %s", token)
	}
}

func TestGetGCRAccessToken_CustomTokenSources(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// create a mock store to use
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)

	// Mock a user config, re-arranging the token sources.
	mockUserCfg := mock_config.NewMockUserConfig(mockCtrl)
	mockUserCfg.EXPECT().TokenSources().Return([]string{"store", "gcloud", "env"}) // reversed from default

	const (
		gcloudCreds = "gcloud sdk creds!"
		storeCreds  = "private creds!"
		envCreds    = "environment creds!"
	)
	// mock the helper methods used by getGCRAccessToken
	tested := &gcrCredHelper{
		store:   mockStore,
		userCfg: mockUserCfg,
		envToken: func() (string, error) {
			return envCreds, nil
		},
		gcloudSDKToken: func(_ cmd.Command) (string, error) {
			return gcloudCreds, nil
		},
		credStoreToken: func(_ store.GCRCredStore) (string, error) {
			return storeCreds, nil
		},
	}

	token, err := tested.getGCRAccessToken()

	if err != nil {
		t.Fatalf("getGCRAccessToken returned an error: %v", err)
	} else if token != storeCreds {
		t.Fatalf("Expected: %s got: %s", storeCreds, token)
	}
}

func TestGetGCRAccessToken_CustomTokenSources_ValidSourcesDisabled(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// create a mock store to use
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)

	// Mock a user config, disabling some token sources.
	mockUserCfg := mock_config.NewMockUserConfig(mockCtrl)
	mockUserCfg.EXPECT().TokenSources().Return([]string{"gcloud"}) // gcloud only configured source

	const (
		storeCreds = "private creds!"
		envCreds   = "environment creds!"
	)
	// mock the helper methods used by getGCRAccessToken
	tested := &gcrCredHelper{
		store:   mockStore,
		userCfg: mockUserCfg,
		envToken: func() (string, error) {
			return envCreds, nil
		},
		gcloudSDKToken: func(_ cmd.Command) (string, error) {
			return "", errors.New("no token here")
		},
		credStoreToken: func(_ store.GCRCredStore) (string, error) {
			return storeCreds, nil
		},
	}

	token, err := tested.getGCRAccessToken()

	if err == nil {
		t.Fatalf("Expected an error, got token: %s", token)
	}
}

func TestGetGCRAccessToken_OldGcloudSdkTokenSourceString(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// create a mock store to use
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)

	// Mock a user config, disabling some token sources.
	mockUserCfg := mock_config.NewMockUserConfig(mockCtrl)
	mockUserCfg.EXPECT().TokenSources().Return([]string{"gcloud_sdk"}) // the string that was initially used for specify gcloud

	const (
		envCreds    = "environment creds!"
		gcloudCreds = "gcloud sdk creds!"
		storeCreds  = "private creds!"
	)
	// mock the helper methods used by getGCRAccessToken
	tested := &gcrCredHelper{
		store:   mockStore,
		userCfg: mockUserCfg,
		envToken: func() (string, error) {
			return envCreds, nil
		},
		gcloudSDKToken: func(_ cmd.Command) (string, error) {
			return gcloudCreds, nil
		},
		credStoreToken: func(_ store.GCRCredStore) (string, error) {
			return storeCreds, nil
		},
	}

	token, err := tested.getGCRAccessToken()

	if err != nil {
		t.Fatalf("tokenFromGcloudSDK returned an error: %v", err)
	} else if token != gcloudCreds {
		t.Fatalf("Expected: '%s' got: '%s'", gcloudCreds, token)
	}
}

func TestGetGCRAccessToken_CustomTokenSources_InvalidSource(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// create a mock store to use
	mockStore := mock_store.NewMockGCRCredStore(mockCtrl)

	mockUserCfg := mock_config.NewMockUserConfig(mockCtrl)
	mockUserCfg.EXPECT().TokenSources().Return([]string{"invalid"})

	const (
		gcloudCreds = "gcloud sdk creds!"
		storeCreds  = "private creds!"
		envCreds    = "environment creds!"
	)
	// mock the helper methods used by getGCRAccessToken
	tested := &gcrCredHelper{
		store:   mockStore,
		userCfg: mockUserCfg,
		envToken: func() (string, error) {
			return envCreds, nil
		},
		gcloudSDKToken: func(_ cmd.Command) (string, error) {
			return gcloudCreds, nil
		},
		credStoreToken: func(_ store.GCRCredStore) (string, error) {
			return storeCreds, nil
		},
	}

	token, err := tested.getGCRAccessToken()

	if err == nil {
		t.Fatalf("Expected an error, got token: %s", token)
	}
}

func TestTokenFromGcloudSDK(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	const gcloudCreds = "gcloud sdk creds!"

	// This test is more-or-less tautological, but it's important to verify
	// that gcloud is being queried in a supported way.
	mockCmd := mock_cmd.NewMockCommand(mockCtrl)
	mockCmd.EXPECT().Exec("config", "config-helper", "--force-auth-refresh", "--format=value(credential.access_token)").Return([]uint8(gcloudCreds), nil)

	token, err := tokenFromGcloudSDK(mockCmd)

	if err != nil {
		t.Fatalf("tokenFromGcloudSDK returned an error: %v", err)
	} else if token != gcloudCreds {
		t.Fatalf("Expected: '%s' got: '%s'", gcloudCreds, token)
	}
}
