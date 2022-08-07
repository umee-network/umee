package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// GetSupplied returns an sdk.Coin representing how much of a given denom a
// user has supplied, including interest accrued.
func (k Keeper) GetSupplied(ctx sdk.Context, supplierAddr sdk.AccAddress, denom string) (sdk.Coin, error) {
	if err := k.ValidateAcceptedDenom(ctx, denom); err != nil {
		return sdk.Coin{}, err
	}

	// sum wallet-held and collateral-enabled uTokens in the associated uToken denom
	uDenom := types.ToUTokenDenom(denom)
	balance := k.bankKeeper.GetBalance(ctx, supplierAddr, uDenom)
	collateral := k.GetCollateral(ctx, supplierAddr, uDenom)

	// convert uTokens to tokens
	return k.ExchangeUToken(ctx, balance.Add(collateral))
}

// GetAllSupplied returns the total tokens supplied by a user, including
// any interest accrued.
func (k Keeper) GetAllSupplied(ctx sdk.Context, supplierAddr sdk.AccAddress) (sdk.Coins, error) {
	// get all uTokens set as collateral
	collateral := k.GetBorrowerCollateral(ctx, supplierAddr)

	// get all uTokens not set as collateral by filtering non-uTokens from supplier balance
	uTokens := sdk.Coins{}
	balance := k.bankKeeper.GetAllBalances(ctx, supplierAddr)
	for _, coin := range balance {
		if k.ValidateAcceptedUTokenDenom(ctx, coin.Denom) == nil {
			uTokens = uTokens.Add(coin)
		}
	}

	// convert the sum of found uTokens to base tokens
	return k.ExchangeUTokens(ctx, collateral.Add(uTokens...))
}

// GetTotalSupply returns the total supplied by all suppliers in a given denom,
// including any interest accrued.
func (k Keeper) GetTotalSupply(ctx sdk.Context, denom string) (sdk.Coin, error) {
	if err := k.ValidateAcceptedDenom(ctx, denom); err != nil {
		return sdk.Coin{}, err
	}

	// convert associated uToken's total supply to base tokens
	uTokenDenom := types.ToUTokenDenom(denom)
	return k.ExchangeUToken(ctx, k.GetUTokenSupply(ctx, uTokenDenom))
}
