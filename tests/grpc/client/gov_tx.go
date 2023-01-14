package client

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	proposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

func (tc *TxClient) TxVoteYes(proposalID uint64) (*sdk.TxResponse, error) {
	voter, err := tc.keyringRecord.GetAddress()
	if err != nil {
		return nil, err
	}

	voteType, err := govtypes.VoteOptionFromString("VOTE_OPTION_YES")
	if err != nil {
		return nil, err
	}

	msg := govtypes.NewMsgVote(
		voter,
		proposalID,
		voteType,
	)
	return tc.BroadcastTx(msg)
}

func (tc *TxClient) TxUpdateOracleParams(
	historicStampPeriod uint64,
	maximumPriceStamps uint64,
	medianStampPeriod uint64,
) (*sdk.TxResponse, error) {

	changes := []proposal.ParamChange{
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

	content := proposal.NewParameterChangeProposal(
		"update historic stamp period",
		"auto grpc proposal",
		changes,
	)

	deposit, err := sdk.ParseCoinsNormalized("10000000uumee")
	if err != nil {
		return nil, err
	}

	fromAddr, err := tc.keyringRecord.GetAddress()
	if err != nil {
		return nil, err
	}

	msg, err := govtypes.NewMsgSubmitProposal(content, deposit, fromAddr)
	if err != nil {
		return nil, err
	}

	return tc.BroadcastTx(msg)
}
