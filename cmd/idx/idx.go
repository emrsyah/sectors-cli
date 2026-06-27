// Package idx implements the `sectors idx` command group: Indonesia Stock
// Exchange data (companies, reports, brokers, transactions, rankings, news).
package idx

import "github.com/spf13/cobra"

// NewCmd builds the `idx` command tree.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "idx",
		Short: "Indonesia Stock Exchange (IDX) data",
	}
	cmd.AddCommand(
		newCompanyCmd(),
		newCompaniesCmd(),
		newFreeFloatCmd(),
		newSubsectorCmd(),
		newBrokersCmd(),
		newTransactionCmd(),
		newRankingCmd(),
		newNewsCmd(),
		newListCmd(),
	)
	return cmd
}
