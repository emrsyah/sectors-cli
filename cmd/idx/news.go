package idx

import (
	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
	"github.com/supertypeai/sectors-cli/internal/sectors"
)

func newNewsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "news",
		Short: "IDX news articles, insider filings, and suspensions",
	}
	cmd.AddCommand(
		newNewsListCmd(),
		newNewsFilingsCmd(),
		newNewsSuspensionsCmd(),
	)
	return cmd
}

func newNewsListCmd() *cobra.Command {
	var (
		sector, subSector, commodityType, start, end, tags, extension, keyword, symbols string
		limit, offset                                                                   int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Paginated IDX or mining news articles",
		Long: `Returns paginated news articles.

--extension chooses the data source (idx or mining); each source supports a
different set of filters. IDX-only filters: --sector, --sub-sector, --tags,
--symbols. Mining-only filter: --commodity-type.

Examples:
  sectors idx news list --keyword dividend --limit 10
  sectors idx news list --extension mining --commodity-type Coal
  sectors idx news list --symbols BBCA,BMRI --sub-sector banks`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.NewsRetrieveParams{
				Sector:        cmdutil.OptStr(cmd, "sector", sector),
				SubSector:     cmdutil.OptStr(cmd, "sub-sector", subSector),
				CommodityType: cmdutil.OptStr(cmd, "commodity-type", commodityType),
				Start:         cmdutil.OptStr(cmd, "start", start),
				End:           cmdutil.OptStr(cmd, "end", end),
				Limit:         cmdutil.OptInt(cmd, "limit", limit),
				Offset:        cmdutil.OptInt(cmd, "offset", offset),
				Tags:          cmdutil.OptStr(cmd, "tags", tags),
				Extension:     cmdutil.OptEnum[sectors.NewsRetrieveParamsExtension](cmd, "extension", extension),
				Keyword:       cmdutil.OptStr(cmd, "keyword", keyword),
				Symbols:       cmdutil.OptStr(cmd, "symbols", symbols),
			}
			resp, err := client.NewsRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&extension, "extension", "", "data source: idx|mining (default idx)")
	f.StringVar(&sector, "sector", "", "sector slugs, comma-separated (IDX only)")
	f.StringVar(&subSector, "sub-sector", "", "sub-sector slugs, comma-separated (IDX only)")
	f.StringVar(&commodityType, "commodity-type", "", "commodity type (mining only)")
	f.StringVar(&tags, "tags", "", "tag slugs, comma-separated (IDX only)")
	f.StringVar(&symbols, "symbols", "", "ticker symbols, comma-separated (IDX only)")
	f.StringVar(&keyword, "keyword", "", "substring match on the article title")
	f.StringVar(&start, "start", "", "start date (YYYY-MM-DD)")
	f.StringVar(&end, "end", "", "end date (YYYY-MM-DD)")
	f.IntVar(&limit, "limit", 20, "max results (max 30)")
	f.IntVar(&offset, "offset", 0, "pagination offset")
	return cmd
}

func newNewsFilingsCmd() *cobra.Command {
	var (
		symbol, sector, subSector, start, end, tags, transactionType, holderType string
		limit, offset                                                            int
	)
	cmd := &cobra.Command{
		Use:   "filings",
		Short: "IDX insider trading filings (buy/sell by insiders & major holders)",
		Long: `Returns IDX insider trading filings.

--transaction-type: buy, sell, others
--holder-type: corporate-investor, insider, institution`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.FilingsRetrieveParams{
				Symbol:          cmdutil.OptStr(cmd, "symbol", symbol),
				Sector:          cmdutil.OptStr(cmd, "sector", sector),
				SubSector:       cmdutil.OptStr(cmd, "sub-sector", subSector),
				Start:           cmdutil.OptStr(cmd, "start", start),
				End:             cmdutil.OptStr(cmd, "end", end),
				Limit:           cmdutil.OptInt(cmd, "limit", limit),
				Offset:          cmdutil.OptInt(cmd, "offset", offset),
				TransactionType: cmdutil.OptEnum[sectors.FilingsRetrieveParamsTransactionType](cmd, "transaction-type", transactionType),
				Tags:            cmdutil.OptStr(cmd, "tags", tags),
				HolderType:      cmdutil.OptEnum[sectors.FilingsRetrieveParamsHolderType](cmd, "holder-type", holderType),
			}
			resp, err := client.FilingsRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&symbol, "symbol", "", "ticker symbol filter")
	f.StringVar(&sector, "sector", "", "sector slug filter")
	f.StringVar(&subSector, "sub-sector", "", "sub-sector slug filter")
	f.StringVar(&transactionType, "transaction-type", "", "buy|sell|others")
	f.StringVar(&holderType, "holder-type", "", "corporate-investor|insider|institution")
	f.StringVar(&tags, "tags", "", "tag slugs, comma-separated")
	f.StringVar(&start, "start", "", "start date (YYYY-MM-DD)")
	f.StringVar(&end, "end", "", "end date (YYYY-MM-DD)")
	f.IntVar(&limit, "limit", 20, "max results (max 30)")
	f.IntVar(&offset, "offset", 0, "pagination offset")
	return cmd
}

func newNewsSuspensionsCmd() *cobra.Command {
	var symbol, start, end string
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "suspensions",
		Short: "Historical IDX stock suspensions with reasons",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.SuspensionsRetrieveParams{
				Symbol: cmdutil.OptStr(cmd, "symbol", symbol),
				Start:  cmdutil.OptStr(cmd, "start", start),
				End:    cmdutil.OptStr(cmd, "end", end),
				Limit:  cmdutil.OptInt(cmd, "limit", limit),
				Offset: cmdutil.OptInt(cmd, "offset", offset),
			}
			resp, err := client.SuspensionsRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&symbol, "symbol", "", "ticker symbol filter")
	f.StringVar(&start, "start", "", "start date (YYYY-MM-DD)")
	f.StringVar(&end, "end", "", "end date (YYYY-MM-DD)")
	f.IntVar(&limit, "limit", 20, "max results (max 30)")
	f.IntVar(&offset, "offset", 0, "pagination offset")
	return cmd
}
