package keeper

import (
	"context"

	"github.com/umee-network/umee/x/oracle/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) AggregateExchangeRatePrevote(context.Context,
	*types.MsgAggregateExchangeRatePrevote) (*types.MsgAggregateExchangeRatePrevoteResponse,
	error) {
	return nil, nil
}

func (m msgServer) AggregateExchangeRateVote(context.Context,
	*types.MsgAggregateExchangeRateVote) (*types.MsgAggregateExchangeRateVoteResponse,
	error) {
	return nil, nil
}

func (m msgServer) DelegateFeedConsent(context.Context,
	*types.MsgDelegateFeedConsent) (*types.MsgDelegateFeedConsentResponse,
	error) {
	return nil, nil
}

var _ types.MsgServer = msgServer{}
