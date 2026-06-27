package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCollectTools_CoversLeafCommands(t *testing.T) {
	tools := collectTools(rootCmd, nil)
	if len(tools) < 50 {
		t.Fatalf("expected the full command surface, got only %d tools", len(tools))
	}

	byName := map[string]tool{}
	for _, tl := range tools {
		byName[tl.Name] = tl
		if strings.Contains(tl.Name, "-") {
			t.Errorf("tool name %q should use underscores, not hyphens", tl.Name)
		}
	}

	// Auth/help/completion/manifest must not be exposed as tools.
	for _, skip := range []string{"auth_login", "auth_status", "help", "completion", "manifest"} {
		if _, ok := byName[skip]; ok {
			t.Errorf("%q should be excluded from the manifest", skip)
		}
	}

	// A representative command with a path arg + optional flag.
	rep, ok := byName["idx_company_report"]
	if !ok {
		t.Fatal("missing idx_company_report")
	}
	props := rep.Schema["properties"].(map[string]any)
	if _, ok := props["symbol"]; !ok {
		t.Error("idx_company_report missing 'symbol' property")
	}
	req, _ := rep.Schema["required"].([]string)
	if !contains(req, "symbol") {
		t.Errorf("idx_company_report should require 'symbol', got %v", req)
	}
}

func TestToolFromCmd_RequiredFlagsAndEnums(t *testing.T) {
	byName := map[string]tool{}
	for _, tl := range collectTools(rootCmd, nil) {
		byName[tl.Name] = tl
	}

	// Required flags should land in `required`.
	exp := byName["mining_commodities_exports"]
	req, _ := exp.Schema["required"].([]string)
	if !contains(req, "commodity_type") || !contains(req, "year") {
		t.Errorf("exports required = %v, want commodity_type & year", req)
	}

	// Enum flags ("a|b|c" in usage) should produce an enum.
	reg := byName["idx_brokers_registry"]
	cohort := reg.Schema["properties"].(map[string]any)["cohort"].(map[string]any)
	enum, ok := cohort["enum"].([]string)
	if !ok || !contains(enum, "institutional") {
		t.Errorf("cohort enum = %v, want institutional present", cohort["enum"])
	}
}

func TestRenderManifest_Formats(t *testing.T) {
	tools := []tool{{Name: "x_y", Description: "d", Schema: map[string]any{"type": "object"}}}

	anth, err := renderManifest(tools, "anthropic")
	if err != nil || !strings.Contains(anth, `"input_schema"`) {
		t.Errorf("anthropic format missing input_schema: %v\n%s", err, anth)
	}
	oai, err := renderManifest(tools, "openai")
	if err != nil || !strings.Contains(oai, `"type": "function"`) {
		t.Errorf("openai format missing function wrapper: %v\n%s", err, oai)
	}
	js, err := renderManifest(tools, "json")
	if err != nil || !strings.Contains(js, `"parameters"`) {
		t.Errorf("json format missing parameters: %v\n%s", err, js)
	}
	if _, err := renderManifest(tools, "bogus"); err == nil {
		t.Error("bogus format should error")
	}

	// Output must be valid JSON.
	var v any
	if err := json.Unmarshal([]byte(anth), &v); err != nil {
		t.Errorf("anthropic output not valid JSON: %v", err)
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
