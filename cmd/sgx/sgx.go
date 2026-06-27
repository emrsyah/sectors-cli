// Package sgx implements the `sectors sgx` command group: Singapore Exchange
// (SGX) data — screener, reports, news, filings, and transactions.
package sgx

import "github.com/spf13/cobra"

// NewCmd builds the `sgx` command tree.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sgx",
		Short: "Singapore Exchange (SGX) data",
	}
	cmd.AddCommand(
		newCompaniesCmd(),
		newTopCmd(),
		newReportCmd(),
		newSectorsCmd(),
		newTagsCmd(),
		newNewsCmd(),
		newFilingsCmd(),
		newBuybacksCmd(),
		newShortSellCmd(),
		newDailyCmd(),
	)
	return cmd
}
