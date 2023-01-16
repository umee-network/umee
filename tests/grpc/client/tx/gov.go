package tx

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	proposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

func (c *Client) TxVoteYes(proposalID uint64) (*sdk.TxResponse, error) {
	voter, err := c.keyringRecord.GetAddress()
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
	return c.BroadcastTx(msg)
}

func (c *Client) TxSubmitProposal(
	changes []proposal.ParamChange,
) (*sdk.TxResponse, error) {

	content := proposal.NewParameterChangeProposal(
		"update historic stamp period",
		"auto grpc proposal",
		changes,
	)

	deposit, err := sdk.ParseCoinsNormalized("10000000uumee")
	if err != nil {
		return nil, err
	}

	fromAddr, err := c.keyringRecord.GetAddress()
	if err != nil {
		return nil, err
	}

	msg, err := govtypes.NewMsgSubmitProposal(content, deposit, fromAddr)
	if err != nil {
		return nil, err
	}

	return c.BroadcastTx(msg)
}
