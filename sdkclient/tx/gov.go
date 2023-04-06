package tx

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	proposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

func (c *Client) GovVoteYes(proposalID uint64) error {
	return c.BroadcastTxVotes(proposalID)
}

func (c *Client) GovParamChange(title, description string, changes []proposal.ParamChange, deposit sdk.Coins,
) (*sdk.TxResponse, error) {
	content := proposal.NewParameterChangeProposal(title, description, changes)
	fromAddr, err := c.keyringRecord[0].GetAddress()
	if err != nil {
		return nil, err
	}
	msg, err := govtypes.NewMsgSubmitProposal(content, deposit, fromAddr)
	if err != nil {
		return nil, err
	}

	return c.BroadcastTx(msg)
}

func (c *Client) GovSubmitProposal(changes []proposal.ParamChange, deposit sdk.Coins) (*sdk.TxResponse, error) {
	content := proposal.NewParameterChangeProposal(
		"update historic stamp period",
		"auto grpc proposal",
		changes,
	)

	fromAddr, err := c.keyringRecord[0].GetAddress()
	if err != nil {
		return nil, err
	}
	msg, err := govtypes.NewMsgSubmitProposal(content, deposit, fromAddr)
	if err != nil {
		return nil, err
	}

	return c.BroadcastTx(msg)
}

func (c *Client) TxSubmitProposalWithMsg(msgs []sdk.Msg) (*sdk.TxResponse, error) {
	deposit, err := sdk.ParseCoinsNormalized("1000uumee")
	if err != nil {
		return nil, err
	}

	fromAddr, err := c.keyringRecord[0].GetAddress()
	if err != nil {
		return nil, err
	}

	submitProposal, err := v1.NewMsgSubmitProposal(msgs, deposit, fromAddr.String(), "")
	if err != nil {
		return nil, err
	}

	return c.BroadcastTx(submitProposal)
}
