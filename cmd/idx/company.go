package idx

import (
	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
	"github.com/supertypeai/sectors-cli/internal/sectors"
)

func newCompanyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "company",
		Short: "Company-level IDX data (reports, financials, actions)",
	}
	cmd.AddCommand(
		newCompanyReportCmd(),
		newCompanySegmentsCmd(),
		newCompanyFinancialsCmd(),
		newCompanyCorpActionsCmd(),
		newCompanyShareholdersCmd(),
		newCompanyIPOCmd(),
		newCompanyQuarterlyDatesCmd(),
	)
	return cmd
}

func newCompanyReportCmd() *cobra.Command {
	var sections string
	cmd := &cobra.Command{
		Use:   "report <symbol>",
		Short: "Comprehensive company report for an IDX symbol",
		Long: `Returns a comprehensive company report organized into sections.

By default all sections are included. Narrow the response with --sections
(comma-separated): overview, valuation, future, peers, financials, dividend,
management, ownership.

Examples:
  sectors idx company report BBCA
  sectors idx company report BREN --sections overview,valuation`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.CompanyReportRetrieve2Params{
				Sections: cmdutil.OptStr(cmd, "sections", sections),
			}
			resp, err := client.CompanyReportRetrieve2WithResponse(cmd.Context(), cmdutil.Sym(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	cmd.Flags().StringVar(&sections, "sections", "", "comma-separated sections to include (default all)")
	return cmd
}

func newCompanySegmentsCmd() *cobra.Command {
	var year int
	cmd := &cobra.Command{
		Use:   "segments <symbol>",
		Short: "Revenue & cost segment breakdown (Sankey-ready)",
		Long: `Returns a revenue and cost segment breakdown for a company and financial year.

Not all companies have segment data — use ` + "`sectors idx list segments-companies`" + `
to check availability. Defaults to the latest available year.

Examples:
  sectors idx company segments TLKM
  sectors idx company segments BUMI --financial-year 2023`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.CompanyGetSegmentsRetrieveParams{
				FinancialYear: cmdutil.OptInt(cmd, "financial-year", year),
			}
			resp, err := client.CompanyGetSegmentsRetrieveWithResponse(cmd.Context(), cmdutil.Sym(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	cmd.Flags().IntVar(&year, "financial-year", 0, "financial year (default latest)")
	return cmd
}

func newCompanyFinancialsCmd() *cobra.Command {
	var (
		reportDate string
		approx     bool
		nQuarters  int
	)
	cmd := &cobra.Command{
		Use:   "financials <symbol>",
		Short: "Quarterly financial data for an IDX symbol",
		Long: `Returns quarterly financial data. Fields vary by sector — banks and insurers
include extra metrics like net_interest_income, gross_loan, and total_deposit.

Use --report-date (YYYY-MM-DD; see ` + "`idx company quarterly-dates`" + `) to target a
specific quarter, or --n-quarters to fetch the most recent N quarters.

Examples:
  sectors idx company financials BBCA --n-quarters 4
  sectors idx company financials BMRI --report-date 2024-03-31`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.FinancialsQuarterlyRetrieveParams{
				ReportDate: cmdutil.OptStr(cmd, "report-date", reportDate),
				Approx:     cmdutil.OptBool(cmd, "approx", approx),
				NQuarters:  cmdutil.OptInt(cmd, "n-quarters", nQuarters),
			}
			resp, err := client.FinancialsQuarterlyRetrieveWithResponse(cmd.Context(), cmdutil.Sym(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&reportDate, "report-date", "", "specific quarter report date (YYYY-MM-DD)")
	f.BoolVar(&approx, "approx", true, "approximate quarter matching")
	f.IntVar(&nQuarters, "n-quarters", 0, "most recent N quarters")
	return cmd
}

func newCompanyCorpActionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "corporate-actions <symbol>",
		Short: "Dividends, splits, rights issues and other corporate actions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			resp, err := client.CompanyCorporateActionsRetrieveWithResponse(cmd.Context(), cmdutil.Sym(args[0]))
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
}

func newCompanyShareholdersCmd() *cobra.Command {
	var year int
	cmd := &cobra.Command{
		Use:   "shareholders <symbol>",
		Short: "Monthly shareholders-composition snapshots",
		Long: `Returns monthly shareholders-composition snapshots for a symbol and year.

Data is available from 2021; earlier years return an empty data array and future
years are rejected. Defaults to the current year.

Examples:
  sectors idx company shareholders BBCA
  sectors idx company shareholders BMRI --year 2023`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.CompanyShareholdersCompositionRetrieveParams{
				Year: cmdutil.OptInt(cmd, "year", year),
			}
			resp, err := client.CompanyShareholdersCompositionRetrieveWithResponse(cmd.Context(), cmdutil.Sym(args[0]), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	cmd.Flags().IntVar(&year, "year", 0, "calendar year (default current)")
	return cmd
}

func newCompanyIPOCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ipo-performance <symbol>",
		Short: "Price change since listing across 7/30/90/365-day windows",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			resp, err := client.ListingPerformanceRetrieveWithResponse(cmd.Context(), cmdutil.Sym(args[0]))
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
}

func newCompanyQuarterlyDatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "quarterly-dates <symbol>",
		Short: "Available quarterly report dates (inputs for `financials --report-date`)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			resp, err := client.CompanyGetQuarterlyFinancialDatesRetrieveWithResponse(cmd.Context(), cmdutil.Sym(args[0]))
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
}
