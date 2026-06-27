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
	"github.com/supertypeai/sectors-cli/internal/output"
)

var rootCmd = &cobra.Command{
	Use:   "sectors",
	Short: "CLI for the Sectors Financial API (IDX, SGX, KLSE, mining)",
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
		if !errors.Is(err, cmdutil.ErrHandled) {
			// A framework/validation error cobra didn't print (SilenceErrors).
			output.EmitError(0, err.Error(), nil)
		}
		os.Exit(1)
	}
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.String(cmdutil.FlagAPIKey, "", "Sectors API key (overrides $SECTORS_API_KEY and config file)")
	pf.String(cmdutil.FlagBaseURL, "", "override the API base URL (default https://api.sectors.app)")
	pf.StringP(cmdutil.FlagOutput, "o", "auto", "output format: auto, json, or pretty")
	pf.Duration(cmdutil.FlagTimeout, 30*time.Second, "HTTP request timeout")

	rootCmd.AddCommand(
		newAuthCmd(),
		idx.NewCmd(),
	)
}
