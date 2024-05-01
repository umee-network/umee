package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v6/x/auction"
	"github.com/umee-network/umee/v6/x/metoken"
	"github.com/umee-network/umee/v6/x/metoken/errors"
)

// swapResponse wraps all the coins of a successful swap
type swapResponse struct {
	meTokens  sdk.Coin
	fee       sdk.Coin
	reserved  sdk.Coin
	leveraged sdk.Coin
}

func newSwapResponse(meTokens, fee, reserved, leveraged sdk.Coin) swapResponse {
	return swapResponse{
		meTokens:  meTokens,
		fee:       fee,
		reserved:  reserved,
		leveraged: leveraged,
	}
}

// swap executes all the necessary calculations and transactions to perform a swap between users asset and meToken.
// A swap includes the following actions:
// - Calculate the calculateFee to charge to the user.
// - Calculate the amount of meTokens to be minted.
// - Mint meTokens.
// - Calculate the amount of user's assets to send to x/metoken reserves and x/leverage pools.
// - Transfer the calculated portion to the x/metoken reserves.
// - Transfer the calculated portion to the x/leverage liquidity pools.
// - Transfer to the user the minted meTokens.
//
// It returns: minted meTokens, charged fee, assets transferred to reserves and assets transferred to x/leverage.
func (k Keeper) swap(userAddr sdk.AccAddress, meTokenDenom string, asset sdk.Coin) (swapResponse, error) {
	index, err := k.RegisteredIndex(meTokenDenom)
	if err != nil {
		return swapResponse{}, err
	}

	indexPrices, err := k.Prices(index)
	if err != nil {
		return swapResponse{}, err
	}

	// meTokenAmount, fee, amountToReserves, amountToLeverage, err :=
	swapCarry, err := k.calculateSwap(index, indexPrices, asset)
	if err != nil {
		return swapResponse{}, err
	}

	if swapCarry.meTokens.IsZero() {
		return swapResponse{}, fmt.Errorf("insufficient %s for swap", asset.Denom)
	}

	balances, err := k.IndexBalances(meTokenDenom)
	if err != nil {
		return swapResponse{}, err
	}

	if balances.MetokenSupply.Amount.Add(swapCarry.meTokens).GT(index.MaxSupply) {
		return swapResponse{}, fmt.Errorf(
			"not possible to mint the desired amount of %s, reaching the max supply",
			meTokenDenom,
		)
	}

	if err = k.bankKeeper.SendCoinsFromAccountToModule(
		*k.ctx,
		userAddr,
		metoken.ModuleName,
		sdk.NewCoins(asset),
	); err != nil {
		return swapResponse{}, err
	}

	supplied, err := k.supplyToLeverage(sdk.NewCoin(asset.Denom, swapCarry.toLeverage))
	if err != nil {
		return swapResponse{}, err
	}

	// adjust amount if supplied to x/leverage is less than the calculated amount
	if supplied.LT(swapCarry.toLeverage) {
		tokenDiff := swapCarry.toLeverage.Sub(supplied)
		swapCarry.toReserves = swapCarry.toReserves.Add(tokenDiff)
		swapCarry.toLeverage = swapCarry.toLeverage.Sub(tokenDiff)
	}

	meTokens := sdk.NewCoins(sdk.NewCoin(meTokenDenom, swapCarry.meTokens))
	if err = k.bankKeeper.MintCoins(*k.ctx, metoken.ModuleName, meTokens); err != nil {
		return swapResponse{}, err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(*k.ctx, metoken.ModuleName, userAddr, meTokens)
	if err != nil {
		return swapResponse{}, err
	}

	feeToAuction, feeToRevenue := k.breakFee(swapCarry.fee)
	if err = k.fundAuction(asset.Denom, feeToAuction); err != nil {
		return swapResponse{}, err
	}

	balances.MetokenSupply.Amount = balances.MetokenSupply.Amount.Add(swapCarry.meTokens)
	balance, i := balances.AssetBalance(asset.Denom)
	if i < 0 {
		return swapResponse{}, sdkerrors.ErrNotFound.Wrapf(
			"balance for denom %s not found", asset.Denom)
	}

	balance.Reserved = balance.Reserved.Add(swapCarry.toReserves)
	balance.Leveraged = balance.Leveraged.Add(swapCarry.toLeverage)
	balance.Fees = balance.Fees.Add(feeToRevenue)
	balances.SetAssetBalance(balance)

	if err = k.setIndexBalances(balances); err != nil {
		return swapResponse{}, err
	}

	return newSwapResponse(
		meTokens[0],
		sdk.NewCoin(asset.Denom, swapCarry.fee),
		sdk.NewCoin(asset.Denom, swapCarry.toReserves),
		sdk.NewCoin(asset.Denom, swapCarry.toLeverage),
	), nil
}

// supplyToLeverage before supplying to x/leverage check if it's possible to supply the desired amount
// based on x/leverage module constrains. When the full amount cannot be supplied, supply the max possible.
func (k Keeper) supplyToLeverage(tokensToSupply sdk.Coin) (sdkmath.Int, error) {
	isLimited, availableToSupply, err := k.availableToSupply(tokensToSupply.Denom)
	if err != nil {
		return sdkmath.Int{}, errors.Wrap(err, true)
	}

	if isLimited {
		if !availableToSupply.IsPositive() {
			return sdkmath.ZeroInt(), nil
		}

		if availableToSupply.LT(tokensToSupply.Amount) {
			tokensToSupply.Amount = availableToSupply
		}
	}

	if _, recoverable, err := k.leverageKeeper.SupplyFromModule(
		*k.ctx,
		metoken.ModuleName,
		tokensToSupply,
	); err != nil {
		return sdkmath.Int{}, errors.Wrap(err, recoverable)
	}

	return tokensToSupply.Amount, nil
}

// availableToSupply calculates the max amount could be supplied to x/leverage.
// Returns true and the max available if it is limited or false if it's unlimited.
func (k Keeper) availableToSupply(denom string) (bool, sdkmath.Int, error) {
	token, err := k.leverageKeeper.GetTokenSettings(*k.ctx, denom)
	if err != nil {
		return true, sdkmath.Int{}, err
	}

	// when the max_supply is set to zero, the supply is unlimited
	if token.MaxSupply.IsZero() {
		return false, sdkmath.ZeroInt(), nil
	}

	total, err := k.leverageKeeper.GetTotalSupply(*k.ctx, denom)
	if err != nil {
		return true, sdkmath.Int{}, err
	}

	return true, token.MaxSupply.Sub(total.Amount), nil
}

// breakFee calculates the protocol fee for the burn auction and the reminder
func (k Keeper) breakFee(fee sdkmath.Int) (toAuction sdkmath.Int, revenue sdkmath.Int) {
	p := k.GetParams()
	toAuction = p.RewardsAuctionFactor.Mul(fee)
	return toAuction, fee.Sub(toAuction)
}

// calculateSwap returns the amount of meToken to send to the user, the fee to be charged to him,
// the amount of assets to send to x/metoken reserves and to x/leverage pools.
// The formulas used for the calculations are:
//
//	assets_to_swap = assets_from_user - fee
//	metokens_to_mint = assets_to_swap * exchange_rate
//	amount_to_reserves = assets_to_swap * reserve_portion
//	amount_to_leverage = assets_to_swap - amount_to_reserves
//
// It returns calculated amount of tokens that need to be transferred as a result of a transfer.
func (k Keeper) calculateSwap(index metoken.Index, indexPrices metoken.IndexPrices, asset sdk.Coin) (
	swapCarry, error,
) {
	assetSettings, i := index.AcceptedAsset(asset.Denom)
	if i < 0 {
		return swapCarry{}, sdkerrors.ErrNotFound.Wrapf(
			"asset %s is not accepted in the index",
			asset.Denom,
		)
	}

	_, fee, err := k.swapFee(index, indexPrices, asset)
	if err != nil {
		return swapCarry{}, err
	}

	amountToSwap := asset.Amount.Sub(fee.Amount)
	meTokens, err := indexPrices.SwapRate(sdk.NewCoin(asset.Denom, amountToSwap))
	if err != nil {
		return swapCarry{}, err
	}

	toReserves := assetSettings.ReservePortion.MulInt(amountToSwap).TruncateInt()
	toLeverage := amountToSwap.Sub(toReserves)

	return swapCarry{meTokens: meTokens, fee: fee.Amount, toReserves: toReserves, toLeverage: toLeverage}, nil
}

func (k Keeper) fundAuction(denom string, amount sdkmath.Int) error {
	if amount.IsZero() {
		return nil
	}
	coins := sdk.Coins{sdk.NewCoin(denom, amount)}
	auction.EmitFundRewardsAuction(k.ctx, coins)
	return k.bankKeeper.SendCoinsFromModuleToModule(*k.ctx, metoken.ModuleName, auction.ModuleName, coins)
}

type swapCarry struct {
	meTokens   sdkmath.Int
	fee        sdkmath.Int
	toReserves sdkmath.Int
	toLeverage sdkmath.Int
}
