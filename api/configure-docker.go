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

package api

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	"github.com/GoogleCloudPlatform/docker-credential-gcr/store"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/cliconfig/configfile"
	"github.com/google/subcommands"
	"golang.org/x/net/context"
)

type dockerConfigCmd struct {
	cmd
	overwrite bool // overwrite any previously configured credential store
}

// see https://github.com/docker/docker/blob/master/cliconfig/credentials/native_store.go
const credHelperPrefix = "docker-credential-"

// NewDockerConfigSubcommand returns a subcommands.Command which configures
// the docker client to use this credential helper
func NewDockerConfigSubcommand() subcommands.Command {
	return &dockerConfigCmd{
		cmd{
			name:     "configure-docker",
			synopsis: fmt.Sprintf("configures the Docker client to use %s", os.Args[0]),
		},
		false,
	}
}

func (c *dockerConfigCmd) SetFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.overwrite, "overwrite", false, "overwrite any previously configured credential helper")
}

func (c *dockerConfigCmd) Execute(context.Context, *flag.FlagSet, ...interface{}) subcommands.ExitStatus {
	binaryName := filepath.Base(os.Args[0])
	if !strings.HasPrefix(binaryName, credHelperPrefix) {
		printErrorln("Binary name must be prefixed with '%s': %s", credHelperPrefix, binaryName)
		return subcommands.ExitFailure
	}

	// the Docker client can only use binaries on the $PATH
	if _, err := exec.LookPath(binaryName); err != nil {
		printErrorln("'%s' must exist on your PATH", binaryName)
		return subcommands.ExitFailure
	}

	dockerConfig, err := cliconfig.Load("")
	if err != nil {
		printErrorln("Unable to load docker config: %v", err)
		return subcommands.ExitFailure
	}

	// 'credsStore' takes the suffix of the credential helper binary
	credHelperSuffix := binaryName[len(credHelperPrefix):]

	// Only proceed if the creds store is empty or we're allowed to overwrite.
	if dockerConfig.CredentialsStore != credHelperSuffix && dockerConfig.CredentialsStore != "" && !c.overwrite {
		printErrorln("Docker is currently configured to use '%s%s' as its credential store. Retry with --overwrite", credHelperPrefix, dockerConfig.CredentialsStore)
		return subcommands.ExitFailure
	}

	// Populate the AuthConfigs exclusivley with GCR registries.
	// 'docker build' work on 1.11 and 1.12.
	s, err := store.NewGCRCredStore()
	if err != nil {
		printErrorln("Unable to read credentialStore: %v", err)
	}
	authsModified := setAuthConfigs(dockerConfig, s)

	// Optimization. Don't modify the dockerConfig if we're already fully configured.
	if authsModified || dockerConfig.CredentialsStore != credHelperSuffix {
		// Overwrite the existing set of AuthConfigs since they aren't visible anymore, anyway.
		dockerConfig.CredentialsStore = credHelperSuffix
		if err = dockerConfig.Save(); err != nil {
			printErrorln("Unable to save docker config: %v", err)
			return subcommands.ExitFailure
		}
	}

	fmt.Printf("%s configured to use %s as its credential store\n", dockerConfig.Filename, binaryName)
	return subcommands.ExitSuccess
}

// Ensures that the AuthConfigs in the given ConfigFile are exactly the set
// of config.SupportedGCRRegistries with the https scheme plus any 3p creds
// we have stored.
// Returns true if the ConfigFile was modified, false otherwise.
func setAuthConfigs(dockerConfig *configfile.ConfigFile, s store.GCRCredStore) bool {
	newAuthconfigs := make(map[string]types.AuthConfig)
	for registry := range config.SupportedGCRRegistries {
		registry = "https://" + registry
		newAuthconfigs[registry] = types.AuthConfig{}
	}

	creds, err := s.AllThirdPartyCreds()
	// Only add 3p creds if we can retrieve them, but FUBAR cred store is OK
	if err == nil {
		for registry := range creds {
			newAuthconfigs[registry] = types.AuthConfig{}
		}
	}

	if !reflect.DeepEqual(newAuthconfigs, dockerConfig.AuthConfigs) {
		dockerConfig.AuthConfigs = newAuthconfigs
		return true
	}

	return false
}

func printErrorln(fmtString string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+fmtString+"\n", v...)
}
