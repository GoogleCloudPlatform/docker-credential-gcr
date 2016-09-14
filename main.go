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
Program docker-credential-gcr implements the Docker credential helper API
and allows for more advanced login/authentication schemes for GCR customers.

See README.md
*/
package main

import (
	"flag"
	"os"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/api"
	"github.com/google/subcommands"
	"golang.org/x/net/context"
)

const (
	gcrGroup             = "GCR authentication"
	dockerCredStoreGroup = "Docker credential store API"
	configGroup          = "Config"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(api.NewStoreSubcommand(), dockerCredStoreGroup)
	subcommands.Register(api.NewGetSubcommand(), dockerCredStoreGroup)
	subcommands.Register(api.NewEraseSubcommand(), dockerCredStoreGroup)
	subcommands.Register(api.NewListSubcommand(), dockerCredStoreGroup)
	subcommands.Register(api.NewGCRLoginSubcommand(), gcrGroup)
	subcommands.Register(api.NewGCRLogoutSubcommand(), gcrGroup)
	subcommands.Register(api.NewDockerConfigSubcommand(), configGroup)
	subcommands.Register(api.NewConfigSubcommand(), configGroup)
	subcommands.Register(api.NewVersionSubcommand(), "")
	subcommands.Register(api.NewClearSubcommand(), "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
