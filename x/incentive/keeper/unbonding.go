package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/incentive"
)

// addUnbonding creates an unbonding and adds it to the account's current unbondings in the store.
// Assumes the validity of the unbonding has already been checked.
func (k Keeper) addUnbonding(ctx sdk.Context, addr sdk.AccAddress, uToken sdk.Coin, tier incentive.BondTier) error {
	unbonding := incentive.Unbonding{
		Amount: uToken,
		End:    k.GetLastRewardsTime(ctx) + k.unbondTime(ctx, tier),
	}
	unbondings := incentive.AccountUnbondings{
		Account:    addr.String(),
		Tier:       uint32(tier),
		Denom:      uToken.Denom,
		Unbondings: append(k.GetUnbondings(ctx, addr, uToken.Denom, tier), unbonding),
	}
	return k.SetUnbondings(ctx, unbondings, tier, uToken.Denom)
}

// bondTier converts from the uint32 used in message types to the enumeration, returning an error
// if it is not valid. Does not allow incentive.BondTierUnspecified
func bondTier(n uint32) (incentive.BondTier, error) {
	if n == 0 || n > uint32(incentive.BondTierLong) {
		return incentive.BondTierUnspecified, incentive.ErrInvalidTier.Wrapf("%d", n)
	}
	return incentive.BondTier(n), nil
}

// unbondTime returns how long a given tier must wait to unbond, in seconds.
// returns zero on invalid unbond tier, though this will not happen in practice.
func (k Keeper) unbondTime(ctx sdk.Context, tier incentive.BondTier) uint64 {
	switch tier {
	case incentive.BondTierLong:
		return k.GetUnbondingDurationLong(ctx)
	case incentive.BondTierMiddle:
		return k.GetUnbondingDurationMiddle(ctx)
	case incentive.BondTierShort:
		return k.GetUnbondingDurationShort(ctx)
	default:
		return 0
	}
}
