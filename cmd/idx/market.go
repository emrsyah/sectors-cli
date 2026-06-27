package idx

import (
	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
	"github.com/supertypeai/sectors-cli/internal/sectors"
)

// --- transaction -----------------------------------------------------------

func newTransactionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transaction",
		Short: "IDX price, volume, and market-cap time series",
	}
	cmd.AddCommand(
		newTransactionDailyCmd(),
		newTransactionIDXTotalCmd(),
		newTransactionIndexDailyCmd(),
	)
	return cmd
}

func newTransactionDailyCmd() *cobra.Command {
	var start, end string
	cmd := &cobra.Command{
		Use:   "daily <symbol>",
		Short: "Daily close price, volume, and market cap for a symbol (≤90 days)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.DailyRetrieveParams{
				Start: cmdutil.OptStr(cmd, "start", start),
				End:   cmdutil.OptStr(cmd, "end", end),
			}
			resp, err := client.DailyRetrieveWithResponse(cmd.Context(), cmdutil.Sym(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&start, "start", "", "start date (YYYY-MM-DD)")
	f.StringVar(&end, "end", "", "end date (YYYY-MM-DD)")
	return cmd
}

func newTransactionIDXTotalCmd() *cobra.Command {
	var start, end string
	cmd := &cobra.Command{
		Use:   "idx-total",
		Short: "Historical total IDX market capitalization (≤90 days)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.IDXTotalRetrieveParams{
				Start: cmdutil.OptStr(cmd, "start", start),
				End:   cmdutil.OptStr(cmd, "end", end),
			}
			resp, err := client.IDXTotalRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&start, "start", "", "start date (YYYY-MM-DD, earliest 2021-01-01)")
	f.StringVar(&end, "end", "", "end date (YYYY-MM-DD)")
	return cmd
}

func newTransactionIndexDailyCmd() *cobra.Command {
	var start, end string
	cmd := &cobra.Command{
		Use:   "index-daily <index_code>",
		Short: "Daily closing price for an IDX index (≤90 days)",
		Long: `Returns daily closing price for an IDX index (e.g. lq45, ihsg, idx30) over a
date range of up to 90 days.

Examples:
  sectors idx transaction index-daily lq45
  sectors idx transaction index-daily ihsg --start 2024-01-01 --end 2024-03-31`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.IndexDailyRetrieveParams{
				Start: cmdutil.OptStr(cmd, "start", start),
				End:   cmdutil.OptStr(cmd, "end", end),
			}
			resp, err := client.IndexDailyRetrieveWithResponse(cmd.Context(), cmdutil.Slug(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&start, "start", "", "start date (YYYY-MM-DD)")
	f.StringVar(&end, "end", "", "end date (YYYY-MM-DD)")
	return cmd
}

// --- ranking ---------------------------------------------------------------

func newRankingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ranking",
		Short: "IDX most-traded stocks and top movers",
	}
	cmd.AddCommand(
		newRankingMostTradedCmd(),
		newRankingTopChangesCmd(),
	)
	return cmd
}

func newRankingMostTradedCmd() *cobra.Command {
	var subSector, start, end string
	var adjusted bool
	var nStock int
	cmd := &cobra.Command{
		Use:   "most-traded",
		Short: "Most traded IDX stocks by volume, keyed by date (≤90 days)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MostTradedRetrieveParams{
				SubSector: cmdutil.OptStr(cmd, "sub-sector", subSector),
				Start:     cmdutil.OptStr(cmd, "start", start),
				End:       cmdutil.OptStr(cmd, "end", end),
				Adjusted:  cmdutil.OptBool(cmd, "adjusted", adjusted),
				NStock:    cmdutil.OptInt(cmd, "n-stock", nStock),
			}
			resp, err := client.MostTradedRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&subSector, "sub-sector", "", "sub-sector slug filter")
	f.StringVar(&start, "start", "", "start date (YYYY-MM-DD)")
	f.StringVar(&end, "end", "", "end date (YYYY-MM-DD)")
	f.BoolVar(&adjusted, "adjusted", false, "use adjusted volume")
	f.IntVar(&nStock, "n-stock", 5, "number of stocks (max 10)")
	return cmd
}

func newRankingTopChangesCmd() *cobra.Command {
	var subSector, classifications, periods string
	var nStock, minMcapBillion int
	cmd := &cobra.Command{
		Use:   "top-changes",
		Short: "Top gainers and losers across multiple periods",
		Long: `Returns top gainers and losers across time periods.

--classifications: top_gainers, top_losers (comma-separated, default all)
--periods: 1d, 7d, 14d, 30d, 365d (comma-separated, default all)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.CompaniesTopChangesRetrieveParams{
				SubSector:       cmdutil.OptStr(cmd, "sub-sector", subSector),
				NStock:          cmdutil.OptInt(cmd, "n-stock", nStock),
				Classifications: cmdutil.OptStr(cmd, "classifications", classifications),
				Periods:         cmdutil.OptStr(cmd, "periods", periods),
				MinMcapBillion:  cmdutil.OptInt(cmd, "min-mcap-billion", minMcapBillion),
			}
			resp, err := client.CompaniesTopChangesRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&subSector, "sub-sector", "", "sub-sector slug filter")
	f.IntVar(&nStock, "n-stock", 5, "number of stocks per list (max 10)")
	f.StringVar(&classifications, "classifications", "", "top_gainers,top_losers (default all)")
	f.StringVar(&periods, "periods", "", "1d,7d,14d,30d,365d (default all)")
	f.IntVar(&minMcapBillion, "min-mcap-billion", 5000, "minimum market cap in billion IDR")
	return cmd
}
