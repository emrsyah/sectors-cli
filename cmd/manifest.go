package cmd

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
)

// tool is the neutral, format-independent description of one CLI command,
// derived entirely from the Cobra tree so it never drifts from the real flags.
type tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Schema      map[string]any `json:"-"`
}

// globalFlags are inherited persistent flags that are not per-command inputs and
// so are excluded from generated tool schemas.
var globalFlags = map[string]bool{
	cmdutil.FlagAPIKey: true, cmdutil.FlagBaseURL: true, cmdutil.FlagOutput: true,
	cmdutil.FlagTimeout: true, cmdutil.FlagRetries: true, cmdutil.FlagRetryWait: true,
	cmdutil.FlagSelect: true, cmdutil.FlagMax: true, cmdutil.FlagCount: true,
	cmdutil.FlagVerbose: true, cmdutil.FlagDryRun: true, cmdutil.FlagNoCache: true,
	cmdutil.FlagCacheTTL: true, "help": true, "version": true,
}

// skipTrees are top-level command groups not exposed as agent tools.
var skipTrees = map[string]bool{
	"help": true, "completion": true, "manifest": true, "auth": true, "cache": true,
}

func newManifestCmd() *cobra.Command {
	var format, filter string
	cmd := &cobra.Command{
		Use:   "manifest",
		Short: "Print machine-readable tool schemas for every command",
		Long: `Generates tool/function-calling definitions for the entire command tree,
derived from the commands' own flags and help — so an agent host can load the
CLI as a callable toolset without hand-maintaining schemas.

Formats:
  json       neutral: [{name, description, parameters}]
  anthropic  [{name, description, input_schema}]  (tool use)
  openai     [{type:"function", function:{name, description, parameters}}]

Use --filter to export a subset (glob on the tool name), e.g. --filter "idx_*".`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			tools := collectTools(rootCmd, nil)

			if filter != "" {
				kept := tools[:0]
				for _, t := range tools {
					if ok, _ := path.Match(filter, t.Name); ok {
						kept = append(kept, t)
					}
				}
				tools = kept
			}
			sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })

			out, err := renderManifest(tools, format)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), out)
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "json", "output format: json, anthropic, or openai")
	cmd.Flags().StringVar(&filter, "filter", "", "glob to filter tools by name (e.g. \"idx_*\")")
	return cmd
}

// collectTools walks the command tree and returns one tool per runnable leaf.
func collectTools(c *cobra.Command, prefix []string) []tool {
	var tools []tool
	for _, sub := range c.Commands() {
		if len(prefix) == 0 && skipTrees[sub.Name()] {
			continue
		}
		if sub.Hidden {
			continue
		}
		p := append(append([]string{}, prefix...), sub.Name())
		if sub.Runnable() {
			tools = append(tools, toolFromCmd(sub, p))
		}
		tools = append(tools, collectTools(sub, p)...)
	}
	return tools
}

func toolFromCmd(c *cobra.Command, p []string) tool {
	props := map[string]any{}
	var required []string

	// Positional args declared in Use, e.g. "report <symbol>" / "get [slug]".
	for _, a := range parseArgs(c.Use) {
		props[a.name] = map[string]any{"type": "string", "description": a.name}
		if a.required {
			required = append(required, a.name)
		}
	}

	// Command-specific flags (skip inherited globals).
	c.Flags().VisitAll(func(f *pflag.Flag) {
		if globalFlags[f.Name] {
			return
		}
		name := strings.ReplaceAll(f.Name, "-", "_")
		prop := map[string]any{
			"type":        jsonType(f.Value.Type()),
			"description": f.Usage,
		}
		if enum := parseEnum(f.Usage); len(enum) > 0 {
			prop["enum"] = enum
		}
		props[name] = prop
		if isRequired(f) {
			required = append(required, name)
		}
	})

	schema := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		sort.Strings(required)
		schema["required"] = required
	}

	return tool{
		Name:        strings.ReplaceAll(strings.Join(p, "_"), "-", "_"),
		Description: c.Short,
		Schema:      schema,
	}
}

type argSpec struct {
	name     string
	required bool
}

// parseArgs extracts <required> and [optional] positional tokens from a Use line.
func parseArgs(use string) []argSpec {
	var out []argSpec
	for _, tok := range strings.Fields(use)[1:] { // [0] is the command name
		switch {
		case strings.HasPrefix(tok, "<") && strings.HasSuffix(tok, ">"):
			out = append(out, argSpec{name: clean(tok[1 : len(tok)-1]), required: true})
		case strings.HasPrefix(tok, "[") && strings.HasSuffix(tok, "]"):
			out = append(out, argSpec{name: clean(tok[1 : len(tok)-1]), required: false})
		}
	}
	return out
}

func clean(s string) string { return strings.ReplaceAll(s, "-", "_") }

func jsonType(pflagType string) string {
	switch pflagType {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	default:
		return "string"
	}
}

// parseEnum pulls "a|b|c" choice lists out of a flag's usage string.
func parseEnum(usage string) []string {
	for _, field := range strings.Fields(usage) {
		if strings.Contains(field, "|") {
			parts := strings.Split(strings.Trim(field, ".,"), "|")
			if len(parts) >= 2 {
				return parts
			}
		}
	}
	return nil
}

func isRequired(f *pflag.Flag) bool {
	if f.Annotations == nil {
		return false
	}
	_, ok := f.Annotations[cobra.BashCompOneRequiredFlag]
	return ok
}

func renderManifest(tools []tool, format string) (string, error) {
	var payload any
	switch format {
	case "json", "":
		arr := make([]map[string]any, len(tools))
		for i, t := range tools {
			arr[i] = map[string]any{"name": t.Name, "description": t.Description, "parameters": t.Schema}
		}
		payload = arr
	case "anthropic":
		arr := make([]map[string]any, len(tools))
		for i, t := range tools {
			arr[i] = map[string]any{"name": t.Name, "description": t.Description, "input_schema": t.Schema}
		}
		payload = arr
	case "openai":
		arr := make([]map[string]any, len(tools))
		for i, t := range tools {
			arr[i] = map[string]any{
				"type":     "function",
				"function": map[string]any{"name": t.Name, "description": t.Description, "parameters": t.Schema},
			}
		}
		payload = arr
	default:
		return "", fmt.Errorf("invalid --format %q: want json, anthropic, or openai", format)
	}

	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
