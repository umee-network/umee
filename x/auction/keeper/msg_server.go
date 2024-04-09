package keeper

import (
	"context"

	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/util/sdkutil"
	"github.com/umee-network/umee/v6/x/auction"
)

var _ auction.MsgServer = msgServer{}

type msgServer struct {
	kb Builder
}

// NewMsgServer returns an implementation of auction.MsgServer
func NewMsgServer(kb Builder) auction.MsgServer {
	return &msgServer{kb: kb}
}

// GovSetRewardsParams implements types.MsgServer
func (m msgServer) GovSetRewardsParams(ctx context.Context, msg *auction.MsgGovSetRewardsParams) (
	*auction.MsgGovSetRewardsParamsResponse, error,
) {
	sdkCtx, err := sdkutil.StartMsg(ctx, msg)
	if err != nil {
		return nil, err
	}

	k := m.kb.Keeper(&sdkCtx)
	byEmergencyGroup, err := checkers.EmergencyGroupAuthority(msg.Authority, k.ugov)
	if err != nil {
		return nil, err
	}

	if err := k.SetRewardsParams(msg, byEmergencyGroup); err != nil {
		return nil, err
	}
	return &auction.MsgGovSetRewardsParamsResponse{}, nil
}

// RewardsBid implements types.MsgServer
func (m msgServer) RewardsBid(ctx context.Context, msg *auction.MsgRewardsBid) (
	*auction.MsgRewardsBidResponse, error,
) {
	sdkCtx, err := sdkutil.StartMsg(ctx, msg)
	if err != nil {
		return nil, err
	}

	k := m.kb.Keeper(&sdkCtx)
	if err := k.RewardsBid(msg); err != nil {
		return nil, err
	}
	return &auction.MsgRewardsBidResponse{}, nil
}
