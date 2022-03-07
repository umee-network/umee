package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// GetLent returns an sdk.Coin representing how much of a given denom a
// lender has lent, including interest accrued.
func (k Keeper) GetLent(ctx sdk.Context, lenderAddr sdk.AccAddress, denom string) (sdk.Coin, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	// sum wallet-held and collateral-enabled uTokens in the associated uToken denom
	uDenom := k.FromTokenToUTokenDenom(ctx, denom)
	balance := k.bankKeeper.GetBalance(ctx, lenderAddr, uDenom)
	collateral := k.GetCollateralAmount(ctx, lenderAddr, uDenom)

	// convert uTokens to tokens
	return k.ExchangeUToken(ctx, balance.Add(collateral))
}

// GetLenderLent returns the total tokens lent by a lender across all denoms,
// including any interest accrued.
func (k Keeper) GetLenderLent(ctx sdk.Context, lenderAddr sdk.AccAddress) (sdk.Coins, error) {
	// get all uTokens set as collateral
	collateral := k.GetBorrowerCollateral(ctx, lenderAddr)

	// get all uTokens not set as collateral by filtering non-uTokens from lender balance
	uTokens := sdk.Coins{}
	balance := k.bankKeeper.GetAllBalances(ctx, lenderAddr)
	for _, coin := range balance {
		if k.IsAcceptedUToken(ctx, coin.Denom) {
			uTokens = uTokens.Add(coin)
		}
	}

	// convert the sum of found uTokens to base tokens
	return k.ExchangeUTokens(ctx, collateral.Add(uTokens...))
}

// GetTotalLent returns the total lent by all lenders in a given denom,
// including any interest accrued.
func (k Keeper) GetTotalLent(ctx sdk.Context, denom string) (sdk.Coin, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	// convert associated uToken's total supply to base tokens
	uTokenDenom := k.FromTokenToUTokenDenom(ctx, denom)
	return k.ExchangeUToken(ctx, k.GetUTokenSupply(ctx, uTokenDenom))
}
