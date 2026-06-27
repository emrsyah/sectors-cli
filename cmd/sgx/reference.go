package sgx

import (
	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
	"github.com/supertypeai/sectors-cli/internal/sectors"
)

func newSectorsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sectors",
		Short: "All SGX sector slugs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
				r, err := c.SGXSectorsListWithResponse(cmd.Context())
				if err != nil {
					return 0, nil, err
				}
				return r.StatusCode(), r.Body, nil
			})
		},
	}
}

func newTagsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tags",
		Short: "All SGX news tag slugs (inputs for `news --tags`)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
				r, err := c.SGXTagsRetrieveWithResponse(cmd.Context())
				if err != nil {
					return 0, nil, err
				}
				return r.StatusCode(), r.Body, nil
			})
		},
	}
}
