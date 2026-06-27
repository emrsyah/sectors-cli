// Package cmdutil holds helpers shared across every command package: building
// an authenticated client, rendering responses/errors, and turning optional
// flags into the pointer-typed query parameters the generated client expects.
//
// Command packages depend on this; this package depends on no command package,
// so there are no import cycles.
package cmdutil

import (
	"errors"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/internal/config"
	"github.com/supertypeai/sectors-cli/internal/output"
	"github.com/supertypeai/sectors-cli/internal/sectors"
)

// ErrHandled signals that a command already rendered its own error to stderr,
// so the top-level Execute should exit non-zero without printing anything more.
var ErrHandled = errors.New("handled")

// Global persistent flag names, defined once on the root command and read here
// from any (possibly deeply nested) subcommand via inherited flags.
const (
	FlagAPIKey  = "api-key"
	FlagBaseURL = "base-url"
	FlagOutput  = "output"
	FlagTimeout = "timeout"
)

// NewClient builds an authenticated API client from the resolved configuration
// (flag → env → config file precedence for the API key and base URL).
func NewClient(cmd *cobra.Command) (*sectors.ClientWithResponses, error) {
	apiKeyFlag, _ := cmd.Flags().GetString(FlagAPIKey)
	baseURLFlag, _ := cmd.Flags().GetString(FlagBaseURL)
	timeout, _ := cmd.Flags().GetDuration(FlagTimeout)

	key, err := config.ResolveAPIKey(apiKeyFlag)
	if err != nil {
		return nil, err
	}
	base, err := config.ResolveBaseURL(baseURLFlag)
	if err != nil {
		return nil, err
	}
	return sectors.New(base, key, &http.Client{Timeout: timeout})
}

// Format returns the validated --output format for this command.
func Format(cmd *cobra.Command) output.Format {
	s, _ := cmd.Flags().GetString(FlagOutput)
	f, err := output.ParseFormat(s)
	if err != nil {
		return output.FormatAuto
	}
	return f
}

// Fail renders a structured error to stderr and returns ErrHandled.
func Fail(status int, msg string, body []byte) error {
	output.EmitError(status, msg, body)
	return ErrHandled
}

// Do builds a client, runs fn, and emits the result. fn performs one API call
// and returns its HTTP status code and raw body. It keeps no-parameter commands
// down to a single expression.
func Do(cmd *cobra.Command, fn func(*sectors.ClientWithResponses) (int, []byte, error)) error {
	client, err := NewClient(cmd)
	if err != nil {
		return Fail(0, err.Error(), nil)
	}
	status, body, err := fn(client)
	if err != nil {
		return Fail(0, err.Error(), nil)
	}
	return Emit(cmd, status, body)
}

// Emit checks the HTTP status and either prints the body as JSON or renders an
// error. Use it for every command's response handling.
func Emit(cmd *cobra.Command, statusCode int, body []byte) error {
	if statusCode < 200 || statusCode >= 300 {
		return Fail(statusCode, "request failed", body)
	}
	if err := output.EmitJSON(cmd.OutOrStdout(), body, Format(cmd)); err != nil {
		return Fail(0, err.Error(), nil)
	}
	return nil
}

// Sym normalizes a ticker path argument (e.g. bbca -> BBCA).
func Sym(s string) string { return strings.ToUpper(s) }

// Code normalizes a broker-code path argument (e.g. mg -> MG).
func Code(s string) string { return strings.ToUpper(s) }

// Slug normalizes a kebab-case slug path argument (e.g. Banks -> banks).
func Slug(s string) string { return strings.ToLower(s) }

// --- optional-flag → pointer helpers --------------------------------------
//
// The API distinguishes "omitted" from "set to a zero value", and the generated
// client models optional query params as pointers. These return a pointer only
// when the user actually set the flag.

func OptStr(cmd *cobra.Command, name, val string) *string {
	if cmd.Flags().Changed(name) {
		return &val
	}
	return nil
}

func OptInt(cmd *cobra.Command, name string, val int) *int {
	if cmd.Flags().Changed(name) {
		return &val
	}
	return nil
}

func OptBool(cmd *cobra.Command, name string, val bool) *bool {
	if cmd.Flags().Changed(name) {
		return &val
	}
	return nil
}

func OptFloat(cmd *cobra.Command, name string, val float64) *float64 {
	if cmd.Flags().Changed(name) {
		return &val
	}
	return nil
}

// OptEnum is like OptStr but for the generated string-enum param types. The
// value is validated server-side, so an invalid one surfaces as a 400.
func OptEnum[T ~string](cmd *cobra.Command, name, val string) *T {
	if cmd.Flags().Changed(name) {
		v := T(val)
		return &v
	}
	return nil
}
