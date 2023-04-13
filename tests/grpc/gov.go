package grpc

import (
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	gtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	proposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	"github.com/umee-network/umee/v4/client"
	"github.com/umee-network/umee/v4/x/uibc"
)

var govDeposit sdk.Coins

func init() {
	var err error
	govDeposit, err = sdk.ParseCoinsNormalized("10000000uumee")
	if err != nil {
		panic(err)
	}
}

func SubmitAndPassProposal(umee client.Client, changes []proposal.ParamChange) error {
	resp, err := umee.Tx.GovSubmitProposal(changes, govDeposit)
	if err != nil {
		return err
	}

	return MakeVoteAndCheckProposal(umee, *resp)
}

func OracleParamChanges(
	historicStampPeriod uint64,
	maximumPriceStamps uint64,
	medianStampPeriod uint64,
) []proposal.ParamChange {
	return []proposal.ParamChange{
		{
			Subspace: "oracle",
			Key:      "HistoricStampPeriod",
			Value:    fmt.Sprintf("\"%d\"", historicStampPeriod),
		},
		{
			Subspace: "oracle",
			Key:      "MaximumPriceStamps",
			Value:    fmt.Sprintf("\"%d\"", maximumPriceStamps),
		},
		{
			Subspace: "oracle",
			Key:      "MedianStampPeriod",
			Value:    fmt.Sprintf("\"%d\"", medianStampPeriod),
		},
	}
}

func UIBCIBCTransferSatusUpdate(umeeClient client.Client, status uibc.IBCTransferStatus) error {
	msg := uibc.MsgGovSetIBCStatus{
		Authority:   authtypes.NewModuleAddress(gtypes.ModuleName).String(),
		Title:       "Update the ibc transfer status",
		Description: "Update the ibc transfer status",
		IbcStatus:   status,
	}

	var err error
	for retry := 0; retry < 5; retry++ {
		// retry if txs fails, because sometimes account sequence mismatch occurs due to txs pending
		if resp, err := umeeClient.Tx.TxSubmitProposalWithMsg([]sdk.Msg{&msg}); err == nil {
			return MakeVoteAndCheckProposal(umeeClient, *resp)
		}
		time.Sleep(time.Second * 1)
	}

	return err
}

func MakeVoteAndCheckProposal(umeeClient client.Client, resp sdk.TxResponse) error {
	var proposalID string
	for _, event := range resp.Logs[0].Events {
		if event.Type == "submit_proposal" {
			for _, attribute := range event.Attributes {
				if attribute.Key == "proposal_id" {
					proposalID = attribute.Value
				}
			}
		}
	}

	if proposalID == "" {
		return fmt.Errorf("failed to parse proposalID from %s", resp)
	}

	proposalIDInt, err := strconv.ParseUint(proposalID, 10, 64)
	if err != nil {
		return err
	}

	err = umeeClient.Tx.TxGovVoteYesAll(proposalIDInt)
	if err != nil {
		return err
	}

	prop, err := umeeClient.GovProposal(proposalIDInt)
	if err != nil {
		return err
	}

	now := time.Now()
	sleepDuration := prop.VotingEndTime.Sub(now) + 3*time.Second
	fmt.Printf("sleeping %s until end of voting period + 1 block\n", sleepDuration)
	time.Sleep(sleepDuration)

	prop, err = umeeClient.GovProposal(proposalIDInt)
	if err != nil {
		return err
	}

	propStatus := prop.Status.String()
	if propStatus != "PROPOSAL_STATUS_PASSED" {
		return fmt.Errorf("proposal %d failed to pass with status: %s", proposalIDInt, propStatus)
	}
	return nil
}
