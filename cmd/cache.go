package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
)

func newCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage the on-disk response cache",
		Long: `GET responses are cached on disk to cut latency, cost, and rate-limit
pressure for repetitive reads. Lifetimes vary by endpoint volatility (reference
lists ~24h, intraday data ~1m, reports ~5m). Bypass with --no-cache or override
with --cache-ttl on any command.`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "path",
			Short: "Print the cache directory",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				dir, err := cmdutil.CacheDir()
				if err != nil {
					return cmdutil.Fail(0, err.Error(), nil)
				}
				fmt.Fprintln(cmd.OutOrStdout(), dir)
				return nil
			},
		},
		&cobra.Command{
			Use:   "clear",
			Short: "Delete all cached responses",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				dir, err := cmdutil.CacheDir()
				if err != nil {
					return cmdutil.Fail(0, err.Error(), nil)
				}
				if err := os.RemoveAll(dir); err != nil {
					return cmdutil.Fail(0, err.Error(), nil)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Cleared cache at %s\n", dir)
				return nil
			},
		},
	)
	return cmd
}
