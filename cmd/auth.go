package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/internal/config"
)

var authLoginKey string

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Sectors API authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store your Sectors API key in the config file",
	Long: `Store your Sectors API key so future commands authenticate automatically.

The key is written to the per-user config file (see ` + "`sectors auth status`" + `)
with 0600 permissions. Provide it with --api-key:

  sectors auth login --api-key sk_live_xxx

You can also skip this entirely and export SECTORS_API_KEY instead.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if authLoginKey == "" {
			return fail(0, "missing --api-key", nil)
		}
		f, err := config.Load()
		if err != nil {
			return fail(0, err.Error(), nil)
		}
		f.APIKey = authLoginKey
		path, err := config.Save(f)
		if err != nil {
			return fail(0, err.Error(), nil)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Saved API key to %s\n", path)
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show how authentication is currently resolved",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		path, _ := config.Path()
		w := cmd.OutOrStdout()

		var source, key string
		switch {
		case flagAPIKey != "":
			source, key = "--api-key flag", flagAPIKey
		case os.Getenv(config.EnvAPIKey) != "":
			source, key = "$"+config.EnvAPIKey, os.Getenv(config.EnvAPIKey)
		default:
			if f, err := config.Load(); err == nil && f.APIKey != "" {
				source, key = "config file ("+path+")", f.APIKey
			}
		}

		if key == "" {
			fmt.Fprintf(w, "Not authenticated.\nConfig file: %s\nSet --api-key, export %s, or run `sectors auth login`.\n", path, config.EnvAPIKey)
			return nil
		}
		fmt.Fprintf(w, "Authenticated via %s\nKey: %s\nConfig file: %s\n", source, mask(key), path)
		return nil
	},
}

// mask shows only the last 4 characters of a secret.
func mask(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return "****" + s[len(s)-4:]
}

func init() {
	authLoginCmd.Flags().StringVar(&authLoginKey, "api-key", "", "the Sectors API key to store")
	authCmd.AddCommand(authLoginCmd, authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
