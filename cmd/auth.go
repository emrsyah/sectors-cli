package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
	"github.com/supertypeai/sectors-cli/internal/config"
)

func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Sectors API authentication",
	}

	var loginKey string
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Store your Sectors API key in the config file",
		Long: `Store your Sectors API key so future commands authenticate automatically.

The key is written to the per-user config file (see ` + "`sectors auth status`" + `)
with 0600 permissions. Provide it with --api-key:

  sectors auth login --api-key sk_live_xxx

You can also skip this entirely and export SECTORS_API_KEY instead.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if loginKey == "" {
				return cmdutil.Fail(0, "missing --api-key", nil)
			}
			f, err := config.Load()
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			f.APIKey = loginKey
			path, err := config.Save(f)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Saved API key to %s\n", path)
			return nil
		},
	}
	loginCmd.Flags().StringVar(&loginKey, "api-key", "", "the Sectors API key to store")

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show how authentication is currently resolved",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := config.Path()
			w := cmd.OutOrStdout()

			// --api-key here is the global persistent flag inherited from root.
			apiKeyFlag, _ := cmd.Flags().GetString(cmdutil.FlagAPIKey)

			var source, key string
			switch {
			case apiKeyFlag != "":
				source, key = "--api-key flag", apiKeyFlag
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

	authCmd.AddCommand(loginCmd, statusCmd)
	return authCmd
}

// mask shows only the last 4 characters of a secret.
func mask(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return "****" + s[len(s)-4:]
}
