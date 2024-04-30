package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/auction"
	"github.com/umee-network/umee/v6/x/metoken"
	"github.com/umee-network/umee/v6/x/metoken/errors"
)

// redeemResponse wraps all the coins of a successful redemption
type redeemResponse struct {
	fee          sdk.Coin
	fromReserves sdk.Coin
	fromLeverage sdk.Coin
}

func newRedeemResponse(fee sdk.Coin, fromReserves sdk.Coin, fromLeverage sdk.Coin) redeemResponse {
	return redeemResponse{
		fee:          fee,
		fromReserves: fromReserves,
		fromLeverage: fromLeverage,
	}
}

// redeem executes all the necessary calculations and transactions to perform a redemption between users meTokens and
// an accepted asset by the Index.
// A redemption includes the following actions:
// - Calculate the fee to charge to the user.
// - Calculate the amount of assets to return to the user
// - Calculate the amount of assets to be withdrawn from x/metoken reserves and x/leverage pools.
// - Withdraw the calculated portion from the x/metoken reserves.
// - Withdraw the calculated portion from the x/leverage liquidity pools.
// - Transfer meTokens from users account to x/metoken.
// - Burn meTokens.
//
// It returns: fee charged to the user, assets withdrawn from x/metoken and x/leverage
func (k Keeper) redeem(userAddr sdk.AccAddress, meToken sdk.Coin, assetDenom string) (redeemResponse, error) {
	index, err := k.RegisteredIndex(meToken.Denom)
	if err != nil {
		return redeemResponse{}, err
	}

	indexPrices, err := k.Prices(index)
	if err != nil {
		return redeemResponse{}, err
	}

	balances, err := k.IndexBalances(meToken.Denom)
	if err != nil {
		return redeemResponse{}, err
	}

	if !balances.MetokenSupply.IsPositive() {
		return redeemResponse{}, fmt.Errorf("not enough %s supply", meToken.Denom)
	}

	amountFromReserves, amountFromLeverage, err := k.calculateRedeem(index, indexPrices, meToken, assetDenom)
	if err != nil {
		return redeemResponse{}, err
	}

	if amountFromReserves.IsZero() && amountFromLeverage.IsZero() {
		return redeemResponse{}, fmt.Errorf("insufficient %s for redemption", meToken.Denom)
	}

	tokensWithdrawn, err := k.withdrawFromLeverage(sdk.NewCoin(assetDenom, amountFromLeverage))
	if err != nil {
		return redeemResponse{}, err
	}

	// if there is a difference between the desired to withdraw from x/leverage and the withdrawn,
	// take it from x/metoken reserves
	if tokensWithdrawn.Amount.LT(amountFromLeverage) {
		tokenDiff := amountFromLeverage.Sub(tokensWithdrawn.Amount)
		amountFromReserves = amountFromReserves.Add(tokenDiff)
		amountFromLeverage = amountFromLeverage.Sub(tokenDiff)
	}

	balance, i := balances.AssetBalance(assetDenom)
	if i < 0 {
		return redeemResponse{}, sdkerrors.ErrNotFound.Wrapf(
			"balance for denom %s not found",
			assetDenom,
		)
	}

	if balance.Reserved.LT(amountFromReserves) {
		return redeemResponse{}, fmt.Errorf("not enough %s liquidity for redemption", assetDenom)
	}

	_, fees, err := k.redeemFee(
		index,
		indexPrices,
		sdk.NewCoin(assetDenom, amountFromReserves.Add(amountFromLeverage)),
	)
	if err != nil {
		return redeemResponse{}, err
	}
	feeToAuction, feeToRevenue := k.breakFee(fees.Amount)
	if err = k.fundAuction(fees.Denom, feeToAuction); err != nil {
		return redeemResponse{}, err
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(*k.ctx, userAddr, metoken.ModuleName, sdk.Coins{meToken})
	if err != nil {
		return redeemResponse{}, err
	}

	toRedeem := sdk.NewCoin(assetDenom, amountFromReserves.Add(amountFromLeverage).Sub(fees.Amount))
	if err = k.bankKeeper.SendCoinsFromModuleToAccount(
		*k.ctx,
		metoken.ModuleName,
		userAddr,
		sdk.NewCoins(toRedeem),
	); err != nil {
		return redeemResponse{}, err
	}

	// once all the transactions are completed, update the index balances
	// subtract burned meTokens from total supply
	balances.MetokenSupply.Amount = balances.MetokenSupply.Amount.Sub(meToken.Amount)
	// update reserved, leveraged and fee balances
	balance.Reserved = balance.Reserved.Sub(amountFromReserves)
	balance.Leveraged = balance.Leveraged.Sub(amountFromLeverage)
	balance.Fees = balance.Fees.Add(feeToRevenue)
	balances.SetAssetBalance(balance)

	// save index balance
	if err = k.setIndexBalances(balances); err != nil {
		return redeemResponse{}, err
	}

	// burn meTokens
	if err = k.bankKeeper.BurnCoins(*k.ctx, metoken.ModuleName, sdk.NewCoins(meToken)); err != nil {
		return redeemResponse{}, err
	}

	return newRedeemResponse(
		fees,
		sdk.NewCoin(assetDenom, amountFromReserves),
		sdk.NewCoin(assetDenom, amountFromLeverage),
	), nil
}

// withdrawFromLeverage before withdrawing from x/leverage check if it's possible to withdraw the desired amount
// based on x/leverage module constraints. When the full amount is not available withdraw the max possible.
// Returning args are:
//   - tokensWithdrawn: the amount tokens withdrawn from x/leverage.
//   - error
func (k Keeper) withdrawFromLeverage(tokensToWithdraw sdk.Coin) (sdk.Coin, error) {
	uTokensFromLeverage, err := k.leverageKeeper.ToUToken(*k.ctx, tokensToWithdraw)
	if err != nil {
		return sdk.Coin{}, errors.Wrap(err, true)
	}

	availableUTokensFromLeverage, err := k.leverageKeeper.ModuleMaxWithdraw(*k.ctx, uTokensFromLeverage)
	if err != nil {
		return sdk.Coin{}, errors.Wrap(err, true)
	}

	if availableUTokensFromLeverage.IsZero() {
		return coin.Zero(tokensToWithdraw.Denom), nil
	}

	tokensWithdrawn, recoverable, err := k.leverageKeeper.WithdrawToModule(
		*k.ctx,
		metoken.ModuleName,
		sdk.NewCoin(uTokensFromLeverage.Denom, sdk.MinInt(availableUTokensFromLeverage, uTokensFromLeverage.Amount)),
	)
	if err != nil {
		return sdk.Coin{}, errors.Wrap(err, recoverable)
	}

	return tokensWithdrawn, nil
}

// calculateRedeem returns the fee to be charged to the user,
// the amount of assets to withdraw from x/metoken reserves and from x/leverage pools.
// The formulas used for the calculations are:
//
//	assets_to_withdraw = metokens_to_burn * exchange_rate
//	amount_from_reserves = assets_to_withdraw * reserve_portion
//	amount_from_leverage = assets_to_withdraw - amount_from_reserves
//
// It returns the amount of assets to withdraw from x/metoken reserves and x/leverage liquidity pools.
func (k Keeper) calculateRedeem(
	index metoken.Index,
	indexPrices metoken.IndexPrices,
	meToken sdk.Coin,
	assetDenom string,
) (sdkmath.Int, sdkmath.Int, error) {
	assetSettings, i := index.AcceptedAsset(assetDenom)
	if i < 0 {
		return sdkmath.ZeroInt(), sdkmath.ZeroInt(), sdkerrors.ErrNotFound.Wrapf(
			"asset %s is not accepted in the index",
			assetDenom,
		)
	}

	amountToWithdraw, err := indexPrices.RedeemRate(meToken, assetDenom)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.ZeroInt(), err
	}

	if amountToWithdraw.IsZero() {
		return sdkmath.ZeroInt(), sdkmath.ZeroInt(), nil
	}

	amountFromReserves := assetSettings.ReservePortion.MulInt(amountToWithdraw).TruncateInt()
	amountFromLeverage := amountToWithdraw.Sub(amountFromReserves)

	return amountFromReserves, amountFromLeverage, nil
}
