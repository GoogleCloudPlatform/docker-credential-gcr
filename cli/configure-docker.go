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
	"strings"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
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

	return setConfig(dockerConfig, credHelperSuffix)
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

	for registry := range config.DefaultGCRRegistries {
		dockerConfig.CredentialHelpers[registry] = helperSuffix
	}

	if err := dockerConfig.Save(); err != nil {
		printErrorln("Unable to save docker config: %v", err)
		return subcommands.ExitFailure
	}

	fmt.Printf("%s configured to use this credential helper for GCR registries\n", dockerConfig.Filename)
	return subcommands.ExitSuccess
}

func printErrorln(fmtString string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+fmtString+"\n", v...)
}
