package mining

import (
	"github.com/spf13/cobra"

	"github.com/emrsyah/sectors-cli/cmd/cmdutil"
	"github.com/emrsyah/sectors-cli/internal/sectors"
)

func newCommoditiesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commodities",
		Short: "Commodity prices, exports, and global trade data",
	}
	cmd.AddCommand(
		newCommoditiesListCmd(),
		newCommoditiesPriceCmd(),
		newCommoditiesExportsCmd(),
		newCommoditiesGlobalCmd(),
		newCommoditiesSalesDestinationCmd(),
	)
	return cmd
}

func newCommoditiesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "All commodities in the price database, with coverage metadata",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
				r, err := c.MiningCommoditiesRetrieveWithResponse(cmd.Context())
				if err != nil {
					return 0, nil, err
				}
				return r.StatusCode(), r.Body, nil
			})
		},
	}
}

func newCommoditiesPriceCmd() *cobra.Command {
	var startYear, endYear int
	cmd := &cobra.Command{
		Use:   "price <commodity_name>",
		Short: "Monthly price history for a commodity (max 3-year range)",
		Long: `Returns historical monthly price data for a commodity (e.g. Gold, Coal).

Use ` + "`sectors mining commodities list`" + ` to discover valid commodity names.

Examples:
  sectors mining commodities price Gold
  sectors mining commodities price Coal --start-year 2022 --end-year 2024`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MiningCommoditiesPriceRetrieveParams{
				StartYear: cmdutil.OptInt(cmd, "start-year", startYear),
				EndYear:   cmdutil.OptInt(cmd, "end-year", endYear),
			}
			resp, err := client.MiningCommoditiesPriceRetrieveWithResponse(cmd.Context(), args[0], params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.IntVar(&startYear, "start-year", 0, "start year (default current year - 2)")
	f.IntVar(&endYear, "end-year", 0, "end year (default current year)")
	return cmd
}

func newCommoditiesExportsCmd() *cobra.Command {
	var commodityType string
	var year, limit int
	cmd := &cobra.Command{
		Use:   "exports",
		Short: "Top export destinations for a commodity in a year",
		Long: `Ranks countries by total export value for a commodity and year.

--commodity-type and --year are required.

Example:
  sectors mining commodities exports --commodity-type Coal --year 2023`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MiningExportsRetrieveParams{
				CommodityType: commodityType,
				Year:          year,
				Limit:         cmdutil.OptInt(cmd, "limit", limit),
			}
			resp, err := client.MiningExportsRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&commodityType, "commodity-type", "", "commodity type (required)")
	f.IntVar(&year, "year", 0, "year (required)")
	f.IntVar(&limit, "limit", 20, "max results (max 30)")
	_ = cmd.MarkFlagRequired("commodity-type")
	_ = cmd.MarkFlagRequired("year")
	return cmd
}

func newCommoditiesGlobalCmd() *cobra.Command {
	var commodityType, country string
	var limit int
	cmd := &cobra.Command{
		Use:   "global",
		Short: "Global commodity data (production, reserves, trade)",
		Long: `Returns global commodity data by commodity and/or country.

At least one of --commodity-type or --country is required.

Examples:
  sectors mining commodities global --commodity-type Nickel
  sectors mining commodities global --country Indonesia`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MiningGlobalCommodityRetrieveParams{
				CommodityType: cmdutil.OptStr(cmd, "commodity-type", commodityType),
				Country:       cmdutil.OptStr(cmd, "country", country),
				Limit:         cmdutil.OptInt(cmd, "limit", limit),
			}
			resp, err := client.MiningGlobalCommodityRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&commodityType, "commodity-type", "", "commodity type (required if --country absent)")
	f.StringVar(&country, "country", "", "country, exact match (required if --commodity-type absent)")
	f.IntVar(&limit, "limit", 20, "max results (max 30)")
	cmd.MarkFlagsOneRequired("commodity-type", "country")
	return cmd
}

func newCommoditiesSalesDestinationCmd() *cobra.Command {
	var year int
	cmd := &cobra.Command{
		Use:   "sales-destination <slug>",
		Short: "Sales-destination breakdown for a mining company by year",
		Long:  "Revenue and volume distribution by destination country for a company. Defaults to the latest year.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MiningSalesDestinationRetrieveParams{
				Year: cmdutil.OptInt(cmd, "year", year),
			}
			resp, err := client.MiningSalesDestinationRetrieveWithResponse(cmd.Context(), cmdutil.Slug(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	cmd.Flags().IntVar(&year, "year", 0, "year (default latest)")
	return cmd
}
