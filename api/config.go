package api

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	"github.com/google/subcommands"
	"golang.org/x/net/context"
)

const (
	tokenSourceFlag = "token-source"
	resetAllFlag    = "unset-all"
)

type configCmd struct {
	cmd
	tokenSources string
	resetAll     bool
}

// NewConfigSubcommand returns a subcommands.Command which allows for user
// configuration of cred helper behavior.
func NewConfigSubcommand() subcommands.Command {
	return &configCmd{
		cmd{
			name:     "config",
			synopsis: "configure the credential helper",
		},
		"unused",
		false,
	}
}

func (c *configCmd) SetFlags(fs *flag.FlagSet) {
	validSources := strings.Join(config.DefaultTokenSources[:], ", ")
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
			default:
				printError(flag, errors.New("Unknown flag!"))
				result = subcommands.ExitFailure
			}
		})
	}
	return result
}

func resetAll() error {
	cfg, err := config.NewUserConfig()
	if err != nil {
		return err
	}
	return cfg.ResetAll()
}

func setTokenSources(rawSource string) error {
	cfg, err := config.NewUserConfig()
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
