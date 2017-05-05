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
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	"github.com/GoogleCloudPlatform/docker-credential-gcr/store"
	"github.com/GoogleCloudPlatform/docker-credential-gcr/util"
	"github.com/docker/docker/api/types"
	cliconfig "github.com/docker/docker/cli/config"
	"github.com/docker/docker/cli/config/configfile"
	"github.com/google/subcommands"
)

type dockerConfigCmd struct {
	cmd
	// overwrite any previously configured credential store and/or credentials
	overwrite bool
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
	fs.BoolVar(&c.overwrite, "overwrite", false, "overwrite any previously configured credential store and/or credentials")
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

	// 'credsStore' and 'credHelpers' take the suffix of the credential helper
	// binary.
	credHelperSuffix := binaryName[len(credHelperPrefix):]

	major, minor, _, _, err := util.DockerClientVersion()
	if err != nil {
		printErrorln("Unable to determine Docker version: %v", err)
		fmt.Printf("WARNING: Configuring %s as a registry-specific helper. This is only supported by Docker client versions 1.13+\n", binaryName)
		return setConfig(dockerConfig, credHelperSuffix)
	} else if credHelpersSupported(major, minor) {
		// If we can act as a registry-specific credential helper, do so...
		return setConfig(dockerConfig, credHelperSuffix)
	} else if credsStoreSupported(major, minor) {
		// Else, attempt to act as the cred store...
		return c.setLegacyConfig(dockerConfig, credHelperSuffix)
	}

	// Netiher cred helper nor cred store is supported by the detected docker
	// version.
	fmt.Fprintln(os.Stderr, "ERROR: Docker client version 1.10+ required")
	return subcommands.ExitFailure
}

// credHelpersSupported returns true if the installed version of Docker supports
// credential helpers (1.13+), error if Docker is not installed or there was an
// error determining the version.
func credHelpersSupported(majorVersion, minorVersion int) bool {
	return majorVersion >= 17 || (majorVersion == 1 && minorVersion >= 13)
}

// credsStoreSupported returns true if the installed version of Docker supports
// credential stores (1.11+), error if Docker is not installed or there was an
// error determining the version.
func credsStoreSupported(majorVersion, minorVersion int) bool {
	return majorVersion >= 17 || (majorVersion == 1 && minorVersion >= 11)
}

// Configure Docker to use the credential helper for GCR's registries only.
// Defining additional 'auths' entries is unnecessary in versions which
// support registry-specific credential helpers.
func setConfig(dockerConfig *configfile.ConfigFile, helperSuffix string) subcommands.ExitStatus {
	// We always overwrite since there's no way that we can accidentally
	// disable other credentials as a registry-specific credential helper.
	if dockerConfig.CredentialHelpers == nil {
		dockerConfig.CredentialHelpers = map[string]string{}
	}

	for registry := range config.SupportedGCRRegistries {
		dockerConfig.CredentialHelpers[registry] = helperSuffix
	}

	if err := dockerConfig.Save(); err != nil {
		printErrorln("Unable to save docker config: %v", err)
		return subcommands.ExitFailure
	}

	fmt.Printf("%s configured to use this credential helper for GCR registries\n", dockerConfig.Filename)
	return subcommands.ExitSuccess
}

// Configure Docker to use the credential helper as the default credential
// store. Also add 'auths' entries to the config to support versions where the
// docker config was the source of truth for the set of stored credentials.
func (c *dockerConfigCmd) setLegacyConfig(dockerConfig *configfile.ConfigFile, helperSuffix string) subcommands.ExitStatus {
	// Only proceed if the creds store is empty or we're allowed to overwrite.
	// Replacing a cred store effectivley makes any previously stored
	// credentials unreachable.
	otherCredStoreConfigured := dockerConfig.CredentialsStore != helperSuffix && dockerConfig.CredentialsStore != ""
	dcgcrConfiguredAsCredStore := dockerConfig.CredentialsStore == helperSuffix
	credentialsStored := len(dockerConfig.AuthConfigs) > 0
	if otherCredStoreConfigured && !c.overwrite {
		// If another credential store is configured, demand explicit
		// overwrite permissions.
		printErrorln("Docker is currently configured to use '%s%s' as its credential store. Please retry with --overwrite. This will render any previously store credentials unaccessible.", credHelperPrefix, dockerConfig.CredentialsStore)
		return subcommands.ExitFailure
	} else if credentialsStored && !dcgcrConfiguredAsCredStore && !c.overwrite {
		// If there are credentials stored somewhere other than this credential
		// helper, demand explicit overwrite permissions.
		printErrorln("%d credentials are currently stored which would be overwritten. Retry with --overwrite.", len(dockerConfig.AuthConfigs))
		return subcommands.ExitFailure
	}

	// Populate the AuthConfigs portion of the config.
	// This allows 'docker build' work on Docker client versions 1.11 and 1.12,
	// where AuthConfigs was
	s, err := store.NewGCRCredStore()
	if err != nil {
		printErrorln("Unable to read credentialStore: %v", err)
	}
	authsModified := setAuthConfigs(dockerConfig, s)

	// Optimization. Don't modify the dockerConfig if we're already fully configured.
	if authsModified || dockerConfig.CredentialsStore != helperSuffix {
		// Overwrite the existing set of AuthConfigs since they aren't visible anymore, anyway.
		dockerConfig.CredentialsStore = helperSuffix
		if err = dockerConfig.Save(); err != nil {
			printErrorln("Unable to save docker config: %v", err)
			return subcommands.ExitFailure
		}
	}

	fmt.Printf("%s successfully configured\n", dockerConfig.Filename)
	if c.overwrite {
		fmt.Println("Any previously stored credentials have been overwritten.")
	}
	return subcommands.ExitSuccess
}

// Ensures that the AuthConfigs in the given ConfigFile are exactly the set
// of config.SupportedGCRRegistries with the https scheme plus any 3p creds
// we have stored.
// Returns true if the ConfigFile was modified, false otherwise.
func setAuthConfigs(dockerConfig *configfile.ConfigFile, s store.GCRCredStore) bool {
	newAuthconfigs := make(map[string]types.AuthConfig)
	for registry := range config.SupportedGCRRegistries {
		// 'auths' members take the HTTP scheme
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
