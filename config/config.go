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

// Package config provides constants used in configuring the behavior of the app.
package config

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	// GCRCredHelperClientID is the client_id to be used when performing the
	// OAuth2 Authorization Code grant flow.
	// See https://developers.google.com/identity/protocols/OAuth2InstalledApp
	GCRCredHelperClientID = "99426463878-o7n0bshgue20tdpm25q4at0vs2mr4utq.apps.googleusercontent.com"

	// GCRCredHelperClientNotSoSecret is the client_secret to be used when
	// performing the OAuth2 Authorization Code grant flow.
	// See https://developers.google.com/identity/protocols/OAuth2InstalledApp
	GCRCredHelperClientNotSoSecret = "HpVi8cnKx8AAkddzaNrSWmS8"

	// From http://semver.org/
	// MAJOR version when you make incompatible API changes,
	// MINOR version when you add functionality in a backwards-compatible manner, and
	// PATCH version when you make backwards-compatible bug fixes.

	// MajorVersion is the credential helper's major version number.
	MajorVersion = 1
	// MinorVersion is the credential helper's minor version number.
	MinorVersion = 0
	// PatchVersion is the credential helper's patch version number.
	PatchVersion = 0
)

// SupportedGCRRegistries maps registry URLs to a bool representing whether
// or not the GCR credentials can be used to authenticate requests for that
// repository.
var SupportedGCRRegistries = map[string]bool{
	"gcr.io":            true,
	"us.gcr.io":         true,
	"eu.gcr.io":         true,
	"asia.gcr.io":       true,
	"b.gcr.io":          true,
	"bucket.gcr.io":     true,
	"appengine.gcr.io":  true,
	"gcr.kubernetes.io": true,
}

// GCROAuth2Endpoint describes the oauth2.Endpoint to be used when
// authenticating a GCR user.
var GCROAuth2Endpoint = google.Endpoint

// GCRScopes is/are the OAuth2 scope(s) to request during access_token creation.
var GCRScopes = []string{"https://www.googleapis.com/auth/cloud-platform"}

// OAuthHTTPContext is the HTTP context to use when performing OAuth2 calls.
var OAuthHTTPContext = oauth2.NoContext
