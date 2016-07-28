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

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	"github.com/google/subcommands"
	"golang.org/x/net/context"
)

type versionCmd struct {
	cmd
}

// NewVersionSubcommand returns a subcommands.Command which prints the binary
// version to stdout.
func NewVersionSubcommand() subcommands.Command {
	return &versionCmd{
		cmd{
			name:     "version",
			synopsis: "print the version of the binary to stdout",
		},
	}
}

func (p *versionCmd) Execute(context.Context, *flag.FlagSet, ...interface{}) subcommands.ExitStatus {
	fmt.Fprintf(os.Stdout, "GCR Docker Credential Helper v%d.%d\n", config.VersionMajor, config.VersionMinor)
	return subcommands.ExitSuccess
}
