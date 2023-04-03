package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/leverage/types"
)

// DonateCollateral burns some collateral uTokens already present in the module, then adds their equivalent amount
// in tokens reserves. Currently, this is only used as the penalty for incentive module's MsgEmergencyUnbond.
func (k Keeper) DonateCollateral(ctx sdk.Context, fromAddr sdk.AccAddress, uToken sdk.Coin) error {
	token, err := k.ExchangeUToken(ctx, uToken)
	if err != nil {
		return err
	}
	if err = k.burnCollateral(ctx, fromAddr, uToken); err != nil {
		return err
	}

	// increase module reserves
	reserves := k.GetReserves(ctx, token.Denom)
	return k.setReserves(ctx, reserves.Add(token))
}

// forceSetCollateral claims rewards and finishes completed unbondings for an account which may have collateral
// bonded to the incentive module, then detects if the sum of its bonded and unbonding amounts are greater than
// the new value. If they are, forcefully finishes unbondings than instantly unbonds collateral until their sum
// is equal to the collateral amount.
func (k Keeper) forceSetCollateral(ctx sdk.Context, addr sdk.AccAddress, collateral sdk.Coin) error {
	if k.incentiveKeeper == nil {
		return types.ErrIncentiveKeeperNotSet
	}
	return k.incentiveKeeper.ForceSetCollateral(ctx, addr, collateral)
}

// getBondedAndUnbonding sums the bonded and unbonding collateral amounts for an account which may have collateral
// bonded to the incentive module.
func (k Keeper) getBondedAndUnbonding(ctx sdk.Context, addr sdk.AccAddress, uDenom string) (sdk.Coin, error) {
	if k.incentiveKeeper == nil {
		return sdk.Coin{}, types.ErrIncentiveKeeperNotSet
	}
	return k.incentiveKeeper.RestrictedCollateral(ctx, addr, uDenom), nil
}
