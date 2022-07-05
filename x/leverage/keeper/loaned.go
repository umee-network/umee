package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// GetLoaned returns an sdk.Coin representing how much of a given denom a
// supplier has loaned, including interest accrued.
func (k Keeper) GetLoaned(ctx sdk.Context, supplierAddr sdk.AccAddress, denom string) (sdk.Coin, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	// sum wallet-held and collateral-enabled uTokens in the associated uToken denom
	uDenom := k.FromTokenToUTokenDenom(ctx, denom)
	balance := k.bankKeeper.GetBalance(ctx, supplierAddr, uDenom)
	collateral := k.GetCollateralAmount(ctx, supplierAddr, uDenom)

	// convert uTokens to tokens
	return k.ExchangeUToken(ctx, balance.Add(collateral))
}

// GetSupplierLoaned returns the total tokens loaned by a supplier across all denoms,
// including any interest accrued.
func (k Keeper) GetSupplierLoaned(ctx sdk.Context, supplierAddr sdk.AccAddress) (sdk.Coins, error) {
	// get all uTokens set as collateral
	collateral := k.GetBorrowerCollateral(ctx, supplierAddr)

	// get all uTokens not set as collateral by filtering non-uTokens from supplier balance
	uTokens := sdk.Coins{}
	balance := k.bankKeeper.GetAllBalances(ctx, supplierAddr)
	for _, coin := range balance {
		if k.IsAcceptedUToken(ctx, coin.Denom) {
			uTokens = uTokens.Add(coin)
		}
	}

	// convert the sum of found uTokens to base tokens
	return k.ExchangeUTokens(ctx, collateral.Add(uTokens...))
}

// GetTotalLoaned returns the total loaned by all suppliers in a given denom,
// including any interest accrued.
func (k Keeper) GetTotalLoaned(ctx sdk.Context, denom string) (sdk.Coin, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	// convert associated uToken's total supply to base tokens
	uTokenDenom := k.FromTokenToUTokenDenom(ctx, denom)
	return k.ExchangeUToken(ctx, k.GetUTokenSupply(ctx, uTokenDenom))
}
