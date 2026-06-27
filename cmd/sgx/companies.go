package sgx

import (
	"github.com/spf13/cobra"

	"github.com/emrsyah/sectors-cli/cmd/cmdutil"
	"github.com/emrsyah/sectors-cli/internal/sectors"
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
		Short: "Screen & rank SGX companies (SQL-like or natural language)",
		Long: `High-performance screener for SGX-listed companies.

Two ways to query:
  --q     natural-language query (e.g. "largest Singapore banks by market cap").
          When set, it OVERRIDES every other flag.
  --where SQL-like filter with =, !=, >, >=, <, <=, like, in, and/or.

Sort with --order-by (prefix a field with "-" for descending); paginate with
--limit (max 200) and --offset.

Examples:
  sectors sgx companies --q "top 5 REITs by dividend yield"
  sectors sgx companies --where "sub_sector='banks'" --order-by "-market_cap" --limit 10`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.SGXCompaniesRetrieveParams{
				Where:              cmdutil.OptStr(cmd, "where", where),
				Q:                  cmdutil.OptStr(cmd, "q", q),
				OrderBy:            cmdutil.OptStr(cmd, "order-by", orderBy),
				Desc:               cmdutil.OptBool(cmd, "desc", desc),
				Limit:              cmdutil.OptInt(cmd, "limit", limit),
				Offset:             cmdutil.OptInt(cmd, "offset", offset),
				IncludeQueryValues: cmdutil.OptBool(cmd, "include-query-values", queryValues),
			}
			resp, err := client.SGXCompaniesRetrieveWithResponse(cmd.Context(), params)
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

func newTopCmd() *cobra.Command {
	var sector, classifications string
	var nStock, minMcapMillion int
	cmd := &cobra.Command{
		Use:   "top",
		Short: "Top SGX companies ranked by classification",
		Long: `Returns top SGX-listed companies grouped per requested classification.

--classifications (comma-separated): dividend_yield, revenue, earnings,
market_cap, pe (default all).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.SGXCompaniesTopRetrieveParams{
				Sector:          cmdutil.OptStr(cmd, "sector", sector),
				NStock:          cmdutil.OptInt(cmd, "n-stock", nStock),
				Classifications: cmdutil.OptStr(cmd, "classifications", classifications),
				MinMcapMillion:  cmdutil.OptInt(cmd, "min-mcap-million", minMcapMillion),
			}
			resp, err := client.SGXCompaniesTopRetrieveWithResponse(cmd.Context(), params)
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
	f.IntVar(&minMcapMillion, "min-mcap-million", 1000, "minimum market cap in million SGD")
	return cmd
}

func newReportCmd() *cobra.Command {
	var sections string
	cmd := &cobra.Command{
		Use:   "report <symbol>",
		Short: "Full company report for an SGX-listed symbol",
		Long: `Returns a full company report for an SGX symbol (e.g. D05, U11, Z74).

Narrow with --sections (comma-separated): overview, valuation, financials,
dividend.

Examples:
  sectors sgx report D05
  sectors sgx report U11 --sections overview,dividend`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.SGXCompanyReportRetrieve2Params{
				Sections: cmdutil.OptStr(cmd, "sections", sections),
			}
			resp, err := client.SGXCompanyReportRetrieve2WithResponse(cmd.Context(), cmdutil.Sym(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	cmd.Flags().StringVar(&sections, "sections", "", "comma-separated sections to include (default all)")
	return cmd
}
