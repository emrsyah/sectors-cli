package idx

import (
	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
	"github.com/supertypeai/sectors-cli/internal/sectors"
)

func newCompaniesCmd() *cobra.Command {
	var (
		where       string
		q           string
		orderBy     string
		desc        bool
		limit       int
		offset      int
		queryValues bool
	)
	cmd := &cobra.Command{
		Use:   "companies",
		Short: "Screen & rank IDX companies (SQL-like or natural language)",
		Long: `High-performance screener for IDX-listed companies.

Two ways to query:
  --q     natural-language query (e.g. "top 5 growing banks in 2024"). When set,
          it OVERRIDES every other flag.
  --where SQL-like filter with =, !=, >, >=, <, <=, like, in, and/or, plus
          bracketed time-series fields like revenue[2024].

Sort with --order-by (prefix a field with "-" for descending), and paginate with
--limit (max 200) and --offset.

Examples:
  sectors idx companies --q "top growing banking companies by revenue in 2024"
  sectors idx companies --where "sub_sector='banks'" --order-by "-revenue[2024]" --limit 10
  sectors idx companies --where "indices in ['lq45','idxbumn20']"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.CompaniesRetrieveParams{
				Where:              cmdutil.OptStr(cmd, "where", where),
				Q:                  cmdutil.OptStr(cmd, "q", q),
				OrderBy:            cmdutil.OptStr(cmd, "order-by", orderBy),
				Desc:               cmdutil.OptBool(cmd, "desc", desc),
				Limit:              cmdutil.OptInt(cmd, "limit", limit),
				Offset:             cmdutil.OptInt(cmd, "offset", offset),
				IncludeQueryValues: cmdutil.OptBool(cmd, "include-query-values", queryValues),
			}
			resp, err := client.CompaniesRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&where, "where", "", "SQL-like filter expression")
	f.StringVar(&q, "q", "", "natural-language query (overrides all other flags)")
	f.StringVar(&orderBy, "order-by", "", "sort field; prefix with - for descending")
	f.BoolVar(&desc, "desc", false, "sort descending")
	f.IntVar(&limit, "limit", 50, "max results (max 200)")
	f.IntVar(&offset, "offset", 0, "pagination offset")
	f.BoolVar(&queryValues, "include-query-values", false, "include a query_values object in the response")
	return cmd
}

func newFreeFloatCmd() *cobra.Command {
	var (
		sector      string
		subSector   string
		industry    string
		subIndustry string
	)
	cmd := &cobra.Command{
		Use:   "free-float",
		Short: "Free-float percentage for IDX companies, ordered descending",
		Long: `Returns the free-float percentage for IDX-listed companies, optionally filtered
by one level of the sector taxonomy (kebab-case slugs from ` + "`idx list`" + `).

Examples:
  sectors idx free-float --sub-sector banks
  sectors idx free-float --sector infrastructures`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.FreeFloatRetrieveParams{
				Sector:      cmdutil.OptStr(cmd, "sector", sector),
				SubSector:   cmdutil.OptStr(cmd, "sub-sector", subSector),
				Industry:    cmdutil.OptStr(cmd, "industry", industry),
				SubIndustry: cmdutil.OptStr(cmd, "sub-industry", subIndustry),
			}
			resp, err := client.FreeFloatRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&sector, "sector", "", "sector slug")
	f.StringVar(&subSector, "sub-sector", "", "sub-sector slug")
	f.StringVar(&industry, "industry", "", "industry slug")
	f.StringVar(&subIndustry, "sub-industry", "", "sub-industry slug")
	return cmd
}

func newSubsectorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subsector",
		Short: "Subsector-level IDX data",
	}
	cmd.AddCommand(newSubsectorReportCmd())
	return cmd
}

func newSubsectorReportCmd() *cobra.Command {
	var sections string
	cmd := &cobra.Command{
		Use:   "report <sub_sector>",
		Short: "Comprehensive report for an IDX subsector",
		Long: `Returns a comprehensive report for an IDX subsector (kebab-case slug from
` + "`idx list subsectors`" + `).

Narrow with --sections (comma-separated): statistics, market_cap, stability,
valuation, growth, companies.

Examples:
  sectors idx subsector report banks
  sectors idx subsector report utilities --sections statistics,valuation`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.SubsectorReportRetrieve2Params{
				Sections: cmdutil.OptStr(cmd, "sections", sections),
			}
			resp, err := client.SubsectorReportRetrieve2WithResponse(cmd.Context(), cmdutil.Slug(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	cmd.Flags().StringVar(&sections, "sections", "", "comma-separated sections to include (default all)")
	return cmd
}
