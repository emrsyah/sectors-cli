package sgx

import (
	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
	"github.com/supertypeai/sectors-cli/internal/sectors"
)

func newDailyCmd() *cobra.Command {
	var start, end string
	cmd := &cobra.Command{
		Use:   "daily <symbol>",
		Short: "Daily close price and volume for an SGX symbol (≤90 days)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.SGXDailyRetrieveParams{
				Start: cmdutil.OptStr(cmd, "start", start),
				End:   cmdutil.OptStr(cmd, "end", end),
			}
			resp, err := client.SGXDailyRetrieveWithResponse(cmd.Context(), cmdutil.Sym(args[0]), params)
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

func newBuybacksCmd() *cobra.Command {
	var symbol, start, end string
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "buybacks",
		Short: "SGX share buyback records",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.SGXBuybacksRetrieveParams{
				Symbol: cmdutil.OptStr(cmd, "symbol", symbol),
				Start:  cmdutil.OptStr(cmd, "start", start),
				End:    cmdutil.OptStr(cmd, "end", end),
				Limit:  cmdutil.OptInt(cmd, "limit", limit),
				Offset: cmdutil.OptInt(cmd, "offset", offset),
			}
			resp, err := client.SGXBuybacksRetrieveWithResponse(cmd.Context(), params)
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

func newShortSellCmd() *cobra.Command {
	var symbol, start, end string
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "short-sell",
		Short: "SGX short sell data",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.SGXShortSellRetrieveParams{
				Symbol: cmdutil.OptStr(cmd, "symbol", symbol),
				Start:  cmdutil.OptStr(cmd, "start", start),
				End:    cmdutil.OptStr(cmd, "end", end),
				Limit:  cmdutil.OptInt(cmd, "limit", limit),
				Offset: cmdutil.OptInt(cmd, "offset", offset),
			}
			resp, err := client.SGXShortSellRetrieveWithResponse(cmd.Context(), params)
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
