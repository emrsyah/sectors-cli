package sgx

import (
	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
	"github.com/supertypeai/sectors-cli/internal/sectors"
)

func newNewsCmd() *cobra.Command {
	var sector, subSector, start, end, tags, symbols string
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "news",
		Short: "Paginated SGX news articles",
		Long: `Returns paginated SGX news articles, filterable by sector, sub-sector, tags,
symbols, and date range.

Examples:
  sectors sgx news --symbols D05,U11 --limit 10
  sectors sgx news --sub-sector banks --start 2024-01-01`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.SGXNewsRetrieveParams{
				Sector:    cmdutil.OptStr(cmd, "sector", sector),
				SubSector: cmdutil.OptStr(cmd, "sub-sector", subSector),
				Start:     cmdutil.OptStr(cmd, "start", start),
				End:       cmdutil.OptStr(cmd, "end", end),
				Limit:     cmdutil.OptInt(cmd, "limit", limit),
				Offset:    cmdutil.OptInt(cmd, "offset", offset),
				Tags:      cmdutil.OptStr(cmd, "tags", tags),
				Symbols:   cmdutil.OptStr(cmd, "symbols", symbols),
			}
			resp, err := client.SGXNewsRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&sector, "sector", "", "sector slug filter")
	f.StringVar(&subSector, "sub-sector", "", "sub-sector slug filter")
	f.StringVar(&tags, "tags", "", "tag slugs, comma-separated")
	f.StringVar(&symbols, "symbols", "", "SGX symbols, comma-separated (e.g. D05,U11)")
	f.StringVar(&start, "start", "", "start date (YYYY-MM-DD)")
	f.StringVar(&end, "end", "", "end date (YYYY-MM-DD)")
	f.IntVar(&limit, "limit", 20, "max results (max 30)")
	f.IntVar(&offset, "offset", 0, "pagination offset")
	return cmd
}

func newFilingsCmd() *cobra.Command {
	var symbol, start, end, transactionType, holderType string
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "filings",
		Short: "SGX insider trading filings",
		Long: `Returns SGX insider trading filings (buy/sell by insiders and major holders).

--transaction-type: award, buy, others, sell, transfer
--holder-type: insider, institution`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.SGXFilingsRetrieveParams{
				Symbol:          cmdutil.OptStr(cmd, "symbol", symbol),
				Start:           cmdutil.OptStr(cmd, "start", start),
				End:             cmdutil.OptStr(cmd, "end", end),
				Limit:           cmdutil.OptInt(cmd, "limit", limit),
				Offset:          cmdutil.OptInt(cmd, "offset", offset),
				TransactionType: cmdutil.OptEnum[sectors.SGXFilingsRetrieveParamsTransactionType](cmd, "transaction-type", transactionType),
				HolderType:      cmdutil.OptEnum[sectors.SGXFilingsRetrieveParamsHolderType](cmd, "holder-type", holderType),
			}
			resp, err := client.SGXFilingsRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&symbol, "symbol", "", "ticker symbol filter")
	f.StringVar(&transactionType, "transaction-type", "", "award|buy|others|sell|transfer")
	f.StringVar(&holderType, "holder-type", "", "insider|institution")
	f.StringVar(&start, "start", "", "start date (YYYY-MM-DD)")
	f.StringVar(&end, "end", "", "end date (YYYY-MM-DD)")
	f.IntVar(&limit, "limit", 20, "max results (max 30)")
	f.IntVar(&offset, "offset", 0, "pagination offset")
	return cmd
}
