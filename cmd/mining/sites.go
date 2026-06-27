package mining

import (
	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
	"github.com/supertypeai/sectors-cli/internal/sectors"
)

// --- sites -----------------------------------------------------------------

func newSitesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sites",
		Short: "Mining sites with production, reserves, and location",
	}
	cmd.AddCommand(newSitesListCmd(), newSitesGetCmd())
	return cmd
}

func newSitesListCmd() *cobra.Command {
	var province, commodityType, company, orderBy string
	var year, limit, offset int
	var minProduction float64
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mining sites with filtering and sorting",
		Long: `Lists mining sites with optional filters and sorting.

--order-by: production_volume, strip_ratio, year (prefix with - for descending;
default -year).

Examples:
  sectors mining sites list --commodity-type Coal --order-by -production_volume
  sectors mining sites list --province "Kalimantan Timur" --min-production 1000000`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MiningSitesRetrieveParams{
				Province:      cmdutil.OptStr(cmd, "province", province),
				CommodityType: cmdutil.OptStr(cmd, "commodity-type", commodityType),
				Company:       cmdutil.OptStr(cmd, "company", company),
				Year:          cmdutil.OptInt(cmd, "year", year),
				OrderBy:       cmdutil.OptEnum[sectors.MiningSitesRetrieveParamsOrderBy](cmd, "order-by", orderBy),
				MinProduction: cmdutil.OptFloat(cmd, "min-production", minProduction),
				Limit:         cmdutil.OptInt(cmd, "limit", limit),
				Offset:        cmdutil.OptInt(cmd, "offset", offset),
			}
			resp, err := client.MiningSitesRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&province, "province", "", "province, exact match")
	f.StringVar(&commodityType, "commodity-type", "", "commodity type filter")
	f.StringVar(&company, "company", "", "company slug filter")
	f.IntVar(&year, "year", 0, "production year filter")
	f.StringVar(&orderBy, "order-by", "", "production_volume|strip_ratio|year (prefix - for desc)")
	f.Float64Var(&minProduction, "min-production", 0, "minimum production volume")
	f.IntVar(&limit, "limit", 20, "max results (max 30)")
	f.IntVar(&offset, "offset", 0, "pagination offset")
	return cmd
}

func newSitesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <slug>",
		Short: "Full detail for a single mining site (incl. lat/lng)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
				r, err := c.MiningSitesRetrieve2WithResponse(cmd.Context(), cmdutil.Slug(args[0]))
				if err != nil {
					return 0, nil, err
				}
				return r.StatusCode(), r.Body, nil
			})
		},
	}
}

// --- production ------------------------------------------------------------

func newProductionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "production",
		Short: "National commodity production data",
	}
	cmd.AddCommand(newProductionTotalCmd())
	return cmd
}

func newProductionTotalCmd() *cobra.Command {
	var commodityType string
	cmd := &cobra.Command{
		Use:   "total",
		Short: "Total national production for a commodity, by year (with YoY)",
		Long: `Returns total national production across all years for a commodity, including
year-over-year change.

--commodity-type is required.

Example:
  sectors mining production total --commodity-type Coal`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MiningTotalProductionRetrieveParams{CommodityType: commodityType}
			resp, err := client.MiningTotalProductionRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	cmd.Flags().StringVar(&commodityType, "commodity-type", "", "commodity type (required)")
	_ = cmd.MarkFlagRequired("commodity-type")
	return cmd
}

// --- reserves --------------------------------------------------------------

func newReservesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reserves",
		Short: "Resources & reserves data by province and commodity",
	}
	cmd.AddCommand(newReservesIndexCmd(), newReservesGetCmd())
	return cmd
}

func newReservesIndexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "Which provinces/years/commodities have reserves data",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
				r, err := c.MiningResourcesReservesRetrieveWithResponse(cmd.Context())
				if err != nil {
					return 0, nil, err
				}
				return r.StatusCode(), r.Body, nil
			})
		},
	}
}

func newReservesGetCmd() *cobra.Command {
	var commodityType string
	var year int
	cmd := &cobra.Command{
		Use:   "get <province>",
		Short: "Resources & reserves detail for a province",
		Long: `Returns reserves data for a province, nested by year then commodity.

Use ` + "`sectors mining reserves index`" + ` to find provinces with data.

Examples:
  sectors mining reserves get "Kalimantan Timur"
  sectors mining reserves get "Sulawesi Tenggara" --commodity-type Nickel --year 2023`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MiningResourcesReservesRetrieve2Params{
				CommodityType: cmdutil.OptStr(cmd, "commodity-type", commodityType),
				Year:          cmdutil.OptInt(cmd, "year", year),
			}
			resp, err := client.MiningResourcesReservesRetrieve2WithResponse(cmd.Context(), args[0], params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&commodityType, "commodity-type", "", "commodity type filter")
	f.IntVar(&year, "year", 0, "year filter")
	return cmd
}
