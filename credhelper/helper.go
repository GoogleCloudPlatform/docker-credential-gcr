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

/*
Package credhelper implements a Docker credential helper with special facilities
for GCR authentication.
*/
package credhelper

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/google/docker-credential-gcr/config"
	"github.com/google/docker-credential-gcr/store"

	"golang.org/x/oauth2/google"
)

const gcrOAuth2Username = "oauth2accesstoken"

// gcrCredHelper implements a credentials.Helper interface backed by a GCR
// credential store.
type gcrCredHelper struct {
	store store.GCRCredStore

	// helper methods, package exposed for testing
	defaultCredsToken func() (string, error)
	gcloudSDKToken    func() (string, error)
	credStoreToken    func(store.GCRCredStore) (string, error)
}

// NewGCRCredentialHelper returns a Docker credential helper which
// specializes in GCR's authentication schemes.
func NewGCRCredentialHelper(store store.GCRCredStore) credentials.Helper {
	return &gcrCredHelper{
		store:             store,
		credStoreToken:    tokenFromPrivateStore,
		gcloudSDKToken:    tokenFromGcloudSDK,
		defaultCredsToken: tokenFromAppDefaultCreds,
	}
}

// Delete lists all credentials stored and associated usernames.
func (ch *gcrCredHelper) List() ([]string, []string, error) {
	all3pCreds, err := ch.store.AllThirdPartyCreds()
	if err != nil {
		return nil, nil, helperErr("could not retrieve 3p credentials", err)
	}

	numRegistries := len(all3pCreds) + len(config.SupportedGCRRegistries)
	var registries, usernames []string
	registries = make([]string, 0, numRegistries)
	usernames = make([]string, 0, numRegistries)

	for registry, creds := range all3pCreds {
		registries = append(registries, registry)
		usernames = append(usernames, creds.Username)
	}

	for gcrRegistry := range config.SupportedGCRRegistries {
		registries = append(registries, gcrRegistry)
		usernames = append(usernames, gcrOAuth2Username)
	}

	return registries, usernames, nil
}

// Add adds new third-party credentials to the keychain.
func (ch *gcrCredHelper) Add(creds *credentials.Credentials) error {
	serverURL := creds.ServerURL
	if isAGCRHostname(serverURL) {
		return helperErr("add is unimplemented for GCR, please use one of the supported authentication methods", nil)
	}
	if err := ch.store.SetOtherCreds(creds); err != nil {
		return helperErr("could not store 3p credentials for "+serverURL, err)
	}
	return nil
}

// Delete removes third-party credentials from the store.
func (ch *gcrCredHelper) Delete(serverURL string) error {
	if isAGCRHostname(serverURL) {
		return helperErr("delete is unimplemented for GCR: "+serverURL, nil)
	}
	if err := ch.store.DeleteOtherCreds(serverURL); err != nil {
		return helperErr("could not delete 3p credentials for "+serverURL, err)
	}
	return nil
}

// Get returns the username and secret to use for a given registry server URL.
func (ch *gcrCredHelper) Get(serverURL string) (string, string, error) {
	if isAGCRHostname(serverURL) {
		// Return GCR's access token.
		accessToken, err := ch.getGCRAccessToken()
		if err != nil {
			return "", "", helperErr(fmt.Sprintf("could not retrieve %s's access token", serverURL), err)
		}
		return gcrOAuth2Username, accessToken, nil
	}

	// Attempt to retrieve credentials for another repository
	creds, err := ch.store.GetOtherCreds(serverURL)
	if err != nil {
		return "", "", helperErr("could not retrieve 3p credentials for "+serverURL, err)
	}
	return creds.Username, creds.Secret, nil
}

// getGCRAccessToken attempts to generate an access_token from credentials
// from the helper's own credential store.
func (ch *gcrCredHelper) getGCRAccessToken() (string, error) {
	// First, look for the Application Default Credentials.
	// https://developers.google.com/identity/protocols/application-default-credentials
	token, err := ch.defaultCredsToken()
	if err != nil {
		// Second, attempt to retrieve credentials from the gcloud SDK's store.
		token, err = ch.gcloudSDKToken()
	}
	if err != nil {
		// Finally, retrieve credentials from our credential store.
		token, err = ch.credStoreToken(ch.store)
		if err != nil {
			return "", err
		}
	}

	return token, nil
}

// tokenFromAppDefaultCreds retrieves a gcloud
// access_token from the "Application Default Credentials".
func tokenFromAppDefaultCreds() (string, error) {
	/*
		From https://godoc.org/golang.org/x/oauth2/google:

		DefaultTokenSource is a token source that uses "Application Default Credentials".

		It looks for credentials in the following places, preferring the first location found:

		1. A JSON file whose path is specified by the
		   GOOGLE_APPLICATION_CREDENTIALS environment variable.
		2. A JSON file in a location known to the gcloud command-line tool.
		   On Windows, this is %APPDATA%/gcloud/application_default_credentials.json.
		   On other systems, $HOME/.config/gcloud/application_default_credentials.json.
		3. On Google App Engine it uses the appengine.AccessToken function.
		4. On Google Compute Engine and Google App Engine Managed VMs, it fetches
		   credentials from the metadata server.
		   (In this final case any provided scopes are ignored.)
	*/
	ts, err := google.DefaultTokenSource(config.OAuthHTTPContext, config.GCRScopes...)
	if err != nil {
		return "", err
	}

	token, err := ts.Token()
	if err != nil {
		return "", err
	}

	if !token.Valid() {
		return "", helperErr("token was invalid", nil)
	}

	if token.Type() != "Bearer" {
		return "", helperErr(fmt.Sprintf("expected token type \"Bearer\" but got \"%s\"", token.Type()), nil)
	}

	return token.AccessToken, nil
}

// tokenFromGcloudSDK attempts to generate an access_token using the gcloud SDK.
func tokenFromGcloudSDK() (string, error) {
	// shelling out to gcloud is the only currently supported way of
	// obtaining the gcloud access_token
	if _, err := exec.LookPath("gcloud"); err != nil {
		return "", helperErr("gcloud not found on PATH", nil)
	}

	cmd := exec.Command("gcloud", "auth", "print-access-token")

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", helperErr("gcloud auth print-access-token failed", err)
	}

	token := strings.TrimSpace(out.String())
	if token == "" {
		return "", helperErr("gcloud auth print-access-token returned empty access_token", nil)
	}
	return token, nil
}

func tokenFromPrivateStore(store store.GCRCredStore) (string, error) {
	gcrAuth, err := store.GetGCRAuth()
	if err != nil {
		return "", err
	}
	ts := gcrAuth.TokenSource(config.OAuthHTTPContext)
	tok, err := ts.Token()
	if err != nil {
		return "", err
	}
	if !tok.Valid() {
		return "", helperErr("token was invalid", nil)
	}

	return tok.AccessToken, nil
}

// isAGCRHostname returns true if the given registry server URL is one of GCR's
func isAGCRHostname(serverURL string) bool {
	return config.SupportedGCRRegistries[serverURL]
}

func helperErr(message string, err error) error {
	if err == nil {
		return fmt.Errorf("docker-credential-gcr/helper: %s", message)
	}
	return fmt.Errorf("docker-credential-gcr/helper: %s: %v", message, err)
}
