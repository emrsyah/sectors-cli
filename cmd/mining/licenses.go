package mining

import (
	"github.com/spf13/cobra"

	"github.com/supertypeai/sectors-cli/cmd/cmdutil"
	"github.com/supertypeai/sectors-cli/internal/sectors"
)

// --- licenses --------------------------------------------------------------

func newLicensesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "licenses",
		Short: "Mining licenses (IUP/IUPK) from the ESDM Minerba portal",
	}
	cmd.AddCommand(newLicensesListCmd())
	return cmd
}

func newLicensesListCmd() *cobra.Command {
	var province, commodityType, company, orderBy, licenseType, activity string
	var limit, offset int
	var expiringSoon, cnc bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mining licenses with filters for status, commodity, location",
		Long: `Lists mining licenses (IUP/IUPK).

--order-by: commodity_type, license_effective_date, license_expiry_date,
licensed_area_ha (prefix with - for descending; default license_expiry_date).
--license-type examples: IUP, IUPK. --activity examples: Eksplorasi,
"Operasi Produksi". Use --expiring-soon for licenses expiring within 365 days.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MiningLicensesRetrieveParams{
				Province:      cmdutil.OptStr(cmd, "province", province),
				CommodityType: cmdutil.OptStr(cmd, "commodity-type", commodityType),
				Company:       cmdutil.OptStr(cmd, "company", company),
				OrderBy:       cmdutil.OptEnum[sectors.MiningLicensesRetrieveParamsOrderBy](cmd, "order-by", orderBy),
				Limit:         cmdutil.OptInt(cmd, "limit", limit),
				Offset:        cmdutil.OptInt(cmd, "offset", offset),
				ExpiringSoon:  cmdutil.OptBool(cmd, "expiring-soon", expiringSoon),
				LicenseType:   cmdutil.OptStr(cmd, "license-type", licenseType),
				Activity:      cmdutil.OptStr(cmd, "activity", activity),
				Cnc:           cmdutil.OptBool(cmd, "cnc", cnc),
			}
			resp, err := client.MiningLicensesRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&province, "province", "", "province, exact match")
	f.StringVar(&commodityType, "commodity-type", "", "commodity type filter")
	f.StringVar(&company, "company", "", "company slug filter")
	f.StringVar(&orderBy, "order-by", "", "commodity_type|license_effective_date|license_expiry_date|licensed_area_ha")
	f.StringVar(&licenseType, "license-type", "", "license type (e.g. IUP, IUPK)")
	f.StringVar(&activity, "activity", "", "activity (e.g. Eksplorasi, Operasi Produksi)")
	f.BoolVar(&expiringSoon, "expiring-soon", false, "only licenses expiring within 365 days")
	f.BoolVar(&cnc, "cnc", false, "filter by clean-and-clear status")
	f.IntVar(&limit, "limit", 20, "max results (max 30)")
	f.IntVar(&offset, "offset", 0, "pagination offset")
	return cmd
}

// --- auctions --------------------------------------------------------------

func newAuctionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auctions",
		Short: "Mining license auctions from the ESDM Minerba portal",
	}
	cmd.AddCommand(newAuctionsListCmd(), newAuctionsGetCmd())
	return cmd
}

func newAuctionsListCmd() *cobra.Command {
	var province, commodityType, orderBy, areaType, status, participant string
	var limit, offset, minParticipants int
	var qualified bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mining license auctions (phases/participants omitted)",
		Long: `Lists mining license auctions. Use ` + "`auctions get <wiup_code>`" + ` for the full
record with phases and participants.

--order-by: commodity_type, licensed_area_ha, participant_count, winner_date
(prefix with - for descending; default -winner_date).
--qualified requires --participant.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MiningLicenseAuctionsRetrieveParams{
				Province:        cmdutil.OptStr(cmd, "province", province),
				CommodityType:   cmdutil.OptStr(cmd, "commodity-type", commodityType),
				OrderBy:         cmdutil.OptEnum[sectors.MiningLicenseAuctionsRetrieveParamsOrderBy](cmd, "order-by", orderBy),
				Limit:           cmdutil.OptInt(cmd, "limit", limit),
				Offset:          cmdutil.OptInt(cmd, "offset", offset),
				AreaType:        cmdutil.OptStr(cmd, "area-type", areaType),
				Status:          cmdutil.OptStr(cmd, "status", status),
				Participant:     cmdutil.OptStr(cmd, "participant", participant),
				Qualified:       cmdutil.OptBool(cmd, "qualified", qualified),
				MinParticipants: cmdutil.OptInt(cmd, "min-participants", minParticipants),
			}
			resp, err := client.MiningLicenseAuctionsRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&province, "province", "", "province filter")
	f.StringVar(&commodityType, "commodity-type", "", "commodity type filter")
	f.StringVar(&orderBy, "order-by", "", "commodity_type|licensed_area_ha|participant_count|winner_date")
	f.StringVar(&areaType, "area-type", "", "area type (e.g. WIUPK)")
	f.StringVar(&status, "status", "", "status (e.g. \"Lelang Selesai\")")
	f.StringVar(&participant, "participant", "", "participant name, partial match")
	f.BoolVar(&qualified, "qualified", false, "only qualified participants (requires --participant)")
	f.IntVar(&minParticipants, "min-participants", 0, "minimum participant count")
	f.IntVar(&limit, "limit", 20, "max results (max 30)")
	f.IntVar(&offset, "offset", 0, "pagination offset")
	return cmd
}

func newAuctionsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <wiup_code>",
		Short: "Full auction record incl. phases timeline and participants",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.Do(cmd, func(c *sectors.ClientWithResponses) (int, []byte, error) {
				r, err := c.MiningLicenseAuctionsRetrieve2WithResponse(cmd.Context(), args[0])
				if err != nil {
					return 0, nil, err
				}
				return r.StatusCode(), r.Body, nil
			})
		},
	}
}

// --- contracts -------------------------------------------------------------

func newContractsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contracts",
		Short: "Active mining contracts linking owners to contractors",
	}
	cmd.AddCommand(newContractsListCmd())
	return cmd
}

func newContractsListCmd() *cobra.Command {
	var contractor, mineOwner string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mining contracts, optionally filtered by owner or contractor",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cmdutil.NewClient(cmd)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			params := &sectors.MiningContractsRetrieveParams{
				Contractor: cmdutil.OptStr(cmd, "contractor", contractor),
				MineOwner:  cmdutil.OptStr(cmd, "mine-owner", mineOwner),
			}
			resp, err := client.MiningContractsRetrieveWithResponse(cmd.Context(), params)
			if err != nil {
				return cmdutil.Fail(0, err.Error(), nil)
			}
			return cmdutil.Emit(cmd, resp.StatusCode(), resp.Body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&contractor, "contractor", "", "contractor slug filter")
	f.StringVar(&mineOwner, "mine-owner", "", "mine owner slug filter")
	return cmd
}
