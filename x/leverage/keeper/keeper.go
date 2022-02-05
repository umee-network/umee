package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/x/leverage/types"
)

type Keeper struct {
	cdc          codec.Codec
	storeKey     sdk.StoreKey
	paramSpace   paramtypes.Subspace
	hooks        types.Hooks
	bankKeeper   types.BankKeeper
	oracleKeeper types.OracleKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	bk types.BankKeeper,
	ok types.OracleKeeper,
) Keeper {

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		paramSpace:   paramSpace,
		bankKeeper:   bk,
		oracleKeeper: ok,
	}
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

// TotalUTokenSupply returns an sdk.Coin representing the total balance of a
// given uToken type if valid. If the denom is not an accepted uToken type,
// we return a zero amount.
func (k Keeper) TotalUTokenSupply(ctx sdk.Context, uTokenDenom string) sdk.Coin {
	if k.IsAcceptedUToken(ctx, uTokenDenom) {
		return k.bankKeeper.GetSupply(ctx, uTokenDenom)
		// TODO - Question: Does bank module still track balances sent (locked) via IBC? If it doesn't
		// then the balance returned here would decrease when the tokens are sent off, which is not
		// what we want. In that case, the keeper should keep an sdk.Int total supply for each uToken type.
	}
	return sdk.NewCoin(uTokenDenom, sdk.ZeroInt())
}

// ModuleBalance returns the amount of a given token held in the x/leverage module account
func (k Keeper) ModuleBalance(ctx sdk.Context, denom string) sdk.Int {
	return k.bankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), denom).Amount
}

// LendAsset attempts to deposit assets into the leverage module account in
// exchange for uTokens. If asset type is invalid or account balance is
// insufficient, we return an error.
func (k Keeper) LendAsset(ctx sdk.Context, lenderAddr sdk.AccAddress, loan sdk.Coin) error {
	if !k.IsAcceptedToken(ctx, loan.Denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, loan.String())
	}

	// determine uToken amount to mint
	uToken, err := k.ExchangeToken(ctx, loan)
	if err != nil {
		return err
	}

	// send token balance to leverage module account
	loanTokens := sdk.NewCoins(loan)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, lenderAddr, types.ModuleName, loanTokens); err != nil {
		return err
	}

	// mint uToken
	uTokens := sdk.NewCoins(uToken)
	if err = k.bankKeeper.MintCoins(ctx, types.ModuleName, uTokens); err != nil {
		return err
	}

	if k.GetCollateralSetting(ctx, lenderAddr, uToken.Denom) {
		// For uToken denoms enabled as collateral by this lender, the
		// minted uTokens stay in the module account and the keeper tracks the amount.
		currentCollateral := k.GetCollateralAmount(ctx, lenderAddr, uToken.Denom)
		if err = k.SetCollateralAmount(ctx, lenderAddr, currentCollateral.Add(uToken)); err != nil {
			return err
		}
	} else if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lenderAddr, uTokens); err != nil {
		// For uToken denoms not enabled as collateral by this lender, the uTokens are sent to lender address
		return err
	}

	return nil
}

// WithdrawAsset attempts to deposit uTokens into the leverage module in exchange
// for the original tokens lent. If the uToken type is invalid or account balance
// insufficient on either side, we return an error.
func (k Keeper) WithdrawAsset(ctx sdk.Context, lenderAddr sdk.AccAddress, uToken sdk.Coin) error {
	if !k.IsAcceptedUToken(ctx, uToken.Denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, uToken.String())
	}

	// calculate base asset amount to withdraw
	token, err := k.ExchangeUToken(ctx, uToken)
	if err != nil {
		return err
	}

	// Ensure module account has sufficient unreserved tokens to withdraw
	reservedAmount := k.GetReserveAmount(ctx, token.Denom)
	availableAmount := k.ModuleBalance(ctx, token.Denom)
	if token.Amount.GT(availableAmount.Sub(reservedAmount)) {
		return sdkerrors.Wrap(types.ErrLendingPoolInsufficient, token.String())
	}

	// Withdraw will first attempt to use any uTokens in the lender's wallet
	amountFromWallet := sdk.MinInt(k.bankKeeper.SpendableCoins(ctx, lenderAddr).AmountOf(uToken.Denom), uToken.Amount)
	// Any additional uTokens must come from the lender's collateral
	amountFromCollateral := uToken.Amount.Sub(amountFromWallet)

	if amountFromCollateral.IsPositive() {
		if k.GetCollateralSetting(ctx, lenderAddr, uToken.Denom) {
			// Calculate current borrowed value
			borrowed := k.GetBorrowerBorrows(ctx, lenderAddr)
			borrowedValue, err := k.TotalTokenValue(ctx, borrowed)
			if err != nil {
				return err
			}

			// Check for sufficient collateral
			collateral := k.GetBorrowerCollateral(ctx, lenderAddr)
			if collateral.AmountOf(uToken.Denom).LT(amountFromCollateral) {
				return sdkerrors.Wrap(types.ErrInsufficientBalance, uToken.String())
			}

			// Calculate what borrow limit will be AFTER this withdrawal
			collateralToWithdraw := sdk.NewCoins(sdk.NewCoin(uToken.Denom, amountFromCollateral))
			newBorrowLimit, err := k.CalculateBorrowLimit(ctx, collateral.Sub(collateralToWithdraw))
			if err != nil {
				return err
			}

			// Return error if borrow limit would drop below borrowed value
			if borrowedValue.GT(newBorrowLimit) {
				return sdkerrors.Wrap(types.ErrBorrowLimitLow, newBorrowLimit.String())
			}

			// reduce the lender's collateral by amountFromCollateral
			newCollateral := sdk.NewCoin(uToken.Denom, collateral.AmountOf(uToken.Denom).Sub(amountFromCollateral))
			if err = k.SetCollateralAmount(ctx, lenderAddr, newCollateral); err != nil {
				return err
			}
		} else {
			// If collateral was needed despite being disabled, wallet balance must have been insufficient
			return sdkerrors.Wrap(types.ErrInsufficientBalance, uToken.String())
		}
	}

	// transfer amountFromWallet uTokens to the module account
	uTokens := sdk.NewCoins(sdk.NewCoin(uToken.Denom, amountFromWallet))
	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, lenderAddr, types.ModuleName, uTokens); err != nil {
		return err
	}

	// send the base assets to lender
	tokens := sdk.NewCoins(token)
	if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lenderAddr, tokens); err != nil {
		return err
	}

	// burn the uTokens
	if err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(uToken)); err != nil {
		return err
	}

	return nil
}

// BorrowAsset attempts to borrow tokens from the leverage module account using
// collateral uTokens. If asset type is invalid, collateral is insufficient,
// or module balance is insufficient, we return an error.
func (k Keeper) BorrowAsset(ctx sdk.Context, borrowerAddr sdk.AccAddress, borrow sdk.Coin) error {
	if !borrow.IsValid() {
		return sdkerrors.Wrap(types.ErrInvalidAsset, borrow.String())
	}

	if !k.IsAcceptedToken(ctx, borrow.Denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, borrow.String())
	}

	// Ensure module account has sufficient unreserved tokens to loan out
	reservedAmount := k.GetReserveAmount(ctx, borrow.Denom)
	availableAmount := k.bankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), borrow.Denom).Amount
	if borrow.Amount.GT(availableAmount.Sub(reservedAmount)) {
		return sdkerrors.Wrap(types.ErrLendingPoolInsufficient, borrow.String())
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
		return sdkerrors.Wrap(types.ErrBorrowLimitLow, borrowLimit.String())
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

// RepayAsset attempts to repay an open borrow position with base assets. If asset type is invalid,
// account balance is insufficient, or no open borrow position exists, we return an error.
// Additionally, if the amount provided is greater than the full repayment amount, only the
// necessary amount is transferred. Because amount repaid may be less than the repayment attempted,
// RepayAsset returns the actual amount repaid.
func (k Keeper) RepayAsset(ctx sdk.Context, borrowerAddr sdk.AccAddress, payment sdk.Coin) (sdk.Int, error) {
	if !payment.IsValid() {
		return sdk.ZeroInt(), sdkerrors.Wrap(types.ErrInvalidAsset, payment.String())
	}

	if !k.IsAcceptedToken(ctx, payment.Denom) {
		return sdk.ZeroInt(), sdkerrors.Wrap(types.ErrInvalidAsset, payment.String())
	}

	// Determine amount of selected denom currently owed
	owed := k.GetBorrow(ctx, borrowerAddr, payment.Denom)
	if owed.IsZero() {
		// Borrower has no open borrows in the denom presented as payment
		return sdk.ZeroInt(), sdkerrors.Wrap(types.ErrInvalidRepayment, payment.String())
	}

	// Prevent overpaying
	payment.Amount = sdk.MinInt(owed.Amount, payment.Amount)
	if !payment.IsValid() {
		// Catch invalid payments (e.g. from payment.Amount < 0)
		return sdk.ZeroInt(), sdkerrors.Wrap(types.ErrInvalidRepayment, payment.String())
	}

	// send payment to leverage module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, borrowerAddr,
		types.ModuleName,
		sdk.NewCoins(payment),
	); err != nil {
		return sdk.ZeroInt(), err
	}

	// Subtract repaid amount from borrowed amount
	owed.Amount = owed.Amount.Sub(payment.Amount)

	// Store the remaining borrowed amount in keeper
	if err := k.setBorrow(ctx, borrowerAddr, owed); err != nil {
		return sdk.ZeroInt(), err
	}
	return payment.Amount, nil
}

// SetCollateralSetting enables or disables a uToken denom for use as collateral by a single borrower.
func (k Keeper) SetCollateralSetting(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string, enable bool) error {
	if !k.IsAcceptedUToken(ctx, denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	if enable {
		// Enabling a denom of uTokens as collateral deposits any in the user's current
		// balance into the module account and remembers the amount held.
		uToken := sdk.NewCoin(denom, k.bankKeeper.SpendableCoins(ctx, borrowerAddr).AmountOf(denom))
		uTokens := sdk.NewCoins(uToken)

		if uToken.Amount.IsPositive() {
			currentCollateral := k.GetCollateralAmount(ctx, borrowerAddr, uToken.Denom)
			if err := k.SetCollateralAmount(ctx, borrowerAddr, currentCollateral.Add(uToken)); err != nil {
				return err
			}

			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, borrowerAddr, types.ModuleName, uTokens); err != nil {
				return err
			}
		}
	} else {
		// Determine currently borrowed value
		borrowed := k.GetBorrowerBorrows(ctx, borrowerAddr)
		borrowedValue, err := k.TotalTokenValue(ctx, borrowed)
		if err != nil {
			return err
		}

		// Determine what borrow limit would be AFTER disabling this denom as collateral
		collateral := k.GetBorrowerCollateral(ctx, borrowerAddr)
		collateralToDisable := sdk.NewCoins(sdk.NewCoin(denom, collateral.AmountOf(denom)))
		newBorrowLimit, err := k.CalculateBorrowLimit(ctx, collateral.Sub(collateralToDisable))
		if err != nil {
			return err
		}

		// Return error if borrow limit would drop below borrowed value
		if newBorrowLimit.LT(borrowedValue) {
			return sdkerrors.Wrap(types.ErrBorrowLimitLow, newBorrowLimit.String())
		}

		// Disabling uTokens as collateral withdraws any stored collateral of the denom in question
		// from the module account and returns it to the user
		currentCollateral := k.GetCollateralAmount(ctx, borrowerAddr, denom)
		uTokens := sdk.NewCoins(currentCollateral)

		if currentCollateral.IsPositive() {
			if err := k.SetCollateralAmount(ctx, borrowerAddr, sdk.NewCoin(denom, sdk.ZeroInt())); err != nil {
				return err
			}
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrowerAddr, uTokens); err != nil {
				return err
			}
		}
	}

	return k.setCollateralSetting(ctx, borrowerAddr, denom, enable)
}

// GetCollateralSetting checks if a uToken denom is enabled for use as collateral by a single borrower.
func (k Keeper) GetCollateralSetting(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) bool {
	if !k.IsAcceptedUToken(ctx, denom) {
		return false
	}
	store := ctx.KVStore(k.storeKey)
	// Any value (expected = 0x01) found at key will be interpreted as true.
	key := types.CreateCollateralSettingKey(borrowerAddr, denom)
	return store.Has(key)
}

// LiquidateBorrow attempts to repay one of an eligible borrower's borrows (in part or in full) in exchange
// for a selected denomination of uToken collateral. If the borrower is not over their borrow limit, or
// the repayment or reward denominations are invalid, an error is returned. If the attempted repayment
// is greater than the amount owed or the maximum that can be repaid due to parameters (close factor)
// then a partial liquidation, equal to the maximum valid amount, is performed. The same occurs if the
// value of collateral in the selected reward denomination cannot cover the proposed repayment.
// Because partial liquidation is possible and exchange rates vary, LiquidateBorrow returns the actual
// amount of tokens repaid and uTokens rewarded (in that order).
func (k Keeper) LiquidateBorrow(
	ctx sdk.Context, liquidatorAddr, borrowerAddr sdk.AccAddress, desiredRepayment sdk.Coin, rewardDenom string,
) (sdk.Int, sdk.Int, error) {
	if !desiredRepayment.IsValid() {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdkerrors.Wrap(types.ErrInvalidAsset, desiredRepayment.String())
	}
	if !k.IsAcceptedUToken(ctx, rewardDenom) {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdkerrors.Wrap(types.ErrInvalidAsset, rewardDenom)
	}

	// get total borrowed by borrower (all denoms)
	borrowed := k.GetBorrowerBorrows(ctx, borrowerAddr)

	// get borrower uToken balances, for all uToken denoms enabled as collateral
	collateral := k.GetBorrowerCollateral(ctx, borrowerAddr)

	// use oracle helper functions to find total borrowed value in USD
	borrowValue, err := k.TotalTokenValue(ctx, borrowed)
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	// use collateral weights to compute borrow limit from enabled collateral
	borrowLimit, err := k.CalculateBorrowLimit(ctx, collateral)
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	// confirm borrower's eligibility for liquidation
	if borrowLimit.GTE(borrowValue) {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdkerrors.Wrap(types.ErrLiquidationIneligible, borrowerAddr.String())
	}

	// get reward-specific incentive and dynamic close factor
	baseRewardDenom := k.FromUTokenToTokenDenom(ctx, rewardDenom)
	liquidationIncentive, closeFactor, err := k.LiquidationParams(ctx, baseRewardDenom, borrowValue, borrowLimit)
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

	if repayValue.GTE(maxRepayValue) {
		// repayment *= (maxRepayValue / repayValue)
		repayment.Amount = repayment.Amount.ToDec().Mul(maxRepayValue).Quo(repayValue).TruncateInt()
	}

	// Given repay denom and amount, use oracle to find equivalent amount of
	// rewardDenom's base asset.
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

	// reward amount cannot exceed available collateral
	if reward.Amount.GTE(collateral.AmountOf(rewardDenom)) {
		// reduce repayment.Amount to the maximum value permitted by the available collateral reward
		repayment.Amount = repayment.Amount.Mul(collateral.AmountOf(rewardDenom)).Quo(reward.Amount)
		// use all collateral of reward denom
		reward.Amount = collateral.AmountOf(rewardDenom)
	}

	// final check for invalid liquidation (negative/zero value after reductions above)
	if !repayment.Amount.IsPositive() {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdkerrors.Wrap(types.ErrInvalidAsset, repayment.String())
	}

	// send repayment to leverage module account
	if err = k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, liquidatorAddr,
		types.ModuleName,
		sdk.NewCoins(repayment),
	); err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	// store the remaining borrowed amount in keeper
	owed := borrowed.AmountOf(repayment.Denom).Sub(repayment.Amount)
	if err = k.setBorrow(ctx, borrowerAddr, sdk.NewCoin(repayment.Denom, owed)); err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	// Reduce borrower collateral by reward amount
	newBorrowerCollateral := sdk.NewCoin(rewardDenom, collateral.AmountOf(rewardDenom).Sub(reward.Amount))
	if err = k.SetCollateralAmount(ctx, borrowerAddr, newBorrowerCollateral); err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	// Transfer uToken collateral reward from module account to liquidator
	if k.GetCollateralSetting(ctx, liquidatorAddr, reward.Denom) {
		// For uToken denoms enabled as collateral by liquidator, the uTokens remain in the
		// module account and the keeper tracks the amount
		liquidatorCollateral := k.GetCollateralAmount(ctx, liquidatorAddr, reward.Denom)
		if err = k.SetCollateralAmount(ctx, liquidatorAddr, liquidatorCollateral.Add(reward)); err != nil {
			return sdk.ZeroInt(), sdk.ZeroInt(), err
		}
	} else {
		// For uToken denoms not enabled as collateral by liquidator, the uTokens are sent to their address
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidatorAddr, sdk.NewCoins(reward))
		if err != nil {
			return sdk.ZeroInt(), sdk.ZeroInt(), err
		}
	}

	// Detect bad debt (collateral == 0 after reward) for repayment by reserves next InterestEpoch
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
// borrowed value, and borrow limit. Returns liquidationIncentive (the ratio of bonus collateral
// awarded during Liquidate transactions, and closeFactor (the fraction of a borrower's total
// borrowed value that can be repaid by a liquidator in a single liquidation event.)
func (k Keeper) LiquidationParams(ctx sdk.Context, reward string, borrowed, limit sdk.Dec) (sdk.Dec, sdk.Dec, error) {
	if borrowed.IsNegative() {
		return sdk.ZeroDec(), sdk.ZeroDec(), sdkerrors.Wrap(types.ErrBadValue, borrowed.String())
	}
	if limit.IsNegative() {
		return sdk.ZeroDec(), sdk.ZeroDec(), sdkerrors.Wrap(types.ErrBadValue, limit.String())
	}

	// liquidation incentive is determined by collateral reward denom
	liquidationIncentive, err := k.GetLiquidationIncentive(ctx, reward)
	if err != nil {
		return sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// special case: If borrow limit is zero, close factor is always 1
	if limit.IsZero() {
		return liquidationIncentive, sdk.OneDec(), nil
	}

	params := k.GetParams(ctx)
	// special case: If complete liquidation threshold is zero, close factor is always 1
	if params.CompleteLiquidationThreshold.IsZero() {
		return liquidationIncentive, sdk.OneDec(), nil
	}

	// outside of special cases, close factor scales linearly between MinimumCloseFactor and 1.0,
	// reaching max value when (borrowed / limit) = 1 + CompleteLiquidationThreshold
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

	return liquidationIncentive, closeFactor, nil
}
