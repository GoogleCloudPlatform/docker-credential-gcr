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
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	"github.com/google/subcommands"
)

const (
	defaultToGCRTokenFlag = "default-to-gcr-access-token"
	tokenSourceFlag       = "token-source"
	resetAllFlag          = "unset-all"
)

type configCmd struct {
	cmd
	defaultToGCRToken bool
	tokenSources      string
	resetAll          bool
}

// NewConfigSubcommand returns a subcommands.Command which allows for user
// configuration of cred helper behavior.
func NewConfigSubcommand() subcommands.Command {
	return &configCmd{
		cmd{
			name:     "config",
			synopsis: "configure the credential helper",
		},
		// Because only specified flags are iterated by FlagSet.Visit,
		// these values will always be explicitly set by the user if visited.
		false,
		"unused",
		false,
	}
}

func (c *configCmd) SetFlags(fs *flag.FlagSet) {
	validSources := strings.Join(config.DefaultTokenSources[:], ", ")
	fs.BoolVar(&c.defaultToGCRToken, defaultToGCRTokenFlag, false, "If enabled, the credential helper will attempt to return GCR's access token, rather than an error, when credentials cannot otherwise be found.")
	fs.StringVar(&c.tokenSources, tokenSourceFlag, validSources, "The source(s), in order, to search for GCR credentials. Valid values are: "+validSources+". ")
	fs.BoolVar(&c.resetAll, resetAllFlag, false, "Resets all settings to default.")
}

func (c *configCmd) Execute(_ context.Context, flags *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	result := subcommands.ExitSuccess

	//TODO(jsand@google.com): un-uglify
	if c.resetAll {
		if err := resetAll(); err != nil {
			printError(resetAllFlag, err)
			result = subcommands.ExitFailure
		} else {
			printSuccess("Config reset.")
		}
	} else {
		flags.Visit(func(f *flag.Flag) {
			flag := f.Name
			// FlagSet.XXXVar was used instead of Flag.Value for sanity purposes,
			// but FlagSet.Visit is useful to only process explicitly set flags.
			switch flag {
			case tokenSourceFlag:
				if err := setTokenSources(c.tokenSources); err != nil {
					printError(tokenSourceFlag, err)
					result = subcommands.ExitFailure
				} else {
					printSuccess("Token source(s) set.")
				}
			case defaultToGCRTokenFlag:
				if err := setDefaultToGCR(c.defaultToGCRToken); err != nil {
					printError(defaultToGCRTokenFlag, err)
					result = subcommands.ExitFailure
				} else if c.defaultToGCRToken {
					printSuccess("Will attempt to return GCR's access token when no other credentials are found.")
				} else {
					printSuccess("Will return an error when credentials are not found.")
				}
			default:
				printError(flag, errors.New("Unknown flag!"))
				result = subcommands.ExitFailure
			}
		})
	}
	return result
}

func resetAll() error {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return err
	}
	return cfg.ResetAll()
}

func setDefaultToGCR(defaultToGCR bool) error {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return err
	}
	return cfg.SetDefaultToGCRAccessToken(defaultToGCR)
}

func setTokenSources(rawSource string) error {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return err
	}
	strReader := strings.NewReader(rawSource)
	sources, err := csv.NewReader(strReader).Read()
	if err != nil {
		return err
	}
	for i, src := range sources {
		sources[i] = strings.TrimSpace(src)
	}
	return cfg.SetTokenSources(sources)
}

func printSuccess(msg string) {
	fmt.Fprintf(os.Stdout, "Success: %s\n", msg)
}

func printError(flag string, err error) {
	fmt.Fprintf(os.Stderr, "Failure: %s: %v\n", flag, err)
}
