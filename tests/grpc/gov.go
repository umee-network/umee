package grpc

import (
	"fmt"
	"strconv"
	"time"

	proposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/umee-network/umee/v4/tests/grpc/client"
)

func SubmitAndPassProposal(umeeClient *client.UmeeClient) error {
	oracleParamChanges := OracleParamChanges(10, 2, 20)
	resp, err := umeeClient.TxClient.TxSubmitProposal(oracleParamChanges)
	if err != nil {
		return err
	}

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
		return fmt.Errorf("Failed to parse proposalID from %s", resp)
	}

	proposalIDInt, err := strconv.ParseUint(proposalID, 10, 64)
	if err != nil {
		return err
	}

	_, err = umeeClient.TxClient.TxVoteYes(proposalIDInt)
	if err != nil {
		return err
	}

	prop, err := umeeClient.QueryClient.QueryProposal(proposalIDInt)
	if err != nil {
		return err
	}

	now := time.Now()
	sleepDuration := prop.VotingEndTime.Sub(now) + 2*time.Second
	fmt.Printf("sleeping %s until end of voting period + 1 block\n", sleepDuration)
	time.Sleep(sleepDuration)

	prop, err = umeeClient.QueryClient.QueryProposal(proposalIDInt)
	if err != nil {
		return err
	}

	propStatus := prop.Status.String()
	if propStatus != "PROPOSAL_STATUS_PASSED" {
		return fmt.Errorf("Proposal %d failed to pass with status: %s", proposalIDInt, propStatus)
	}
	return nil
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
