package client

import (
	"context"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (qc *QueryClient) GovQueryClient() govtypes.QueryClient {
	return govtypes.NewQueryClient(qc.grpcConn)
}

func (qc *QueryClient) QueryProposal(proposalID uint64) (*govtypes.Proposal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	queryResponse, err := qc.GovQueryClient().Proposal(ctx, &govtypes.QueryProposalRequest{ProposalId: proposalID})
	if err != nil {
		return nil, err
	}
	return queryResponse.Proposal, nil
}
