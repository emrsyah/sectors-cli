package mining

import (
	"github.com/spf13/cobra"

	"github.com/emrsyah/sectors-cli/cmd/cmdutil"
	"github.com/emrsyah/sectors-cli/internal/sectors"
)

func newCompaniesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "companies",
		Short: "Indonesian mining companies (search, detail, financials, ownership)",
	}
	cmd.AddCommand(
		newCompaniesListCmd(),
		newCompaniesGetCmd(),
		newCompaniesFinancialsCmd(),
		newCompaniesOwnershipCmd(),
		newCompaniesPerformanceCmd(),
	)
	return cmd
}

func newCompaniesListCmd() *cobra.Command {
	var commodityType, keyword, companyType string
	var limit, offset int
	var hasFinancials bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Search mining companies by name, symbol, slug, or operation",
		Long: `Searches Indonesian mining companies with optional filters.

--company-type examples: Holding, Subsidiary.
Use --has-financials to keep only companies that have financial records.

Examples:
  sectors mining companies list --keyword adaro
  sectors mining companies list --commodity-type Coal --has-financials`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MiningCompaniesRetrieveParams{
				CommodityType: cmdutil.OptStr(cmd, "commodity-type", commodityType),
				Limit:         cmdutil.OptInt(cmd, "limit", limit),
				Offset:        cmdutil.OptInt(cmd, "offset", offset),
				Keyword:       cmdutil.OptStr(cmd, "keyword", keyword),
				CompanyType:   cmdutil.OptStr(cmd, "company-type", companyType),
				HasFinancials: cmdutil.OptBool(cmd, "has-financials", hasFinancials),
			}
			resp, err := client.MiningCompaniesRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&keyword, "keyword", "", "search name/symbol/slug/operation")
	f.StringVar(&commodityType, "commodity-type", "", "commodity type filter")
	f.StringVar(&companyType, "company-type", "", "company type (e.g. Holding, Subsidiary)")
	f.BoolVar(&hasFinancials, "has-financials", false, "only companies with financial records")
	f.IntVar(&limit, "limit", 20, "max results")
	f.IntVar(&offset, "offset", 0, "pagination offset")
	return cmd
}

func newCompaniesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <slug>",
		Short: "Full operational detail for a single mining company",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
				r, err := c.MiningCompaniesRetrieve2WithResponse(cmd.Context(), cmdutil.Slug(args[0]))
				if err != nil {
					return 0, nil, err
				}
				return r.StatusCode(), r.Body, nil
			})
		},
	}
}

func newCompaniesFinancialsCmd() *cobra.Command {
	var year int
	cmd := &cobra.Command{
		Use:   "financials <slug>",
		Short: "Annual financials for a mining company (USD millions)",
		Long:  "Returns annual financial records (assets/revenue/profit). Defaults to the latest available year.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MiningCompaniesFinancialsRetrieveParams{
				Year: cmdutil.OptInt(cmd, "year", year),
			}
			resp, err := client.MiningCompaniesFinancialsRetrieveWithResponse(cmd.Context(), cmdutil.Slug(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	cmd.Flags().IntVar(&year, "year", 0, "financial year (default latest)")
	return cmd
}

func newCompaniesOwnershipCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ownership <slug>",
		Short: "Corporate ownership tree (parents and subsidiaries)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
				r, err := c.MiningCompaniesOwnershipRetrieveWithResponse(cmd.Context(), cmdutil.Slug(args[0]))
				if err != nil {
					return 0, nil, err
				}
				return r.StatusCode(), r.Body, nil
			})
		},
	}
}

func newCompaniesPerformanceCmd() *cobra.Command {
	var commodityType string
	var year int
	cmd := &cobra.Command{
		Use:   "performance <slug>",
		Short: "Production, sales, strip ratio and reserves for a year",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MiningCompaniesPerformanceRetrieveParams{
				CommodityType: cmdutil.OptStr(cmd, "commodity-type", commodityType),
				Year:          cmdutil.OptInt(cmd, "year", year),
			}
			resp, err := client.MiningCompaniesPerformanceRetrieveWithResponse(cmd.Context(), cmdutil.Slug(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&commodityType, "commodity-type", "", "commodity type filter")
	f.IntVar(&year, "year", 0, "year (default latest)")
	return cmd
}
