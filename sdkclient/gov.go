package sdkclient

import (
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	proposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

func (c *Client) GovParamChange(title, description string, changes []proposal.ParamChange, deposit sdk.Coins,
) (*sdk.TxResponse, error) {
	content := proposal.NewParameterChangeProposal(title, description, changes)
	fromAddr, err := c.keyringRecord[0].GetAddress()
	if err != nil {
		return nil, err
	}
	msg, err := govv1b1.NewMsgSubmitProposal(content, deposit, fromAddr)
	if err != nil {
		return nil, err
	}

	return c.BroadcastTx(0, msg)
}

// TODO: update the content title and summary
func (c *Client) GovSubmitParamProp(changes []proposal.ParamChange, deposit sdk.Coins) (*sdk.TxResponse, error) {
	content := proposal.NewParameterChangeProposal(
		"update historic stamp period",
		"auto grpc proposal",
		changes,
	)

	fromAddr, err := c.keyringRecord[0].GetAddress()
	if err != nil {
		return nil, err
	}
	msg, err := govv1b1.NewMsgSubmitProposal(content, deposit, fromAddr)
	if err != nil {
		return nil, err
	}

	return c.BroadcastTx(0, msg)
}

func (c *Client) GovSubmitProp(msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	// TODO: deposit should be parsed form the msgs
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

func (c *Client) GovSubmitPropAndGetID(msgs ...sdk.Msg) (uint64, error) {
	resp, err := c.GovSubmitProp(msgs...)
	if err != nil {
		return 0, err
	}
	resp, err = c.GetTxResponse(resp.TxHash, 1)
	if err != nil {
		return 0, err
	}

	var proposalID string
	for _, event := range resp.Events {
		if event.Type == "submit_proposal" {
			for _, attribute := range event.Attributes {
				if attribute.Key == "proposal_id" {
					proposalID = attribute.Value
				}
			}
		}
	}
	if proposalID == "" {
		return 0, fmt.Errorf("failed to parse proposalID from %s", resp)
	}

	return strconv.ParseUint(proposalID, 10, 64)
}

// GovVote votes for a x/gov proposal. If `vote==nil` then vote yes.
func (c *Client) GovVote(proposalID uint64, vote *govv1b1.VoteOption) error {
	voteOpt := govv1b1.OptionYes
	if vote != nil {
		voteOpt = *vote
	}
	voter, err := c.keyringRecord[0].GetAddress()
	if err != nil {
		return err
	}
	msg := govv1b1.NewMsgVote(
		voter,
		proposalID,
		voteOpt,
	)
	_, err = c.BroadcastTx(0, msg)
	return err
}

// GovVoteAllYes creates transactions (one for each account in the keyring) to approve a given proposal.
// Deprecated.
func (c *Client) GovVoteAllYes(proposalID uint64) error {
	for index := range c.keyringRecord {
		voter, err := c.keyringRecord[index].GetAddress()
		if err != nil {
			return err
		}

		voteType, err := govv1b1.VoteOptionFromString("VOTE_OPTION_YES")
		if err != nil {
			return err
		}

		msg := govv1b1.NewMsgVote(
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

/*****************
  QUERIES
  **************** */

func (c Client) GovQClient() govv1.QueryClient {
	return govv1.NewQueryClient(c.GrpcConn)
}

// GovProposal queries a proposal by id
func (c Client) GovProposal(proposalID uint64) (*govv1.Proposal, error) {
	ctx, cancel := c.NewCtxWithTimeout()
	defer cancel()

	queryResponse, err := c.GovQClient().Proposal(ctx, &govv1.QueryProposalRequest{ProposalId: proposalID})
	if err != nil {
		return nil, err
	}
	return queryResponse.Proposal, nil
}
