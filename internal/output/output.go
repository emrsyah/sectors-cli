// Package output renders API responses for the sectors CLI.
//
// The CLI is agent-facing first: the default output is the API's raw JSON,
// emitted losslessly. When stdout is an interactive terminal we pretty-print
// for humans; when piped (the common case for an agent) we stay compact and
// machine-parseable. Errors are emitted as JSON on stderr with a non-zero exit
// code so callers can branch on them programmatically.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// Format selects how successful responses are rendered.
type Format string

const (
	// FormatAuto picks Pretty on a TTY, JSON otherwise.
	FormatAuto Format = "auto"
	// FormatJSON is compact single-line JSON.
	FormatJSON Format = "json"
	// FormatPretty is indented JSON.
	FormatPretty Format = "pretty"
	// FormatTable renders tabular responses as a table, falling back to Pretty
	// for data that isn't a list of records.
	FormatTable Format = "table"
)

// ParseFormat validates a --output value.
func ParseFormat(s string) (Format, error) {
	switch Format(s) {
	case FormatAuto, FormatJSON, FormatPretty, FormatTable:
		return Format(s), nil
	default:
		return "", fmt.Errorf("invalid --output %q: want auto, json, pretty, or table", s)
	}
}

// IsTTY reports whether the given file is an interactive terminal.
func IsTTY(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

// resolve turns FormatAuto into a concrete format based on whether w is a TTY.
func resolve(format Format, w io.Writer) Format {
	if format != FormatAuto {
		return format
	}
	if f, ok := w.(*os.File); ok && IsTTY(f) {
		return FormatPretty
	}
	return FormatJSON
}

// EmitJSON writes a raw JSON document (as returned by the API) to w in the
// requested format. body must already be valid JSON.
func EmitJSON(w io.Writer, body []byte, format Format) error {
	switch resolve(format, w) {
	case FormatTable:
		// Render a table when the payload is a list of records; otherwise fall
		// back to indented JSON so nested documents stay readable.
		if s, ok := renderTable(body); ok {
			_, err := io.WriteString(w, s)
			return err
		}
		return EmitJSON(w, body, FormatPretty)
	case FormatPretty:
		var buf bytes.Buffer
		if err := json.Indent(&buf, body, "", "  "); err != nil {
			// Not JSON (shouldn't happen for 2xx) — pass through verbatim.
			_, err := w.Write(body)
			return err
		}
		buf.WriteByte('\n')
		_, err := w.Write(buf.Bytes())
		return err
	default: // FormatJSON: compact onto a single line.
		var buf bytes.Buffer
		if err := json.Compact(&buf, body); err != nil {
			_, err := w.Write(body)
			return err
		}
		buf.WriteByte('\n')
		_, err := w.Write(buf.Bytes())
		return err
	}
}

// EmitError writes a structured error to stderr as JSON. status is the HTTP
// status code when the error came from the API (0 for client-side errors), and
// body is the API's raw error payload if any.
func EmitError(status int, msg string, body []byte) {
	payload := map[string]any{"error": msg}
	if status != 0 {
		payload["status"] = status
	}
	if len(bytes.TrimSpace(body)) > 0 {
		// Embed the API's error JSON if it parses, else as a string.
		var raw json.RawMessage
		if json.Unmarshal(body, &raw) == nil {
			payload["response"] = raw
		} else {
			payload["response"] = string(body)
		}
	}
	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	_ = enc.Encode(payload)
}
