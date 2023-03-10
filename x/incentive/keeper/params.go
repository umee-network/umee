package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/incentive"
)

func (k Keeper) GetParams(ctx sdk.Context) incentive.Params {
	return incentive.Params{
		MaxUnbondings:           k.GetMaxUnbondings(ctx),
		TierWeightShort:         k.GetTierWeightShort(ctx),
		TierWeightMiddle:        k.GetTierWeightMiddle(ctx),
		UnbondingDurationLong:   k.GetUnbondingDurationLong(ctx),
		UnbondingDurationMiddle: k.GetUnbondingDurationMiddle(ctx),
		UnbondingDurationShort:  k.GetUnbondingDurationShort(ctx),
		CommunityFundAddress:    k.GetCommunityFundAddress(ctx).String(),
	}
}
