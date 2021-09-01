//go:build !windows
// +build !windows

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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

const testAccessToken = "much access so token wow ^•ﻌ•^"

var testCredStorePath = filepath.Clean(credentialStoreFilename) // ./<credentialStoreFilename>

func TestMain(m *testing.M) {
	err := cleanUp()
	if err != nil {
		panic(fmt.Sprintf("Unable to remove previously existing test data store, test results cannot be trusted: %v", err))
	}
	exitCode := m.Run()

	// be polite and clean up
	cleanUp()
	os.Exit(exitCode)
}

func writeCredentialsToStoreFile(t *testing.T, creds *dockerCredentials) {
	f, err := os.Create(testCredStorePath)
	if err != nil {
		t.Fatal("Could not create credential store file.")
	}
	err = json.NewEncoder(f).Encode(*creds)
	if err != nil {
		t.Fatal("Could not write credentials to store.")
	}
}

func cleanUp() error {
	err := os.Remove(testCredStorePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Error attempting to remove %s", testCredStorePath)
	}
	return nil
}

func getCredStore(t *testing.T) GCRCredStore {
	return &credStore{
		credentialPath: testCredStorePath,
	}
}

func TestGetGCRAuth_GCRCredsPresent(t *testing.T) {
	gcrTokens := tokens{
		AccessToken: testAccessToken,
	}
	creds := &dockerCredentials{
		GCRCreds: &gcrTokens,
	}
	writeCredentialsToStoreFile(t, creds)
	tested := getCredStore(t)

	auth, err := tested.GetGCRAuth()
	if err != nil {
		t.Fatalf("GetGCRAuth returned an error: %v", err)
	}
	actual := auth.initialToken.AccessToken
	if actual != testAccessToken {
		t.Fatalf("Expected access token to be \"%s\", was \"%s\"", testAccessToken, actual)
	}
}

func TestGetGCRAuth_GCRCredsAbsent(t *testing.T) {
	creds := &dockerCredentials{}
	writeCredentialsToStoreFile(t, creds)
	tested := getCredStore(t)

	auth, err := tested.GetGCRAuth()
	if err == nil {
		t.Fatalf("Expected error, got %+v", *auth)
	}
}

// TODO: Fix test for windows.
func TestSetGCRAuth_NoFile(t *testing.T) {
	err := cleanUp()
	if err != nil {
		t.Fatal("Could not guarantee that no credential file existed.")
	}
	tested := getCredStore(t)
	const expctedRefresh = "refreshing!"
	expectedExpiry := time.Now()
	gcrTok := &oauth2.Token{
		AccessToken:  testAccessToken,
		RefreshToken: expctedRefresh,
		Expiry:       expectedExpiry,
	}

	err = tested.SetGCRAuth(gcrTok)
	if err != nil {
		t.Fatalf("SetGCRAuth returned an error: %v", err)
	}

	auth, err := tested.GetGCRAuth()
	if err != nil {
		t.Fatalf("GetGCRAuth returned an error: %v", err)
	}

	actualAccessTok := auth.initialToken.AccessToken
	if actualAccessTok != testAccessToken {
		t.Errorf("access_token: Expected \"%s\", got \"%s\"", testAccessToken, actualAccessTok)
	}
	actualRefresh := auth.initialToken.RefreshToken
	if actualRefresh != expctedRefresh {
		t.Errorf("refresh_token: Expected \"%s\", got \"%s\"", expctedRefresh, actualRefresh)
	}
	actualExp := auth.initialToken.Expiry
	if !actualExp.Equal(expectedExpiry) {
		t.Errorf("Expiry: Expected %v, got %v", expectedExpiry, actualExp)
	}
}

func TestSetGCRAuth_OverwriteOld(t *testing.T) {
	gcrTokens := tokens{
		AccessToken: "old_access_token",
	}
	creds := &dockerCredentials{
		GCRCreds: &gcrTokens,
	}
	writeCredentialsToStoreFile(t, creds)

	tested := getCredStore(t)
	const expctedRefresh = "refreshing!"
	expectedExpiry := time.Now()
	gcrTok := &oauth2.Token{
		AccessToken:  testAccessToken,
		RefreshToken: expctedRefresh,
		Expiry:       expectedExpiry,
	}

	err := tested.SetGCRAuth(gcrTok)
	if err != nil {
		t.Fatalf("SetGCRAuth returned an error: %v", err)
	}

	auth, err := tested.GetGCRAuth()
	if err != nil {
		t.Fatalf("GetGCRAuth returned an error: %v", err)
	}

	actualAccessTok := auth.initialToken.AccessToken
	if actualAccessTok != testAccessToken {
		t.Errorf("access_token: Expected \"%s\", got \"%s\"", testAccessToken, actualAccessTok)
	}
	actualRefresh := auth.initialToken.RefreshToken
	if actualRefresh != expctedRefresh {
		t.Errorf("refresh_token: Expected \"%s\", got \"%s\"", expctedRefresh, actualRefresh)
	}
	actualExp := auth.initialToken.Expiry
	if !actualExp.Equal(expectedExpiry) {
		t.Errorf("Expiry: Expected %v, got %v", expectedExpiry, actualExp)
	}
}

func TestSetGCRAuth_PreserveOthers(t *testing.T) {
	gcrTokens := tokens{
		AccessToken: "old_access_token",
	}
	creds := &dockerCredentials{
		GCRCreds: &gcrTokens,
	}
	writeCredentialsToStoreFile(t, creds)

	tested := getCredStore(t)
	const expctedRefresh = "refreshing!"
	expectedExpiry := time.Now()
	gcrTok := &oauth2.Token{
		AccessToken:  testAccessToken,
		RefreshToken: expctedRefresh,
		Expiry:       expectedExpiry,
	}

	err := tested.SetGCRAuth(gcrTok)
	if err != nil {
		t.Fatalf("SetGCRAuth returned an error: %v", err)
	}

	auth, err := tested.GetGCRAuth()
	if err != nil {
		t.Fatalf("GetGCRAuth returned an error: %v", err)
	}

	actualAccessTok := auth.initialToken.AccessToken
	if actualAccessTok != testAccessToken {
		t.Errorf("access_token: Expected \"%s\", got \"%s\"", testAccessToken, actualAccessTok)
	}
	actualRefresh := auth.initialToken.RefreshToken
	if actualRefresh != expctedRefresh {
		t.Errorf("refresh_token: Expected \"%s\", got \"%s\"", expctedRefresh, actualRefresh)
	}
	actualExp := auth.initialToken.Expiry
	if !actualExp.Equal(expectedExpiry) {
		t.Errorf("Expiry: Expected %v, got %v", expectedExpiry, actualExp)
	}
}

func TestDeleteGCRAuth(t *testing.T) {
	gcrTokens := tokens{
		AccessToken: testAccessToken,
	}
	creds := &dockerCredentials{
		GCRCreds: &gcrTokens,
	}
	writeCredentialsToStoreFile(t, creds)
	tested := getCredStore(t)

	err := tested.DeleteGCRAuth()
	if err != nil {
		t.Fatalf("DeleteGCRAuth returned an error: %v", err)
	}

	auth, err := tested.GetGCRAuth()
	if err == nil {
		t.Fatalf("Expected no credentials, got %+v", *auth)
	}
}

func TestDeleteGCRAuth_GCRCredsAbsent(t *testing.T) {
	creds := &dockerCredentials{}
	writeCredentialsToStoreFile(t, creds)
	tested := getCredStore(t)

	err := tested.DeleteGCRAuth()

	if err != nil {
		t.Fatalf("DeleteGCRAuth returned an error: %v", err)
	}
}

// TODO: Fix test for windows.
func TestDeleteGCRAuth_NoFile(t *testing.T) {
	err := cleanUp()
	if err != nil {
		t.Fatal("Could not guarantee that no credential file existed.")
	}
	tested := getCredStore(t)

	err = tested.DeleteGCRAuth()

	if err != nil {
		t.Fatalf("DeleteGCRAuth returned an error: %v", err)
	}
}

// TODO: Fix test for windows.
func TestGCRAuthLifespan(t *testing.T) {
	err := cleanUp()
	if err != nil {
		t.Fatal("Could not guarantee that no credential file existed.")
	}
	tested := getCredStore(t)
	const expctedRefresh = "refreshing!"
	expectedExpiry := time.Now()
	gcrTok := &oauth2.Token{
		AccessToken:  testAccessToken,
		RefreshToken: expctedRefresh,
		Expiry:       expectedExpiry,
	}

	// set the credentials
	err = tested.SetGCRAuth(gcrTok)
	if err != nil {
		t.Fatalf("SetGCRAuth returned an error: %v", err)
	}

	// retrieve them again
	auth, err := tested.GetGCRAuth()
	if err != nil {
		t.Fatalf("GetGCRAuth returned an error: %v", err)
	}
	actualAccessTok := auth.initialToken.AccessToken
	if actualAccessTok != testAccessToken {
		t.Errorf("access_token: Expected \"%s\", got \"%s\"", testAccessToken, actualAccessTok)
	}
	actualRefresh := auth.initialToken.RefreshToken
	if actualRefresh != expctedRefresh {
		t.Errorf("refresh_token: Expected \"%s\", got \"%s\"", expctedRefresh, actualRefresh)
	}
	actualExp := auth.initialToken.Expiry
	if !actualExp.Equal(expectedExpiry) {
		t.Errorf("Expiry: Expected %v, got %v", expectedExpiry, actualExp)
	}

	// delete them
	err = tested.DeleteGCRAuth()
	if err != nil {
		t.Fatalf("DeleteGCRAuth returned an error: %v", err)
	}

	// make sure they're gone
	auth, err = tested.GetGCRAuth()
	if err == nil {
		t.Fatalf("Expected no credentials, got %v", *auth)
	}
}
