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

	return c.BroadcastTx(0, msg)
}

func (c *Client) GovSubmitParamProposal(changes []proposal.ParamChange, deposit sdk.Coins) (*sdk.TxResponse, error) {
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

	return c.BroadcastTx(0, msg)
}

func (c *Client) GovSubmitProposal(msgs []sdk.Msg) (*sdk.TxResponse, error) {
	deposit, err := sdk.ParseCoinsNormalized("1000uumee")
	if err != nil {
		return nil, err
	}

	fromAddr, err := c.keyringRecord[0].GetAddress()
	if err != nil {
		return nil, err
	}

	submitProposal, err := v1.NewMsgSubmitProposal(
		msgs,
		deposit,
		fromAddr.String(),
		"metadata",
		"sometitle",
		"somesummary",
	)
	if err != nil {
		return nil, err
	}

	return c.BroadcastTx(0, submitProposal)
}

// GovVoteAllYes creates transactions (one for each account in the keyring) to approve a given proposal.
func (c *Client) GovVoteAllYes(proposalID uint64) error {
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
		for retry := 0; retry < 3; retry++ {
			// retry if txs fails, because sometimes account sequence mismatch occurs due to txs pending
			if _, err = c.BroadcastTx(index, msg); err == nil {
				break
			}
			c.logger.Println("Tx broadcast failed. RETRYING. Err: ", err)
			time.Sleep(time.Millisecond * 300)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
