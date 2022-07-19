package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

type Keeper struct {
	cdc          codec.Codec
	storeKey     sdk.StoreKey
	paramSpace   paramtypes.Subspace
	hooks        types.Hooks
	bankKeeper   types.BankKeeper
	oracleKeeper types.OracleKeeper

	tokenRegCache simplelru.LRUCache
}

func NewKeeper(
	cdc codec.Codec,
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	bk types.BankKeeper,
	ok types.OracleKeeper,
) (Keeper, error) {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	const tokenRegCacheSize = 100
	tokenRegCache, err := simplelru.NewLRU(tokenRegCacheSize, nil)
	if err != nil {
		return Keeper{}, err
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramSpace:    paramSpace,
		bankKeeper:    bk,
		oracleKeeper:  ok,
		tokenRegCache: tokenRegCache,
	}, nil
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetHooks sets the module's hooks. Note, hooks can only be set once.
func (k *Keeper) SetHooks(h types.Hooks) *Keeper {
	if k.hooks != nil {
		panic("leverage hooks already set")
	}

	k.hooks = h

	return k
}

// ModuleBalance returns the amount of a given token held in the x/leverage module account
func (k Keeper) ModuleBalance(ctx sdk.Context, denom string) sdk.Int {
	return k.bankKeeper.SpendableCoins(ctx, authtypes.NewModuleAddress(types.ModuleName)).AmountOf(denom)
}

// Supply attempts to deposit assets into the leverage module account in
// exchange for uTokens. If asset type is invalid or account balance is
// insufficient, we return an error.
func (k Keeper) Supply(ctx sdk.Context, supplierAddr sdk.AccAddress, loan sdk.Coin) error {
	if err := k.validateSupply(ctx, loan); err != nil {
		return err
	}

	// determine uToken amount to mint
	uToken, err := k.ExchangeToken(ctx, loan)
	if err != nil {
		return err
	}

	// send token balance to leverage module account
	loanTokens := sdk.NewCoins(loan)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, supplierAddr, types.ModuleName, loanTokens); err != nil {
		return err
	}

	// mint uToken and set new total uToken supply
	uTokens := sdk.NewCoins(uToken)
	if err = k.bankKeeper.MintCoins(ctx, types.ModuleName, uTokens); err != nil {
		return err
	}
	if err = k.setUTokenSupply(ctx, k.GetUTokenSupply(ctx, uToken.Denom).Add(uToken)); err != nil {
		return err
	}

	// The uTokens are sent to supplier address
	if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, supplierAddr, uTokens); err != nil {
		return err
	}

	return nil
}

// Withdraw attempts to deposit uTokens into the leverage module in exchange
// for the original tokens supplied. Accepts a uToken amount to exchange for base tokens.
// If the uToken denom is invalid or account or module balance insufficient, returns error.
func (k Keeper) Withdraw(ctx sdk.Context, supplierAddr sdk.AccAddress, coin sdk.Coin) error {
	if !k.IsAcceptedUToken(ctx, coin.Denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, coin.String())
	}

	// calculate base asset amount to withdraw
	token, err := k.ExchangeUToken(ctx, coin)
	if err != nil {
		return err
	}

	// Ensure module account has sufficient unreserved tokens to withdraw
	reservedAmount := k.GetReserveAmount(ctx, token.Denom)
	availableAmount := k.ModuleBalance(ctx, token.Denom)
	if token.Amount.GT(availableAmount.Sub(reservedAmount)) {
		return sdkerrors.Wrap(types.ErrLendingPoolInsufficient, token.String())
	}

	// Withdraw will first attempt to use any uTokens in the supplier's wallet
	amountFromWallet := sdk.MinInt(k.bankKeeper.SpendableCoins(ctx, supplierAddr).AmountOf(coin.Denom), coin.Amount)
	// Any additional uTokens must come from the supplier's collateral
	amountFromCollateral := coin.Amount.Sub(amountFromWallet)

	if amountFromCollateral.IsPositive() {
		// Calculate current borrowed value
		borrowed := k.GetBorrowerBorrows(ctx, supplierAddr)
		borrowedValue, err := k.TotalTokenValue(ctx, borrowed)
		if err != nil {
			return err
		}

		// Check for sufficient collateral
		collateral := k.GetBorrowerCollateral(ctx, supplierAddr)
		if collateral.AmountOf(coin.Denom).LT(amountFromCollateral) {
			return sdkerrors.Wrap(types.ErrInsufficientBalance, coin.String())
		}

		// Calculate what borrow limit will be AFTER this withdrawal
		collateralToWithdraw := sdk.NewCoins(sdk.NewCoin(coin.Denom, amountFromCollateral))
		newBorrowLimit, err := k.CalculateBorrowLimit(ctx, collateral.Sub(collateralToWithdraw))
		if err != nil {
			return err
		}

		// Return error if borrow limit would drop below borrowed value
		if borrowedValue.GT(newBorrowLimit) {
			return types.ErrUndercollaterized.Wrapf(
				"withdraw would decrease borrow limit to %s, below the current borrowed value %s", newBorrowLimit, borrowedValue)
		}

		// reduce the supplier's collateral by amountFromCollateral
		newCollateral := sdk.NewCoin(coin.Denom, collateral.AmountOf(coin.Denom).Sub(amountFromCollateral))
		if err = k.setCollateralAmount(ctx, supplierAddr, newCollateral); err != nil {
			return err
		}
	}

	// transfer amountFromWallet uTokens to the module account
	uTokens := sdk.NewCoins(sdk.NewCoin(coin.Denom, amountFromWallet))
	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, supplierAddr, types.ModuleName, uTokens); err != nil {
		return err
	}

	// send the base assets to supplier
	tokens := sdk.NewCoins(token)
	if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, supplierAddr, tokens); err != nil {
		return err
	}

	// burn the uTokens and set the new total uToken supply
	if err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		return err
	}
	if err = k.setUTokenSupply(ctx, k.GetUTokenSupply(ctx, coin.Denom).Sub(coin)); err != nil {
		return err
	}

	return nil
}

// Borrow attempts to borrow tokens from the leverage module account using
// collateral uTokens. If asset type is invalid, collateral is insufficient,
// or module balance is insufficient, we return an error.
func (k Keeper) Borrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, borrow sdk.Coin) error {
	if err := k.validateBorrow(ctx, borrow); err != nil {
		return err
	}

	token, err := k.GetTokenSettings(ctx, borrow.Denom)
	if err != nil {
		return err
	}

	su := k.supplyUtilizationAfterBorrow(ctx, borrow)
	if su.GT(sdk.OneDec()) {
		// Ensure module account has sufficient unreserved tokens to loan out
		return types.ErrLendingPoolInsufficient.Wrap(borrow.String())
	}
	if su.GT(token.MaxSupplyUtilization) {
		// Ensure max supply utilization is not exceeded
		return types.ErrMaxSupplyUtilization
	}

	// Determine amount of all tokens currently borrowed
	borrowed := k.GetBorrowerBorrows(ctx, borrowerAddr)

	// Calculate current borrow limit
	collateral := k.GetBorrowerCollateral(ctx, borrowerAddr)
	borrowLimit, err := k.CalculateBorrowLimit(ctx, collateral)
	if err != nil {
		return err
	}

	// Calculate borrowed value will be AFTER this borrow
	newBorrowedValue, err := k.TotalTokenValue(ctx, borrowed.Add(borrow))
	if err != nil {
		return err
	}

	// Return error if borrowed value would exceed borrow limit
	if newBorrowedValue.GT(borrowLimit) {
		return types.ErrUndercollaterized.Wrapf("new borrowed value would be %s with borrow limit %s",
			newBorrowedValue, borrowLimit)
	}

	loanTokens := sdk.NewCoins(borrow)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrowerAddr, loanTokens); err != nil {
		return err
	}

	// Determine the total amount of denom borrowed (previously borrowed + newly borrowed)
	newBorrow := borrowed.AmountOf(borrow.Denom).Add(borrow.Amount)
	if err := k.setBorrow(ctx, borrowerAddr, sdk.NewCoin(borrow.Denom, newBorrow)); err != nil {
		return err
	}
	return nil
}

// Repay attempts to repay a borrow position. If asset type is invalid, account balance
// is insufficient, or borrower has no borrows in payment denom to repay, we return an error.
// Additionally, if the amount provided is greater than the full repayment amount, only the
// necessary amount is transferred. Because amount repaid may be less than the repayment attempted,
// Repay returns the actual amount repaid.
func (k Keeper) Repay(ctx sdk.Context, borrowerAddr sdk.AccAddress, payment sdk.Coin) (sdk.Int, error) {
	if !payment.IsValid() {
		return sdk.ZeroInt(), types.ErrInvalidAsset.Wrap(payment.String())
	}

	// Determine amount of selected denom currently owed
	owed := k.GetBorrow(ctx, borrowerAddr, payment.Denom)
	if owed.IsZero() {
		// Borrower has no open borrows in the denom presented as payment
		return sdk.ZeroInt(), types.ErrInvalidRepayment.Wrap(
			"Borrower doesn't have active position in " + payment.Denom)
	}

	// Prevent overpaying
	payment.Amount = sdk.MinInt(owed.Amount, payment.Amount)
	if err := payment.Validate(); err != nil {
		return sdk.ZeroInt(), types.ErrInvalidRepayment.Wrap(err.Error())
	}

	// send payment to leverage module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, borrowerAddr,
		types.ModuleName,
		sdk.NewCoins(payment),
	); err != nil {
		return sdk.ZeroInt(), err
	}

	owed.Amount = owed.Amount.Sub(payment.Amount)
	if err := k.setBorrow(ctx, borrowerAddr, owed); err != nil {
		return sdk.ZeroInt(), err
	}
	return payment.Amount, nil
}

// Collateralize enables selected uTokens for use as collateral by a single borrower.
func (k Keeper) Collateralize(ctx sdk.Context, borrowerAddr sdk.AccAddress, coin sdk.Coin) error {
	if err := k.validateCollateralAsset(ctx, coin); err != nil {
		return err
	}

	currentCollateral := k.GetCollateralAmount(ctx, borrowerAddr, coin.Denom)
	if err := k.setCollateralAmount(ctx, borrowerAddr, currentCollateral.Add(coin)); err != nil {
		return err
	}

	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, borrowerAddr, types.ModuleName, sdk.NewCoins(coin))
	if err != nil {
		return err
	}

	return nil
}

// Decollateralize disables selected uTokens for use as collateral by a single borrower.
func (k Keeper) Decollateralize(ctx sdk.Context, borrowerAddr sdk.AccAddress, coin sdk.Coin) error {
	if err := coin.Validate(); err != nil {
		return err
	}

	// Detect where sufficient collateral exists to disable
	collateral := k.GetBorrowerCollateral(ctx, borrowerAddr)
	if collateral.AmountOf(coin.Denom).LT(coin.Amount) {
		return types.ErrInsufficientBalance
	}

	// Determine what borrow limit would be AFTER disabling this denom as collateral
	newBorrowLimit, err := k.CalculateBorrowLimit(ctx, collateral.Sub(sdk.NewCoins(coin)))
	if err != nil {
		return err
	}

	// Determine currently borrowed value
	borrowed := k.GetBorrowerBorrows(ctx, borrowerAddr)
	borrowedValue, err := k.TotalTokenValue(ctx, borrowed)
	if err != nil {
		return err
	}

	// Return error if borrow limit would drop below borrowed value
	if newBorrowLimit.LT(borrowedValue) {
		return types.ErrUndercollaterized.Wrap("new borrow limit: " + newBorrowLimit.String())
	}

	// Disabling uTokens as collateral withdraws any stored collateral of the denom in question
	// from the module account and returns it to the user
	newCollateralAmount := collateral.AmountOf(coin.Denom).Sub(coin.Amount)
	if err := k.setCollateralAmount(ctx, borrowerAddr, sdk.NewCoin(coin.Denom, newCollateralAmount)); err != nil {
		return err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrowerAddr, sdk.NewCoins(coin))
	if err != nil {
		return err
	}

	return nil
}

// LiquidateBorrow attempts to repay one of an eligible borrower's borrows (in part or in full) in exchange
// for a selected denomination of uToken collateral, specified by its associated token denom. The liquidator
// may also specify a minimum reward amount, again in base token denom that will be adjusted by uToken exchange
// rate, they would accept for the specified repayment. If the borrower is not over their liquidation limit, or
// the repayment or reward denominations are invalid, an error is returned. If the attempted repayment
// is greater than the amount owed or the maximum that can be repaid due to parameters (close factor)
// then a partial liquidation, equal to the maximum valid amount, is performed. The same occurs if the
// value of collateral in the selected reward denomination cannot cover the proposed repayment.
// Because partial liquidation is possible and exchange rates vary, LiquidateBorrow returns the actual
// amount of tokens repaid and uTokens rewarded (in that order).
func (k Keeper) LiquidateBorrow(
	ctx sdk.Context, liquidatorAddr, borrowerAddr sdk.AccAddress, desiredRepayment, desiredReward sdk.Coin,
) (sdk.Int, sdk.Int, error) {
	if !desiredRepayment.IsValid() {
		return sdk.ZeroInt(), sdk.ZeroInt(), types.ErrInvalidAsset.Wrap(desiredRepayment.String())
	}
	if err := k.AssertNotBlacklisted(ctx, desiredRepayment.Denom); err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}
	if !k.IsAcceptedToken(ctx, desiredReward.Denom) {
		return sdk.ZeroInt(), sdk.ZeroInt(), types.ErrInvalidAsset.Wrap(desiredReward.String())
	}

	collateral := k.GetBorrowerCollateral(ctx, borrowerAddr)
	borrowed := k.GetBorrowerBorrows(ctx, borrowerAddr)

	borrowValue, err := k.TotalTokenValue(ctx, borrowed) // total borrowed value in USD
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	liquidationThreshold, err := k.CalculateLiquidationThreshold(ctx, collateral)
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	// confirm borrower's eligibility for liquidation
	if liquidationThreshold.GTE(borrowValue) {
		return sdk.ZeroInt(), sdk.ZeroInt(), types.ErrLiquidationIneligible.Wrapf(
			"%s borrowed value is below the liquidation threshold %s", borrowerAddr, liquidationThreshold)
	}

	// get reward-specific incentive and dynamic close factor
	baseRewardDenom := desiredReward.Denom
	liquidationIncentive, closeFactor, err := k.LiquidationParams(ctx, baseRewardDenom, borrowValue, liquidationThreshold)
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	// actual repayment starts at desiredRepayment but can be lower due to limiting factors
	repayment := desiredRepayment

	// get liquidator's available balance of base asset to repay
	liquidatorBalance := k.bankKeeper.SpendableCoins(ctx, liquidatorAddr).AmountOf(repayment.Denom)

	// repayment cannot exceed liquidator's available balance
	repayment.Amount = sdk.MinInt(repayment.Amount, liquidatorBalance)

	// repayment cannot exceed borrower's borrowed amount of selected denom
	repayment.Amount = sdk.MinInt(repayment.Amount, borrowed.AmountOf(repayment.Denom))

	// repayment cannot exceed borrowed value * close factor
	maxRepayValue := borrowValue.Mul(closeFactor)
	repayValue, err := k.TokenValue(ctx, repayment)
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	if repayValue.GT(maxRepayValue) {
		// repayment *= (maxRepayValue / repayValue)
		repayment.Amount = repayment.Amount.ToDec().Mul(maxRepayValue).Quo(repayValue).TruncateInt()
	}

	// Given repay denom and amount, use oracle to find equivalent amount of rewardDenom.
	baseReward, err := k.EquivalentTokenValue(ctx, repayment, baseRewardDenom)
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	// convert reward tokens back to uTokens
	reward, err := k.ExchangeToken(ctx, baseReward)
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	// apply liquidation incentive
	reward.Amount = reward.Amount.ToDec().Mul(sdk.OneDec().Add(liquidationIncentive)).TruncateInt()

	maxReward := collateral.AmountOf(reward.Denom)
	if maxReward.IsZero() {
		return sdk.ZeroInt(), sdk.ZeroInt(), types.ErrInvalidAsset.Wrapf(
			"borrower doesn't have %s as a collateral", desiredReward.Denom)
	}

	// reward amount cannot exceed available collateral
	if reward.Amount.GT(maxReward) {
		// reduce repayment.Amount to the maximum value permitted by the available collateral reward
		repayment.Amount = repayment.Amount.Mul(maxReward).Quo(reward.Amount)
		reward.Amount = maxReward
	}

	// final check for invalid liquidation (negative/zero value after reductions above)
	if !repayment.Amount.IsPositive() {
		return sdk.ZeroInt(), sdk.ZeroInt(), types.ErrInvalidAsset.Wrap(repayment.String())
	}

	if desiredReward.Amount.IsPositive() {
		// user-controlled minimum ratio of reward to repayment, expressed in collateral base assets (not uTokens)
		rewardTokenEquivalent, err := k.ExchangeUToken(ctx, reward)
		if err != nil {
			return sdk.ZeroInt(), sdk.ZeroInt(), err
		}

		minimumRewardRatio := sdk.NewDecFromInt(desiredReward.Amount).QuoInt(desiredRepayment.Amount)
		actualRewardRatio := sdk.NewDecFromInt(rewardTokenEquivalent.Amount).QuoInt(repayment.Amount)
		if actualRewardRatio.LT(minimumRewardRatio) {
			return sdk.ZeroInt(), sdk.ZeroInt(), types.ErrLiquidationRewardRatio
		}
	}

	// send repayment to leverage module account
	if err = k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, liquidatorAddr,
		types.ModuleName,
		sdk.NewCoins(repayment),
	); err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	// update the remaining borrowed amount
	owed := borrowed.AmountOf(repayment.Denom).Sub(repayment.Amount)
	if err = k.setBorrow(ctx, borrowerAddr, sdk.NewCoin(repayment.Denom, owed)); err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	// Reduce borrower collateral by reward amount
	newBorrowerCollateral := sdk.NewCoin(reward.Denom, maxReward.Sub(reward.Amount))
	if err = k.setCollateralAmount(ctx, borrowerAddr, newBorrowerCollateral); err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	// Send rewards to liquidator's account.
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidatorAddr, sdk.NewCoins(reward))
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	// Detect bad debt (collateral == 0 after reward) for repayment by protocol reserves (see ADR-004)
	if collateral.Sub(sdk.NewCoins(reward)).IsZero() {
		for _, coin := range borrowed {
			// Mark repayment denom as bad debt only if some debt remains after
			// this liquidation. All other borrowed denoms were definitely not
			// repaid in this liquidation so they are always marked as bad debt.
			if coin.Denom != repayment.Denom || owed.IsPositive() {
				if err := k.setBadDebtAddress(ctx, borrowerAddr, coin.Denom, true); err != nil {
					return sdk.ZeroInt(), sdk.ZeroInt(), err
				}
			}
		}
	}

	return repayment.Amount, reward.Amount, nil
}

// LiquidationParams computes dynamic liquidation parameters based on collateral denomination,
// borrowed value, and liquidation threshold. Returns liquidationIncentive (the ratio of bonus collateral
// awarded during Liquidate transactions, and closeFactor (the fraction of a borrower's total
// borrowed value that can be repaid by a liquidator in a single liquidation event.)
func (k Keeper) LiquidationParams(
	ctx sdk.Context,
	rewardDenom string,
	borrowed sdk.Dec,
	limit sdk.Dec,
) (sdk.Dec, sdk.Dec, error) {
	if borrowed.IsNegative() {
		return sdk.ZeroDec(), sdk.ZeroDec(), sdkerrors.Wrap(types.ErrBadValue, borrowed.String())
	}
	if limit.IsNegative() {
		return sdk.ZeroDec(), sdk.ZeroDec(), sdkerrors.Wrap(types.ErrBadValue, limit.String())
	}

	ts, err := k.GetTokenSettings(ctx, rewardDenom)
	if err != nil {
		return sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// special case: If liquidation threshold is zero, close factor is always 1
	if limit.IsZero() {
		return ts.LiquidationIncentive, sdk.OneDec(), nil
	}

	params := k.GetParams(ctx)

	// special case: If borrowed value is less than small liquidation size,
	// close factor is always 1
	if borrowed.LTE(params.SmallLiquidationSize) {
		return ts.LiquidationIncentive, sdk.OneDec(), nil
	}

	// special case: If complete liquidation threshold is zero, close factor is always 1
	if params.CompleteLiquidationThreshold.IsZero() {
		return ts.LiquidationIncentive, sdk.OneDec(), nil
	}

	// outside of special cases, close factor scales linearly between MinimumCloseFactor and 1.0,
	// reaching max value when (borrowed / threshold) = 1 + CompleteLiquidationThreshold
	var closeFactor sdk.Dec
	closeFactor = Interpolate(
		borrowed.Quo(limit).Sub(sdk.OneDec()), // x
		sdk.ZeroDec(),                         // xMin
		params.MinimumCloseFactor,             // yMin
		params.CompleteLiquidationThreshold,   // xMax
		sdk.OneDec(),                          // yMax
	)
	if closeFactor.GTE(sdk.OneDec()) {
		closeFactor = sdk.OneDec()
	}
	if closeFactor.IsNegative() {
		closeFactor = sdk.ZeroDec()
	}

	return ts.LiquidationIncentive, closeFactor, nil
}
