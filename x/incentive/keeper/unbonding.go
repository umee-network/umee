package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/incentive"
)

// addUnbonding creates an unbonding and adds it to the account's current unbondings in the store.
// Assumes the validity of the unbonding has already been checked. Also updates unbonding amounts
// indirectly by calling setUnbondings.
func (k Keeper) addUnbonding(ctx sdk.Context, addr sdk.AccAddress, uToken sdk.Coin, tier incentive.BondTier) error {
	if err := k.decreaseBond(ctx, addr, tier, uToken); err != nil {
		return err
	}
	unbonding := incentive.Unbonding{
		Amount: uToken,
		End:    k.getLastRewardsTime(ctx) + k.unbondTime(ctx, tier),
	}
	unbondings := incentive.AccountUnbondings{
		Account:    addr.String(),
		Tier:       uint32(tier),
		Denom:      uToken.Denom,
		Unbondings: append(k.getUnbondings(ctx, addr, uToken.Denom, tier), unbonding),
	}
	return k.setUnbondings(ctx, unbondings)
}

// unbondTime returns how long a given tier must wait to unbond, in seconds.
// returns zero on invalid unbond tier, though this will not happen in practice.
func (k Keeper) unbondTime(ctx sdk.Context, tier incentive.BondTier) uint64 {
	switch tier {
	case incentive.BondTierLong:
		return k.getUnbondingDurationLong(ctx)
	case incentive.BondTierMiddle:
		return k.getUnbondingDurationMiddle(ctx)
	case incentive.BondTierShort:
		return k.getUnbondingDurationShort(ctx)
	default:
		return 0
	}
}
