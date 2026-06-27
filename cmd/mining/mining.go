// Package mining implements the `sectors mining` command group: the Indonesian
// mining-sector extension — companies, commodities & trade, production & sites,
// and licenses, auctions & contracts.
package mining

import "github.com/spf13/cobra"

// NewCmd builds the `mining` command tree.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mining",
		Short: "Indonesian mining-sector data (companies, commodities, sites, licenses)",
	}
	cmd.AddCommand(
		newCompaniesCmd(),
		newCommoditiesCmd(),
		newSitesCmd(),
		newProductionCmd(),
		newReservesCmd(),
		newLicensesCmd(),
		newAuctionsCmd(),
		newContractsCmd(),
	)
	return cmd
}
