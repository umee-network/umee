package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/incentive"
)

func (k Keeper) getParams(ctx sdk.Context) incentive.Params {
	return incentive.Params{
		MaxUnbondings:           k.getMaxUnbondings(ctx),
		TierWeightShort:         k.getTierWeightShort(ctx),
		TierWeightMiddle:        k.getTierWeightMiddle(ctx),
		UnbondingDurationLong:   k.getUnbondingDurationLong(ctx),
		UnbondingDurationMiddle: k.getUnbondingDurationMiddle(ctx),
		UnbondingDurationShort:  k.getUnbondingDurationShort(ctx),
		CommunityFundAddress:    k.getCommunityFundAddress(ctx).String(),
	}
}
