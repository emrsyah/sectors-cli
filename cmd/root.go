// Package cmd assembles the sectors CLI command tree. Each market lives in its
// own subpackage (cmd/idx, cmd/sgx, …) exposing a NewCmd() constructor; this
// package wires them onto the root command and owns the global flags.
package cmd

import (
	"errors"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
	"github.com/supertypeai/sectors-cli/cmd/idx"
	"github.com/supertypeai/sectors-cli/cmd/klse"
	"github.com/supertypeai/sectors-cli/cmd/mining"
	"github.com/supertypeai/sectors-cli/cmd/sgx"
	"github.com/supertypeai/sectors-cli/internal/output"
)

// Version is the CLI version, overridden at build time via
// -ldflags "-X github.com/supertypeai/sectors-cli/cmd.Version=v1.2.3"
// (goreleaser sets this automatically).
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "sectors",
	Version: Version,
	Short:   "CLI for the Sectors Financial API (IDX, SGX, KLSE, mining)",
	Long: `sectors is a command-line client for the Sectors Financial API v2.

It is built to be driven by humans and AI agents alike: output defaults to the
API's raw JSON (pretty-printed in a terminal, compact when piped), every command
exits non-zero with a JSON error on stderr when a request fails, and --help on
any command documents its parameters straight from the API spec.

Authenticate with ` + "`sectors auth login --api-key <key>`" + ` or set the
SECTORS_API_KEY environment variable. Get a key at https://sectors.app/api.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		s, _ := cmd.Flags().GetString(cmdutil.FlagOutput)
		if _, err := output.ParseFormat(s); err != nil {
			return err
		}
		return nil
	},
}

// Execute runs the root command and exits the process with an appropriate code.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		var handled *cmdutil.HandledError
		if errors.As(err, &handled) {
			// The command already rendered its error; exit with its code.
			os.Exit(handled.Code)
		}
		// A framework/validation error cobra didn't print (SilenceErrors).
		output.EmitError(0, err.Error(), nil)
		os.Exit(1)
	}
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.String(cmdutil.FlagAPIKey, "", "Sectors API key (overrides $SECTORS_API_KEY and config file)")
	pf.String(cmdutil.FlagBaseURL, "", "override the API base URL (default https://api.sectors.app)")
	pf.StringP(cmdutil.FlagOutput, "o", "auto", "output format: auto, json, pretty, or table")
	pf.Duration(cmdutil.FlagTimeout, 30*time.Second, "overall HTTP request timeout (incl. retries)")
	pf.Int(cmdutil.FlagRetries, 3, "max retries for transient failures (429/5xx/network); 0 disables")
	pf.Duration(cmdutil.FlagRetryWait, 10*time.Second, "max backoff wait between retries")
	pf.String(cmdutil.FlagSelect, "", "keep only these comma-separated JSON paths (e.g. \"results[].symbol\")")
	pf.Int(cmdutil.FlagMax, -1, "truncate the result list to at most N items")
	pf.Bool(cmdutil.FlagCount, false, "output only the result count")
	pf.BoolP(cmdutil.FlagVerbose, "v", false, "log each request (method, URL, status, duration) to stderr")
	pf.Bool(cmdutil.FlagDryRun, false, "print the request that would be sent without calling the API")
	pf.Bool(cmdutil.FlagNoCache, false, "bypass the on-disk response cache")
	pf.Duration(cmdutil.FlagCacheTTL, 0, "uniform cache TTL (0 = per-endpoint defaults)")

	rootCmd.AddCommand(
		newAuthCmd(),
		newManifestCmd(),
		newCacheCmd(),
		idx.NewCmd(),
		sgx.NewCmd(),
		klse.NewCmd(),
		mining.NewCmd(),
	)
}
