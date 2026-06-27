package idx

import (
	"github.com/spf13/cobra"

	"github.com/emrsyah/sectors-cli/cmd/cmdutil"
	"github.com/emrsyah/sectors-cli/internal/sectors"
)

// newListCmd groups the helper-list endpoints that enumerate the slugs and
// dates used as inputs to other commands' filters.
func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Helper lists: valid sector/industry/tag slugs and segment coverage",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "industries",
			Short: "All subsector/industry slug pairs (inputs for --industry)",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
					r, err := c.IndustriesListWithResponse(cmd.Context())
					if err != nil {
						return 0, nil, err
					}
					return r.StatusCode(), r.Body, nil
				})
			},
		},
		&cobra.Command{
			Use:   "subindustries",
			Short: "All industry/sub-industry slug pairs (inputs for --sub-industry)",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
					r, err := c.SubindustriesListWithResponse(cmd.Context())
					if err != nil {
						return 0, nil, err
					}
					return r.StatusCode(), r.Body, nil
				})
			},
		},
		&cobra.Command{
			Use:   "subsectors",
			Short: "All sector/subsector slug pairs (inputs for --sector and --sub-sector)",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
					r, err := c.SubsectorsListWithResponse(cmd.Context())
					if err != nil {
						return 0, nil, err
					}
					return r.StatusCode(), r.Body, nil
				})
			},
		},
		&cobra.Command{
			Use:   "tags",
			Short: "All news/filing tag slugs (inputs for --tags)",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
					r, err := c.TagsListWithResponse(cmd.Context())
					if err != nil {
						return 0, nil, err
					}
					return r.StatusCode(), r.Body, nil
				})
			},
		},
		&cobra.Command{
			Use:   "segments-companies",
			Short: "Companies that have revenue-segment data, with available years",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
					r, err := c.CompaniesListCompaniesWithSegmentsListWithResponse(cmd.Context())
					if err != nil {
						return 0, nil, err
					}
					return r.StatusCode(), r.Body, nil
				})
			},
		},
	)
	return cmd
}
