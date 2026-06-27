// Package klse implements the `sectors klse` command group: Bursa Malaysia
// (KLSE) data — companies by sector, rankings, and company reports.
package klse

import (
	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
	"github.com/supertypeai/sectors-cli/internal/sectors"
)

// NewCmd builds the `klse` command tree.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "klse",
		Short: "Bursa Malaysia (KLSE) data",
	}
	cmd.AddCommand(
		newCompaniesCmd(),
		newTopCmd(),
		newReportCmd(),
		newSectorsCmd(),
	)
	return cmd
}

func newCompaniesCmd() *cobra.Command {
	var sector string
	cmd := &cobra.Command{
		Use:   "companies",
		Short: "List KLSE companies in a sector",
		Long: `Lists all KLSE-listed companies in a sector as symbol + company_name pairs.

--sector is required; valid slugs come from ` + "`sectors klse sectors`" + `.

Examples:
  sectors klse companies --sector financials
  sectors klse companies --sector healthcare`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			// sector is a required, non-optional query parameter.
			params := &sectors.KLSECompaniesListParams{Sector: sector}
			resp, err := client.KLSECompaniesListWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	cmd.Flags().StringVar(&sector, "sector", "", "sector slug (required)")
	_ = cmd.MarkFlagRequired("sector")
	return cmd
}

func newTopCmd() *cobra.Command {
	var sector, classifications string
	var nStock, minMcapMillion int
	cmd := &cobra.Command{
		Use:   "top",
		Short: "Top KLSE companies ranked by classification",
		Long: `Returns top KLSE-listed companies grouped per requested classification.

--classifications (comma-separated): dividend_yield, revenue, earnings,
market_cap, pe (default all).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.KLSECompaniesTopRetrieveParams{
				Sector:          cmdutil.OptStr(cmd, "sector", sector),
				NStock:          cmdutil.OptInt(cmd, "n-stock", nStock),
				Classifications: cmdutil.OptStr(cmd, "classifications", classifications),
				MinMcapMillion:  cmdutil.OptInt(cmd, "min-mcap-million", minMcapMillion),
			}
			resp, err := client.KLSECompaniesTopRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&sector, "sector", "", "sector slug filter (default all)")
	f.IntVar(&nStock, "n-stock", 5, "number of stocks per classification (max 10)")
	f.StringVar(&classifications, "classifications", "", "dividend_yield,revenue,earnings,market_cap,pe (default all)")
	f.IntVar(&minMcapMillion, "min-mcap-million", 1000, "minimum market cap in million MYR")
	return cmd
}

func newReportCmd() *cobra.Command {
	var sections string
	cmd := &cobra.Command{
		Use:   "report <symbol>",
		Short: "Full company report for a KLSE-listed symbol",
		Long: `Returns a full company report for a KLSE symbol (4-digit numeric code, e.g.
1155, 5225).

Narrow with --sections (comma-separated): overview, valuation, financials,
dividend.

Examples:
  sectors klse report 1155
  sectors klse report 5225 --sections overview,dividend`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.KLSECompanyReportRetrieve2Params{
				Sections: cmdutil.OptStr(cmd, "sections", sections),
			}
			resp, err := client.KLSECompanyReportRetrieve2WithResponse(cmd.Context(), args[0], params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	cmd.Flags().StringVar(&sections, "sections", "", "comma-separated sections to include (default all)")
	return cmd
}

func newSectorsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sectors",
		Short: "All KLSE sector slugs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
				r, err := c.KLSESectorsListWithResponse(cmd.Context())
				if err != nil {
					return 0, nil, err
				}
				return r.StatusCode(), r.Body, nil
			})
		},
	}
}
