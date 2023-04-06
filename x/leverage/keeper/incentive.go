package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
