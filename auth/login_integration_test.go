//go:build !race
// +build !race

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

package auth

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	"golang.org/x/oauth2"
)

const (
	// The client ID corresponding to GCR's OAuth2 login page.
	expectedClientID  = "99426463878-o7n0bshgue20tdpm25q4at0vs2mr4utq.apps.googleusercontent.com"
	expectedScope     = "https://www.googleapis.com/auth/cloud-platform"
	expectedHost      = "localhost"
	expectedAuthPath  = "/auth"
	expectedTokenPath = "/token"
	expectedGrantType = "authorization_code"

	expectedCode         = "sUp3r@w3$om3c0d3"
	expectedAccessToken  = "@cce$$4dayz"
	expectedRefreshToken = "refreshplz"
	expectedTTL          = 3600
)

func initAuthServer() (net.Listener, error) {
	testLn, testPort, err := getListener()
	if err != nil {
		return nil, err
	}
	config.GCROAuth2Endpoint = oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("http://%s:%d%s", expectedHost, testPort, expectedAuthPath),
		TokenURL: fmt.Sprintf("http://%s:%d%s", expectedHost, testPort, expectedTokenPath),
	}
	return testLn, nil
}

type testBrowser struct {
	shouldSucceed bool
	t             *testing.T
	RedirectURL   chan *url.URL
	State         chan string
}

// the testBrowser's Open method exists to verify the URL which is passed
// when attempting to open the default web browser during the login operation
func (b *testBrowser) Open(urlStr string) error {
	if !b.shouldSucceed {
		return errors.New("you asked for it")
	}

	URL, err := url.Parse(urlStr)

	if err != nil {
		b.t.Errorf("Could not parse URL: %s", urlStr)
		return nil
	}

	if URL.Path != expectedAuthPath {
		b.t.Errorf("Expected Path to be: %s, got: %s", expectedAuthPath, URL.Path)
	}
	if !strings.HasPrefix(URL.Host, expectedHost) {
		b.t.Errorf("Expected Host to begin with: %s, got: %s", expectedHost, URL.Host)
	}

	responseType := URL.Query().Get("response_type")
	if responseType != "code" {
		b.t.Errorf("Expected response_type: %s, got: %s", "code", responseType)
	}

	clientID := URL.Query().Get("client_id")
	if clientID != expectedClientID {
		b.t.Errorf("Expected client_id: %s, got: %s", expectedClientID, clientID)
	}

	redirectURI := URL.Query().Get("redirect_uri")
	redirURL, err := url.Parse(redirectURI)
	if err != nil {
		b.t.Errorf("Unable to parse redirect_uri: %s", redirectURI)
	} else {
		// pass the redirect URL to the browser
		b.RedirectURL <- redirURL

		if !strings.HasPrefix(redirURL.Host, expectedHost) {
			b.t.Errorf("RedirectURL should begin with %s: %s", expectedHost, redirURL.Host)
		}
	}

	scope := URL.Query().Get("scope")
	if scope != expectedScope {
		b.t.Errorf("Expected scope: %s, got: %s", expectedScope, scope)
	}

	// pass the 'state' variable to the browser thread
	b.State <- URL.Query().Get("state")

	return nil
}

func performBrowserActions(t *testing.T, browser *testBrowser) {
	// simulate the authorization code redirect after the user
	// performs the login flow
	redirURL := <-browser.RedirectURL
	state := <-browser.State
	args := redirURL.Query()
	args.Set("code", expectedCode)
	args.Set("state", state)
	redirURL.RawQuery = args.Encode()
	resp, err := http.Get(redirURL.String())
	if err != nil {
		t.Fatalf("Could not send authorization code response: %v", err)
	}
	defer resp.Body.Close()
	// the browser should receive a response AFTER the entire auth flow has completed
	if resp.StatusCode >= 400 {
		t.Errorf("Unsuccessful response: %+v", *resp)
	}
}

func performAuthServerActions(t *testing.T, testLn net.Listener) {
	// perform the auth server-side actions...
	// receive the authorization_code exchange request
	conn, err := testLn.Accept()
	if err != nil {
		t.Errorf("Could not accept tcp connection: %v", err)
	}

	srvConn := httputil.NewServerConn(conn, nil)
	defer srvConn.Close()

	req, err := srvConn.Read()
	if err != nil {
		t.Fatalf("Could not read from connection: %v", err)
	}

	if req.URL.Path != expectedTokenPath {
		t.Errorf("Expected path: %s, got %s", expectedTokenPath, req.URL.Path)
	}

	grantType := req.PostFormValue("grant_type")
	if grantType != expectedGrantType {
		t.Errorf("Expected grant_type: %s, got: %s", expectedGrantType, grantType)
	}

	clientID := req.PostFormValue("client_id")
	if clientID == "" {
		// Newer google oauth libraries deliver client_id in the Authorization header.
		if username, _, headerExists := req.BasicAuth(); !headerExists || username != expectedClientID {
			t.Errorf("Expected username: %s, got: %s", expectedClientID, username)
		}
	} else if clientID != expectedClientID {
		// Older libraries use a client_id form value.
		t.Errorf("Expected client_id: %s, got: %s", expectedClientID, clientID)
	}
	redirectURI := req.PostFormValue("redirect_uri")
	if redirectURI == "" {
		t.Errorf("Expected redirect_uri to be present: %+v", *req)
	}
	code := req.PostFormValue("code")
	if code != expectedCode {
		t.Errorf("Expected authorization_code: %s, got: %s", expectedCode, code)
	}

	if t.Failed() {
		t.Errorf("Request: %+v", *req)
	}

	// Respond with the access_token, refresh_token
	var resp http.Response
	bodyString := fmt.Sprintf(`{
		"access_token":"%s",
		"expires_in":%d,
		"token_type":"%s",
		"refresh_token":"%s"
	}`, expectedAccessToken, expectedTTL, "Bearer", expectedRefreshToken)
	resp.Body = getReadCloserFromString(bodyString)
	resp.StatusCode = 200
	resp.Proto = "HTTP/1.1"
	resp.Close = true
	resp.ContentLength = -1   // unknown length
	srvConn.Write(req, &resp) // ignore errors; expected
}

// turn a string into a ReadCloser as required by http Responses
func getReadCloserFromString(body string) io.ReadCloser {
	reader := strings.NewReader(body)
	return ioutil.NopCloser(reader)
}

// TestBrowserFlow tests much of the flow required to authenticate a GCR user
// via an automatically-launched user agent. The browser is mocked for testing
// feasibility, and the OAuth2 endpoints are reditected to the localhost.
func TestBrowserAllowed(t *testing.T) {
	testLn, err := initAuthServer()
	if err != nil {
		t.Fatalf("Unable to initialize auth server: %v", err)
	}
	defer testLn.Close()
	mockBrowser := &testBrowser{
		shouldSucceed: true,
		t:             t,
		RedirectURL:   make(chan *url.URL),
		State:         make(chan string),
	}
	defer close(mockBrowser.RedirectURL)
	defer close(mockBrowser.State)

	// start a goroutine to act as the browser
	go performBrowserActions(t, mockBrowser)

	// start a goroutine to act as the auth server
	go performAuthServerActions(t, testLn)

	// test the client-side code
	tested := &GCRLoginAgent{
		AllowBrowser: true,
		OpenBrowser:  mockBrowser.Open,
	}
	tok, err := tested.PerformLogin()
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if tok.AccessToken != expectedAccessToken {
		t.Errorf("Expected access_token: %s, got: %s", expectedAccessToken, tok.AccessToken)
	}
	if tok.RefreshToken != expectedRefreshToken {
		t.Errorf("Expected refresh_token: %s, got: %s", expectedRefreshToken, tok.RefreshToken)
	}
	if tok.TokenType != "Bearer" {
		t.Errorf("Expected token_type: %s, got: %s", "Bearer", tok.TokenType)
	}
}

func verifyCommandLineURL(t *testing.T, urlStr string) {
	URL, err := url.Parse(urlStr)

	if err != nil {
		t.Errorf("Could not parse URL: %s", urlStr)
	} else {
		if URL.Path != expectedAuthPath {
			t.Errorf("Expected Path to be: %s, got: %s", expectedAuthPath, URL.Path)
		}
		if !strings.HasPrefix(URL.Host, expectedHost) {
			t.Errorf("Expected Host to begin with: %s, got: %s", expectedHost, URL.Host)
		}

		responseType := URL.Query().Get("response_type")
		if responseType != "code" {
			t.Errorf("Expected response_type: %s, got: %s", "code", responseType)
		}

		clientID := URL.Query().Get("client_id")
		if clientID != expectedClientID {
			t.Errorf("Expected client_id: %s, got: %s", expectedClientID, clientID)
		}

		redirectURI := URL.Query().Get("redirect_uri")
		if redirectURI != "urn:ietf:wg:oauth:2.0:oob" {
			t.Errorf("Expected redirect_uri: %s, got %s", "urn:ietf:wg:oauth:2.0:oob", redirectURI)
		}

		scope := URL.Query().Get("scope")
		if scope != expectedScope {
			t.Errorf("Expected scope: %s, got: %s", expectedScope, scope)
		}
	}
}

func performCommandLineInteraction(t *testing.T, outRead io.Reader) {
	var line string
	for {
		fmt.Fscanln(outRead, &line)
		line = strings.TrimSpace(line)
		if line == "Please visit the following URL and complete the authorization dialog:" {
			// expected
		} else if strings.HasPrefix(line, "http") {
			verifyCommandLineURL(t, line)
		} else if line == "Authorization code:" {
			break
		} else {
			t.Errorf("Unexpected line: %s", line)
			break
		}
	}
}

// multiThreadReadWriter is a io.ReadWriter that allows for simpler
// communication than is afforded by io.Pipe
// TODO: Fix race condition when reading/writing using same slice.
type multiThreadReadWriter struct {
	c chan []byte
}

func (m multiThreadReadWriter) Read(p []byte) (int, error) {
	bytes := <-m.c
	return copy(p, bytes), nil
}
func (m multiThreadReadWriter) Write(p []byte) (int, error) {
	m.c <- p
	return len(p), nil
}
func newMultiThreadReadWriter() io.ReadWriter {
	return multiThreadReadWriter{
		c: make(chan []byte),
	}
}

func TestBrowserForbidden(t *testing.T) {
	mockStdout := newMultiThreadReadWriter()
	mockStdin := strings.NewReader(expectedCode)

	testLn, err := initAuthServer()
	if err != nil {
		t.Fatalf("Unable to initialize auth server: %v", err)
	}
	defer testLn.Close()

	// start a goroutine to act as the auth server and verify interactions
	// with the client
	go performAuthServerActions(t, testLn)

	// start a goroutine to act as a command-line user and verify interactions
	// with the client
	go performCommandLineInteraction(t, mockStdout)

	tested := &GCRLoginAgent{
		AllowBrowser: false,
		In:           mockStdin,
		Out:          mockStdout,
	}
	tok, err := tested.PerformLogin()

	// verify the token
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if tok.AccessToken != expectedAccessToken {
		t.Errorf("Expected access_token: %s, got: %s", expectedAccessToken, tok.AccessToken)
	}
	if tok.RefreshToken != expectedRefreshToken {
		t.Errorf("Expected refresh_token: %s, got: %s", expectedRefreshToken, tok.RefreshToken)
	}
	if tok.TokenType != "Bearer" {
		t.Errorf("Expected token_type: %s, got: %s", "Bearer", tok.TokenType)
	}
}

func TestBrowserAllowed_BrowserOpenFails(t *testing.T) {
	mockStdout := newMultiThreadReadWriter()
	mockStdin := strings.NewReader(expectedCode)

	testLn, err := initAuthServer()
	if err != nil {
		t.Fatalf("Unable to initialize auth server: %v", err)
	}
	defer testLn.Close()
	mockBrowser := &testBrowser{
		shouldSucceed: false,
		t:             t,
		RedirectURL:   nil,
		State:         nil,
	}

	// start a goroutine to act as the auth server and verify interactions
	// with the client
	go performAuthServerActions(t, testLn)

	// start a goroutine to act as a command-line user and verify interactions
	// with the client
	go performCommandLineInteraction(t, mockStdout)

	tested := &GCRLoginAgent{
		AllowBrowser: true,
		In:           mockStdin,
		Out:          mockStdout,
		OpenBrowser:  mockBrowser.Open,
	}
	tok, err := tested.PerformLogin()
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if tok.AccessToken != expectedAccessToken {
		t.Errorf("Expected access_token: %s, got: %s", expectedAccessToken, tok.AccessToken)
	}
	if tok.RefreshToken != expectedRefreshToken {
		t.Errorf("Expected refresh_token: %s, got: %s", expectedRefreshToken, tok.RefreshToken)
	}
	if tok.TokenType != "Bearer" {
		t.Errorf("Expected token_type: %s, got: %s", "Bearer", tok.TokenType)
	}
}
