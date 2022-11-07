package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of MsgServer for the x/leverage
// module.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

// AddIBCDenomRateLimit implements types.MsgServer
func (s msgServer) UpdateIBCDenomsRateLimit(goCtx context.Context, msg *types.MsgUpdateIBCDenomsRateLimit) (*types.MsgUpdateIBCDenomsRateLimitResponse, error) {
	_ = sdk.UnwrapSDKContext(goCtx)

	// checking req msg authority is the gov module address
	if s.keeper.authority != msg.Authority {
		return &types.MsgUpdateIBCDenomsRateLimitResponse{},
			govtypes.ErrInvalidSigner.Wrapf(
				"invalid authority: expected %s, got %s",
				s.keeper.authority, msg.Authority,
			)
	}

	// TODO: Still needs to create store and getters and setters for rate limits

	return &types.MsgUpdateIBCDenomsRateLimitResponse{}, nil
}
