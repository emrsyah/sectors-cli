// Package cmd implements the sectors CLI command tree.
package cmd

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/internal/config"
	"github.com/supertypeai/sectors-cli/internal/output"
	"github.com/supertypeai/sectors-cli/internal/sectors"
)

// errHandled signals that a command already rendered its own error (to stderr)
// and Execute should just exit non-zero without printing anything further.
var errHandled = errors.New("handled")

var (
	flagAPIKey  string
	flagBaseURL string
	flagOutput  string
	flagTimeout time.Duration

	// outFmt is the validated --output value, set in PersistentPreRunE.
	outFmt output.Format
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
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		f, err := output.ParseFormat(flagOutput)
		if err != nil {
			return err
		}
		outFmt = f
		return nil
	},
}

// Execute runs the root command and exits the process with an appropriate code.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if !errors.Is(err, errHandled) {
			// A framework/validation error cobra didn't print (SilenceErrors).
			output.EmitError(0, err.Error(), nil)
		}
		os.Exit(1)
	}
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&flagAPIKey, "api-key", "", "Sectors API key (overrides $SECTORS_API_KEY and config file)")
	pf.StringVar(&flagBaseURL, "base-url", "", "override the API base URL (default https://api.sectors.app)")
	pf.StringVarP(&flagOutput, "output", "o", "auto", "output format: auto, json, or pretty")
	pf.DurationVar(&flagTimeout, "timeout", 30*time.Second, "HTTP request timeout")
}

// newClient builds an authenticated API client from the resolved configuration.
func newClient() (*sectors.ClientWithResponses, error) {
	key, err := config.ResolveAPIKey(flagAPIKey)
	if err != nil {
		return nil, err
	}
	base, err := config.ResolveBaseURL(flagBaseURL)
	if err != nil {
		return nil, err
	}
	return sectors.New(base, key, &http.Client{Timeout: flagTimeout})
}

// fail renders a structured error to stderr and returns errHandled so Execute
// exits non-zero without re-printing.
func fail(status int, msg string, body []byte) error {
	output.EmitError(status, msg, body)
	return errHandled
}

// emit checks the HTTP status and either prints the body as JSON or renders an
// error. Use it for every command's response handling.
func emit(cmd *cobra.Command, statusCode int, body []byte) error {
	if statusCode < 200 || statusCode >= 300 {
		return fail(statusCode, "request failed", body)
	}
	if err := output.EmitJSON(cmd.OutOrStdout(), body, outFmt); err != nil {
		return fail(0, err.Error(), nil)
	}
	return nil
}
