// Command fixspec sanitizes the upstream Sectors OpenAPI spec so oapi-codegen
// can consume it, without mutating the pristine source file.
//
// The upstream spec (sectors-schema.json) contains a few malformed paths — e.g.
// `/v2/company/report/` — that declare a *required path parameter* which does
// not appear in the URL template. oapi-codegen rejects these. They are also
// redundant: a correct `/v2/company/report/{symbol}/` variant always exists.
//
// This tool drops any operation whose declared `in: path` parameter is missing
// from its path template, drops paths left with no operations, and writes the
// result to a separate file. Re-run via `go generate ./...`.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
)

var tmplParam = regexp.MustCompile(`\{(\w+)\}`)

func main() {
	in := flag.String("in", "", "path to upstream OpenAPI spec (JSON)")
	out := flag.String("out", "", "path to write the sanitized spec (JSON)")
	flag.Parse()
	if *in == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: fixspec -in <spec.json> -out <fixed.json>")
		os.Exit(2)
	}

	raw, err := os.ReadFile(*in)
	must(err)

	var doc map[string]any
	must(json.Unmarshal(raw, &doc))

	paths, _ := doc["paths"].(map[string]any)
	httpMethods := map[string]bool{
		"get": true, "post": true, "put": true, "delete": true,
		"patch": true, "head": true, "options": true, "trace": true,
	}

	var dropped []string
	for path, pv := range paths {
		ops, _ := pv.(map[string]any)
		declared := map[string]bool{}
		for _, m := range tmplParam.FindAllStringSubmatch(path, -1) {
			declared[m[1]] = true
		}
		for method, ov := range ops {
			if !httpMethods[method] {
				continue
			}
			op, _ := ov.(map[string]any)
			params, _ := op["parameters"].([]any)
			bad := false
			for _, p := range params {
				pm, _ := p.(map[string]any)
				if pm["in"] == "path" {
					if name, _ := pm["name"].(string); !declared[name] {
						bad = true
						break
					}
				}
			}
			if bad {
				delete(ops, method)
				dropped = append(dropped, method+" "+path)
			}
		}
		// Remove the path entirely if no HTTP operations remain.
		hasOp := false
		for k := range ops {
			if httpMethods[k] {
				hasOp = true
				break
			}
		}
		if !hasOp {
			delete(paths, path)
		}
	}

	// Empty every response body schema so oapi-codegen generates `interface{}`
	// for the typed response fields. The upstream spec's response schemas are
	// frequently inaccurate (arrays typed as single objects, string-vs-number,
	// string-vs-map, offset-less timestamps), which makes the generated response
	// PARSER reject otherwise-valid payloads. The CLI emits the raw response body
	// and never reads the typed fields, so an untyped body is strictly safer:
	// json.Unmarshal into interface{} accepts any valid JSON.
	emptied := emptyResponseSchemas(paths)

	// Also drop `format: date-time` / `date` from any remaining (request-side)
	// string schemas, so nothing generates time.Time / openapi_types.Date.
	stripped := stripDateFormats(doc)

	pretty, err := json.MarshalIndent(doc, "", "  ")
	must(err)
	must(os.WriteFile(*out, pretty, 0o644))

	sort.Strings(dropped)
	fmt.Fprintf(os.Stderr, "fixspec: wrote %s (dropped %d malformed operation(s), emptied %d response schema(s), stripped %d date format(s))\n", *out, len(dropped), emptied, stripped)
	for _, d := range dropped {
		fmt.Fprintln(os.Stderr, "  - "+d)
	}
}

// emptyResponseSchemas replaces every operation's response body schema with an
// empty schema ({}), so the generated client uses interface{} for response
// bodies and its parser never rejects a valid payload. Returns how many it
// replaced. Request parameter schemas are untouched.
func emptyResponseSchemas(paths map[string]any) int {
	count := 0
	for _, pv := range paths {
		ops, _ := pv.(map[string]any)
		for _, ov := range ops {
			op, _ := ov.(map[string]any)
			if op == nil {
				continue
			}
			responses, _ := op["responses"].(map[string]any)
			for _, rv := range responses {
				resp, _ := rv.(map[string]any)
				content, _ := resp["content"].(map[string]any)
				for _, cv := range content {
					mt, _ := cv.(map[string]any)
					if mt == nil {
						continue
					}
					if _, ok := mt["schema"]; ok {
						mt["schema"] = map[string]any{}
						count++
					}
				}
			}
		}
	}
	return count
}

// stripDateFormats recursively removes `format: date-time` and `format: date`
// from string schemas anywhere in the document, returning how many it removed.
func stripDateFormats(node any) int {
	count := 0
	switch n := node.(type) {
	case map[string]any:
		if f, ok := n["format"].(string); ok && (f == "date-time" || f == "date") {
			// Only strip from string-typed (or untyped) schemas.
			if t, ok := n["type"].(string); !ok || t == "string" {
				delete(n, "format")
				count++
			}
		}
		for _, v := range n {
			count += stripDateFormats(v)
		}
	case []any:
		for _, v := range n {
			count += stripDateFormats(v)
		}
	}
	return count
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "fixspec:", err)
		os.Exit(1)
	}
}
