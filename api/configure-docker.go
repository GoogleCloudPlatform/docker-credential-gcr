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
	"strings"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/cliconfig"
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

	hostnames := make([]string, len(config.SupportedGCRRegistries))
	i := 0
	for registry := range config.SupportedGCRRegistries {
		hostnames[i] = registry
		i++
	}

	configModified := addHostNamesToAuths(hostnames, dockerConfig.AuthConfigs)

	// Optimization. Don't modify the dockerConfig if we're already fully configured.
	if configModified || dockerConfig.CredentialsStore != credHelperSuffix {
		dockerConfig.CredentialsStore = credHelperSuffix
		if err = dockerConfig.Save(); err != nil {
			printErrorln("Unable to save docker config: %v", err)
			return subcommands.ExitFailure
		}
	}

	fmt.Printf("%s configured to use %s as its credential store\n", dockerConfig.Filename, binaryName)
	return subcommands.ExitSuccess
}

// Adds all of the default GCR registries defined config.SupportedGCRRegistries,
// if they don't already exist, to the given map as empty types.AuthConfigs.
// Returns true if auths was modified, false otherwise.
func addHostNamesToAuths(hosts []string, auths map[string]types.AuthConfig) bool {
	modified := false
	for _, registry := range hosts {
		// If the registry doesn't already have a scheme, set it to https.
		if !strings.Contains(registry, "://") {
			registry = "https://" + registry
		}
		if _, exists := auths[registry]; !exists {
			// Add an empty auths entry for the registry so that 'docker build' works.
			auths[registry] = types.AuthConfig{}
			modified = true
		}
	}

	return modified
}

func printErrorln(fmtString string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+fmtString+"\n", v...)
}
