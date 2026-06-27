// Package config resolves runtime configuration for the sectors CLI.
//
// Precedence (highest first):
//  1. an explicit value passed on the command line (--api-key / --base-url)
//  2. environment variables (SECTORS_API_KEY / SECTORS_BASE_URL)
//  3. the config file at <user-config-dir>/sectors/config.yaml
//
// The config file is intentionally tiny YAML so `sectors auth login` can write
// it and humans can hand-edit it.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	EnvAPIKey  = "SECTORS_API_KEY"
	EnvBaseURL = "SECTORS_BASE_URL"
)

// File is the on-disk config shape.
type File struct {
	APIKey  string `yaml:"api_key,omitempty"`
	BaseURL string `yaml:"base_url,omitempty"`
}

// Dir returns the directory holding the sectors config file.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "sectors"), nil
}

// Path returns the full path to config.yaml.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads the config file. A missing file is not an error — it returns an
// empty File so callers can still fall back to flags and env vars.
func Load() (File, error) {
	var f File
	path, err := Path()
	if err != nil {
		return f, err
	}
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return f, nil
	}
	if err != nil {
		return f, err
	}
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return f, fmt.Errorf("parsing %s: %w", path, err)
	}
	return f, nil
}

// Save writes the config file, creating the directory if needed (0700, since it
// holds a secret).
func Save(f File) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "config.yaml")
	raw, err := yaml.Marshal(f)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// ResolveAPIKey applies the precedence rules for the API key. flagVal is the
// value of --api-key ("" if unset).
func ResolveAPIKey(flagVal string) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}
	if env := os.Getenv(EnvAPIKey); env != "" {
		return env, nil
	}
	f, err := Load()
	if err != nil {
		return "", err
	}
	if f.APIKey == "" {
		return "", fmt.Errorf("no API key found: set --api-key, the %s env var, or run `sectors auth login`", EnvAPIKey)
	}
	return f.APIKey, nil
}

// ResolveBaseURL applies precedence for the base URL, returning "" if none is
// configured (the client then uses its built-in default).
func ResolveBaseURL(flagVal string) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}
	if env := os.Getenv(EnvBaseURL); env != "" {
		return env, nil
	}
	f, err := Load()
	if err != nil {
		return "", err
	}
	return f.BaseURL, nil
}
