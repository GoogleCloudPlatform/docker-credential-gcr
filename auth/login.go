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
Package auth implements the logic required to authenticate the user and
generate access tokens for use with GCR.
*/
package auth

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	"github.com/toqueteos/webbrowser"
	"golang.org/x/oauth2"
)

const redirectURIAuthCodeInTitleBar = "urn:ietf:wg:oauth:2.0:oob"

// GCRLoginAgent implements the OAuth2 login dance, generating an Oauth2 access_token
// for the user. If AllowBrowser is set to true, the agent will attempt to
// obtain an authorization_code automatically by executing OpenBrowser and
// reading the redirect performed after a successful login. Otherwise, it will
// attempt to use In and Out to direct the user to the login portal and recieve
// the authorization_code in response.
type GCRLoginAgent struct {
	// Whether to execute OpenBrowser when authenticating the user.
	AllowBrowser bool

	// Read input from here; if nil, uses os.Stdin.
	In io.Reader

	// Write output to here; if nil, uses os.Stdout.
	Out io.Writer

	// Open the browser for the given url.  If nil, uses webbrowser.Open.
	OpenBrowser func(url string) error
}

// populate missing fields as described in the struct definition comments
func (a *GCRLoginAgent) init() {
	if a.In == nil {
		a.In = os.Stdin
	}
	if a.Out == nil {
		a.Out = os.Stdout
	}
	if a.OpenBrowser == nil {
		a.OpenBrowser = webbrowser.Open
	}
}

// PerformLogin performs the auth dance necessary to obtain an
// authorization_code from the user and exchange it for an Oauth2 access_token.
func (a *GCRLoginAgent) PerformLogin() (*oauth2.Token, error) {
	a.init()
	conf := &oauth2.Config{
		ClientID:     config.GCRCredHelperClientID,
		ClientSecret: config.GCRCredHelperClientNotSoSecret,
		Scopes:       config.GCRScopes,
		Endpoint:     config.GCROAuth2Endpoint,
	}

	var code string
	var err error

	if a.AllowBrowser {
		// Attempt to recieve the authorization code via the redirect URL
		ln, port, err := getListener()
		if err == nil {
			defer ln.Close()
			// open a web browser and listen on the redirect URL port
			conf.RedirectURL = fmt.Sprintf("http://localhost:%d", port)
			url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
			err = a.OpenBrowser(url)
			if err == nil {
				code, err = handleCodeResponse(ln)
			}
		}
	}

	// If we shouldn't or can't open a browser, default to a command line
	// prompt.
	if code == "" {
		code, err = a.codeViaPrompt(conf)
		if err != nil {
			return nil, err
		}
	}

	return conf.Exchange(config.OAuthHTTPContext, code)
}

func (a *GCRLoginAgent) codeViaPrompt(conf *oauth2.Config) (string, error) {
	// Direct the user to our login portal
	conf.RedirectURL = redirectURIAuthCodeInTitleBar
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Fprintln(a.Out, "Please visit the following URL and complete the authorization dialog:")
	fmt.Fprintf(a.Out, "%v\n", url)

	// Receive the authorization_code in response
	fmt.Fprintln(a.Out, "Authorization code:")
	var code string
	if _, err := fmt.Fscan(a.In, &code); err != nil {
		return "", err
	}

	return code, nil
}

func getListener() (net.Listener, int, error) {
	laddr := net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0} // port: 0 == find free port
	ln, err := net.ListenTCP("tcp4", &laddr)
	if err != nil {
		return nil, 0, err
	}
	return ln, ln.Addr().(*net.TCPAddr).Port, nil
}

func handleCodeResponse(ln net.Listener) (string, error) {
	conn, err := ln.Accept()
	if err != nil {
		return "", err
	}

	srvConn := httputil.NewServerConn(conn, nil)
	defer srvConn.Close()

	req, err := srvConn.Read()
	if err != nil {
		return "", err
	}

	code := req.URL.Query().Get("code")

	resp := &http.Response{
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Close:         true,
		ContentLength: -1, // designates unknown length
	}
	defer srvConn.Write(req, resp)

	// If the code couldn't be obtained, inform the user via the browser and
	// return an error.
	// TODO i18n?
	if code == "" {
		err := fmt.Errorf("Code not present in response: %s", req.URL.String())
		resp.Body = getResponseBody("ERROR: Authenitcation code not present in response, please retry with --no-browser.")
		return "", err
	}

	resp.Body = getResponseBody("Success! You may now close your browser.")
	return code, nil
}

// turn a string into an io.ReadCloser as required by an http.Response
func getResponseBody(body string) io.ReadCloser {
	reader := strings.NewReader(body)
	return ioutil.NopCloser(reader)
}
