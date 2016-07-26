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
	"strings"

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
	binaryName := os.Args[0]
	if !strings.HasPrefix(binaryName, credHelperPrefix) {
		printErrorln("Binary name must be prefixed with '%s': %s", credHelperPrefix, binaryName)
		return subcommands.ExitFailure
	}
	
	// the Docker client can only use binaries on the $PATH
	if _, err := exec.LookPath(binaryName); err != nil {
		printErrorln("'%s' must exist on your PATH", binaryName)
		return subcommands.ExitFailure
	}

	config, err := cliconfig.Load("")
	if err != nil {
		printErrorln("Unable to load docker config: %v", err)
		return subcommands.ExitFailure
	}

	// 'credsStore' takes the suffix of the credential helper binary
	credHelperSuffix := binaryName[len(credHelperPrefix):]

	// Optimization. Don't modify the config if we're already configured.
	if config.CredentialsStore != credHelperSuffix {
		if config.CredentialsStore != "" && !c.overwrite {
			printErrorln("Docker is currently configured to use '%s%s' as its credential store. Retry with --overwrite", credHelperPrefix, config.CredentialsStore)
			return subcommands.ExitFailure
		}

		config.CredentialsStore = credHelperSuffix

		if err = config.Save(); err != nil {
			printErrorln("Unable to save docker config: %v", err)
			return subcommands.ExitFailure
		}
	}

	fmt.Printf("%s configured to use %s as its credential store\n", config.Filename, binaryName)
	return subcommands.ExitSuccess
}

func printErrorln(fmtString string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+fmtString+"\n", v...)
}
