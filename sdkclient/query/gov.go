package query

import (
	"context"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (c *Client) GovClient() govtypes.QueryClient {
	return govtypes.NewQueryClient(c.grpcConn)
}

func (c *Client) GovProposal(proposalID uint64) (*govtypes.Proposal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	queryResponse, err := c.GovClient().Proposal(ctx, &govtypes.QueryProposalRequest{ProposalId: proposalID})
	if err != nil {
		return nil, err
	}
	return queryResponse.Proposal, nil
}
