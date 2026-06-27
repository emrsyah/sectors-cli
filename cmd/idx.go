package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/internal/sectors"
)

// idx is the Indonesia Stock Exchange command group.
var idxCmd = &cobra.Command{
	Use:   "idx",
	Short: "Indonesia Stock Exchange (IDX) data",
}

// idx company ...
var idxCompanyCmd = &cobra.Command{
	Use:   "company",
	Short: "Company-level IDX data (reports, financials, actions)",
}

var idxCompanyReportSections string

var idxCompanyReportCmd = &cobra.Command{
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
		client, err := newClient()
		if err != nil {
			return fail(0, err.Error(), nil)
		}

		params := &sectors.CompanyReportRetrieve2Params{}
		if idxCompanyReportSections != "" {
			params.Sections = &idxCompanyReportSections
		}

		symbol := strings.ToUpper(args[0])
		resp, err := client.CompanyReportRetrieve2WithResponse(cmd.Context(), symbol, params)
		if err != nil {
			return fail(0, err.Error(), nil)
		}
		return emit(cmd, resp.StatusCode(), resp.Body)
	},
}

func init() {
	idxCompanyReportCmd.Flags().StringVar(&idxCompanyReportSections, "sections", "",
		"comma-separated sections to include (default all)")

	idxCompanyCmd.AddCommand(idxCompanyReportCmd)
	idxCmd.AddCommand(idxCompanyCmd)
	rootCmd.AddCommand(idxCmd)
}
