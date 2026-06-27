package idx

import (
	"github.com/spf13/cobra"

	"github.com/emrsyah/sectors-cli/cmd/cmdutil"
	"github.com/emrsyah/sectors-cli/internal/sectors"
)

func newBrokersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "brokers",
		Short: "IDX broker activity, summaries, and rankings",
	}
	cmd.AddCommand(
		newBrokersActivityCmd(),
		newBrokersActivityTopCmd(),
		newBrokersSummaryCmd(),
		newBrokersSummaryTopCmd(),
		newBrokersRegistryCmd(),
		newBrokersTopCmd(),
		newBrokersForeignFlowCmd(),
	)
	return cmd
}

func newBrokersActivityCmd() *cobra.Command {
	var symbol, start, end string
	cmd := &cobra.Command{
		Use:   "activity <broker_code>",
		Short: "All trading activity for one broker over a date range (≤14 days)",
		Long: `Lists every stock a broker touched per day with buy/sell/net values over a date
range (up to 14 days). Optionally filter to a single stock with --symbol.

Broker codes come from ` + "`sectors idx brokers registry`" + `.

Examples:
  sectors idx brokers activity MG
  sectors idx brokers activity AK --symbol BBCA --start 2024-01-01 --end 2024-01-14`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.BrokerActivityRetrieveParams{
				Symbol: cmdutil.OptStr(cmd, "symbol", symbol),
				Start:  cmdutil.OptStr(cmd, "start", start),
				End:    cmdutil.OptStr(cmd, "end", end),
			}
			resp, err := client.BrokerActivityRetrieveWithResponse(cmd.Context(), cmdutil.Code(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&symbol, "symbol", "", "filter to a single stock ticker")
	f.StringVar(&start, "start", "", "start date (YYYY-MM-DD)")
	f.StringVar(&end, "end", "", "end date (YYYY-MM-DD)")
	return cmd
}

func newBrokersActivityTopCmd() *cobra.Command {
	var start, end string
	var nBrokers int
	cmd := &cobra.Command{
		Use:   "activity-top <broker_code>",
		Short: "Stocks a broker has most accumulated/distributed over a range",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.BrokerActivityTopRetrieveParams{
				Start:    cmdutil.OptStr(cmd, "start", start),
				End:      cmdutil.OptStr(cmd, "end", end),
				NBrokers: cmdutil.OptInt(cmd, "n-brokers", nBrokers),
			}
			resp, err := client.BrokerActivityTopRetrieveWithResponse(cmd.Context(), cmdutil.Code(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&start, "start", "", "start date (YYYY-MM-DD)")
	f.StringVar(&end, "end", "", "end date (YYYY-MM-DD)")
	f.IntVar(&nBrokers, "n-brokers", 10, "number of stocks per list (max 100)")
	return cmd
}

func newBrokersSummaryCmd() *cobra.Command {
	var brokerCode, start, end string
	cmd := &cobra.Command{
		Use:   "summary <symbol>",
		Short: "Per-broker daily trading for one IDX ticker (≤14 days)",
		Long: `Lists every broker active on a ticker per day with buy/sell/net values, lots,
frequency, and weighted average price. Optionally filter to one broker with
--broker-code.

Examples:
  sectors idx brokers summary BBCA
  sectors idx brokers summary GOTO --broker-code MG --start 2024-01-01 --end 2024-01-14`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.BrokerSummaryRetrieveParams{
				BrokerCode: cmdutil.OptStr(cmd, "broker-code", brokerCode),
				Start:      cmdutil.OptStr(cmd, "start", start),
				End:        cmdutil.OptStr(cmd, "end", end),
			}
			resp, err := client.BrokerSummaryRetrieveWithResponse(cmd.Context(), cmdutil.Sym(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&brokerCode, "broker-code", "", "filter to a single broker")
	f.StringVar(&start, "start", "", "start date (YYYY-MM-DD)")
	f.StringVar(&end, "end", "", "end date (YYYY-MM-DD)")
	return cmd
}

func newBrokersSummaryTopCmd() *cobra.Command {
	var start, end, cohort, origin string
	var nBrokers int
	cmd := &cobra.Command{
		Use:   "summary-top <symbol>",
		Short: "Top net buyers & sellers of an IDX ticker over a range",
		Long: `Ranks brokers by net buy and net sell value for a single ticker.

--cohort: all, institutional, mixed, retail, unknown
--origin: all, domestic, foreign`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.BrokerSummaryTopRetrieveParams{
				Start:    cmdutil.OptStr(cmd, "start", start),
				End:      cmdutil.OptStr(cmd, "end", end),
				Cohort:   cmdutil.OptEnum[sectors.BrokerSummaryTopRetrieveParamsCohort](cmd, "cohort", cohort),
				NBrokers: cmdutil.OptInt(cmd, "n-brokers", nBrokers),
				Origin:   cmdutil.OptEnum[sectors.BrokerSummaryTopRetrieveParamsOrigin](cmd, "origin", origin),
			}
			resp, err := client.BrokerSummaryTopRetrieveWithResponse(cmd.Context(), cmdutil.Sym(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&start, "start", "", "start date (YYYY-MM-DD)")
	f.StringVar(&end, "end", "", "end date (YYYY-MM-DD)")
	f.StringVar(&cohort, "cohort", "", "all|institutional|mixed|retail|unknown")
	f.IntVar(&nBrokers, "n-brokers", 10, "number of brokers per list (max 100)")
	f.StringVar(&origin, "origin", "", "all|domestic|foreign")
	return cmd
}

func newBrokersRegistryCmd() *cobra.Command {
	var cohort, origin string
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Authoritative registry of IDX exchange-member brokers",
		Long: `Curated registry of IDX brokers (code, name, origin, cohort, license type).
Use this as the source of valid broker codes for the other broker commands.

--cohort: institutional, mixed, retail, unknown
--origin: domestic, foreign`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.BrokersRetrieveParams{
				Cohort: cmdutil.OptEnum[sectors.BrokersRetrieveParamsCohort](cmd, "cohort", cohort),
				Origin: cmdutil.OptEnum[sectors.BrokersRetrieveParamsOrigin](cmd, "origin", origin),
			}
			resp, err := client.BrokersRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&cohort, "cohort", "", "institutional|mixed|retail|unknown")
	f.StringVar(&origin, "origin", "", "domestic|foreign")
	return cmd
}

func newBrokersTopCmd() *cobra.Command {
	var cohort, date, metric, origin string
	var nBrokers int
	cmd := &cobra.Command{
		Use:   "top",
		Short: "Brokers ranked by trade value or net flow for a single date",
		Long: `Ranks brokers by gross trade value (default) or absolute net flow on a date.

--metric: gross, net
--cohort: all, institutional, mixed, retail, unknown
--origin: all, domestic, foreign`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.BrokersTopRetrieveParams{
				Cohort:   cmdutil.OptEnum[sectors.BrokersTopRetrieveParamsCohort](cmd, "cohort", cohort),
				Date:     cmdutil.OptStr(cmd, "date", date),
				Metric:   cmdutil.OptEnum[sectors.BrokersTopRetrieveParamsMetric](cmd, "metric", metric),
				NBrokers: cmdutil.OptInt(cmd, "n-brokers", nBrokers),
				Origin:   cmdutil.OptEnum[sectors.BrokersTopRetrieveParamsOrigin](cmd, "origin", origin),
			}
			resp, err := client.BrokersTopRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&cohort, "cohort", "", "all|institutional|mixed|retail|unknown")
	f.StringVar(&date, "date", "", "target date (YYYY-MM-DD, default latest)")
	f.StringVar(&metric, "metric", "", "gross|net")
	f.IntVar(&nBrokers, "n-brokers", 0, "number of brokers (default all, max 200)")
	f.StringVar(&origin, "origin", "", "all|domestic|foreign")
	return cmd
}

func newBrokersForeignFlowCmd() *cobra.Command {
	var start, end string
	cmd := &cobra.Command{
		Use:   "foreign-flow <symbol>",
		Short: "Daily net foreign-broker inflow for an IDX ticker (≤90 days)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.ForeignFlowRetrieveParams{
				Start: cmdutil.OptStr(cmd, "start", start),
				End:   cmdutil.OptStr(cmd, "end", end),
			}
			resp, err := client.ForeignFlowRetrieveWithResponse(cmd.Context(), cmdutil.Sym(args[0]), params)
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
