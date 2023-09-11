package grpc

import (
	"fmt"
	"strconv"
	"time"

	"github.com/umee-network/umee/v6/x/metoken"

	ltypes "github.com/umee-network/umee/v6/x/leverage/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	"github.com/umee-network/umee/v6/client"
	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/x/uibc"
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
		Authority:   checkers.GovModuleAddr,
		Description: "",
		IbcStatus:   status,
	}

	resp, err := umeeClient.Tx.TxSubmitProposalWithMsg([]sdk.Msg{&msg})
	if err != nil {
		return err
	}

	if len(resp.Logs) == 0 {
		return fmt.Errorf("no logs in response")
	}

	return MakeVoteAndCheckProposal(umeeClient, *resp)
}

func LeverageRegistryUpdate(umeeClient client.Client, addTokens, updateTokens []ltypes.Token) error {
	msg := ltypes.MsgGovUpdateRegistry{
		Authority:    checkers.GovModuleAddr,
		Description:  "",
		AddTokens:    addTokens,
		UpdateTokens: updateTokens,
	}

	resp, err := umeeClient.Tx.TxSubmitProposalWithMsg([]sdk.Msg{&msg})
	if err != nil {
		return err
	}

	if len(resp.Logs) == 0 {
		return fmt.Errorf("no logs in response")
	}

	return MakeVoteAndCheckProposal(umeeClient, *resp)
}

func MetokenRegistryUpdate(umeeClient client.Client, addIndexes, updateIndexes []metoken.Index) error {
	msg := metoken.MsgGovUpdateRegistry{
		Authority:   checkers.GovModuleAddr,
		AddIndex:    addIndexes,
		UpdateIndex: updateIndexes,
	}

	resp, err := umeeClient.Tx.TxSubmitProposalWithMsg([]sdk.Msg{&msg})
	if err != nil {
		return err
	}

	if len(resp.Logs) == 0 {
		return fmt.Errorf("no logs in response")
	}

	return MakeVoteAndCheckProposal(umeeClient, *resp)
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
	sleepDuration := prop.VotingEndTime.Sub(now) + 1*time.Second
	fmt.Printf("sleeping %s until end of voting period + 1 block\n", sleepDuration)
	time.Sleep(sleepDuration)

	var propStatus string
	for retry := 0; retry < 5; retry++ {
		prop, err = umeeClient.GovProposal(proposalIDInt)
		if err != nil {
			return err
		}

		propStatus = prop.Status.String()
		if propStatus == "PROPOSAL_STATUS_PASSED" {
			return nil
		}
		time.Sleep(time.Second * (1 + time.Duration(retry)))
	}

	return fmt.Errorf("proposal %d failed to pass with status: %s", proposalIDInt, propStatus)
}
