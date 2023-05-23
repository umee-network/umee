package tx

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	proposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

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

// TxGovVoteYesAll creates transactions (one for each registered account) to approve a given proposal.
func (c *Client) TxGovVoteYesAll(proposalID uint64) error {
	for index := range c.keyringRecord {
		voter, err := c.keyringRecord[index].GetAddress()
		if err != nil {
			return err
		}

		voteType, err := govtypes.VoteOptionFromString("VOTE_OPTION_YES")
		if err != nil {
			return err
		}

		msg := govtypes.NewMsgVote(
			voter,
			proposalID,
			voteType,
		)

		c.ClientContext.From = c.keyringRecord[index].Name
		c.ClientContext.FromName = c.keyringRecord[index].Name
		c.ClientContext.FromAddress, _ = c.keyringRecord[index].GetAddress()

		for retry := 0; retry < 5; retry++ {
			// retry if txs fails, because sometimes account sequence mismatch occurs due to txs pending
			if _, err = BroadcastTx(*c.ClientContext, *c.txFactory, []sdk.Msg{msg}...); err == nil {
				break
			}
			time.Sleep(time.Second * 1)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
