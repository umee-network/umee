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
	cdc        codec.Codec
	storeKey   sdk.StoreKey
	paramSpace paramtypes.Subspace
	bankKeeper types.BankKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	bk types.BankKeeper,
) Keeper {

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramSpace: paramSpace,
		bankKeeper: bk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// TotalUTokenSupply returns an sdk.Coin representing the total balance of a
// given uToken type if valid. If the denom is not an accepted uToken type,
// we return a zero amount.
func (k Keeper) TotalUTokenSupply(ctx sdk.Context, uTokenDenom string) sdk.Coin {
	if k.IsAcceptedUToken(ctx, uTokenDenom) {
		return k.bankKeeper.GetSupply(ctx, uTokenDenom)
		// Question: Does bank module still track balances sent (locked) via IBC? If it doesn't
		// then the balance returned here would decrease when the tokens are sent off, which is not
		// what we want. In that case, the keeper should keep an sdk.Int total supply for each uToken type.
	}
	return sdk.NewCoin(uTokenDenom, sdk.ZeroInt())
}

// LendAsset attempts to deposit assets into the leverage module account in
// exchange for uTokens. If asset type is invalid or account balance is
// insufficient, we return an error.
func (k Keeper) LendAsset(ctx sdk.Context, lenderAddr sdk.AccAddress, loan sdk.Coin) error {
	if !k.IsAcceptedToken(ctx, loan.Denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, loan.String())
	}

	if !k.bankKeeper.HasBalance(ctx, lenderAddr, loan) {
		// lender does not have the assets they intend to lend
		return sdkerrors.Wrap(types.ErrInsufficientBalance, loan.String())
	}

	// send token balance to leverage module account
	loanTokens := sdk.NewCoins(loan)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, lenderAddr, types.ModuleName, loanTokens); err != nil {
		return err
	}

	// mint uToken
	uToken, err := k.ExchangeToken(ctx, loan)
	if err != nil {
		return err
	}

	uTokens := sdk.NewCoins(uToken)
	if err = k.bankKeeper.MintCoins(ctx, types.ModuleName, uTokens); err != nil {
		return err
	}

	if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lenderAddr, uTokens); err != nil {
		return err
	}

	return nil
}

// WithdrawAsset attempts to deposit uTokens into the leverage module in exchange
// for the original tokens lent. If the uToken type is invalid or account balance
// insufficient on either side, we return an error.
func (k Keeper) WithdrawAsset(ctx sdk.Context, lenderAddr sdk.AccAddress, uToken sdk.Coin) error {
	if !uToken.IsValid() {
		return sdkerrors.Wrap(types.ErrInvalidAsset, uToken.String())
	}

	if !k.bankKeeper.HasBalance(ctx, lenderAddr, uToken) {
		// Lender does not have the uTokens they intend to redeem
		return sdkerrors.Wrap(types.ErrInsufficientBalance, uToken.String())
	}

	// TODO #213: Calculate lender's borrow limit and current borrowed value, if any.
	// Prevent withdrawing assets when it would bring user borrow limit below current borrowed value.

	withdrawal, err := k.ExchangeUToken(ctx, uToken)
	if err != nil {
		return err
	}

	// Ensure module account has sufficient unreserved tokens to withdraw
	reservedAmount := k.GetReserveAmount(ctx, withdrawal.Denom)
	availableAmount := k.bankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), withdrawal.Denom).Amount
	if withdrawal.Amount.GT(availableAmount.Sub(reservedAmount)) {
		return sdkerrors.Wrap(types.ErrLendingPoolInsufficient, withdrawal.String())
	}

	// send the uTokens from the lender to the module account
	uTokens := sdk.NewCoins(uToken)
	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, lenderAddr, types.ModuleName, uTokens); err != nil {
		return err
	}

	// send the original lent tokens back to lender
	tokens := sdk.NewCoins(withdrawal)
	if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lenderAddr, tokens); err != nil {
		return err
	}

	// burn the minted uTokens
	if err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, uTokens); err != nil {
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
	currentlyBorrowed, err := k.GetBorrowerBorrows(ctx, borrowerAddr)
	if err != nil {
		return err
	}

	// Retrieve borrower's account balance.
	// accountBalance := k.bankKeeper.GetAllBalances(ctx, borrowerAddr)

	// TODO #213: Use oracle to compute borrow limit and current borrowed value.
	// Prevent borrows that exceed borrow limit.

	// Note: Prior to oracle implementation, we cannot compare loan value to borrow limit
	loanTokens := sdk.NewCoins(borrow)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrowerAddr, loanTokens); err != nil {
		return err
	}

	// Determine the total amount of denom borrowed (previously borrowed + newly borrowed)
	totalBorrowed := currentlyBorrowed.AmountOf(borrow.Denom).Add(borrow.Amount)
	err = k.SetBorrow(ctx, borrowerAddr, borrow.Denom, totalBorrowed)
	if err != nil {
		return err
	}
	return nil
}

// RepayAsset attempts to repay an open borrow position with base assets. If asset type is invalid,
// account balance is insufficient, or no open borrow position exists, we return an error.
// Additionally, if the amount provided is greater than the full repayment amount, only the
// necessary amount is transferred.
func (k Keeper) RepayAsset(ctx sdk.Context, borrowerAddr sdk.AccAddress, payment sdk.Coin) error {
	if !payment.IsValid() {
		return sdkerrors.Wrap(types.ErrInvalidAsset, payment.String())
	}

	if !k.IsAcceptedToken(ctx, payment.Denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, payment.String())
	}

	// Determine amount of selected denom currently owed
	owed := k.GetBorrow(ctx, borrowerAddr, payment.Denom)
	if owed.IsZero() {
		// Borrower has no open borrows in the denom presented as payment
		return sdkerrors.Wrap(types.ErrInvalidRepayment, payment.String())
	}

	// Prevent overpaying
	payment.Amount = sdk.MinInt(owed.Amount, payment.Amount)
	if !payment.IsValid() {
		// Catch invalid payments (e.g. from payment.Amount < 0)
		return sdkerrors.Wrap(types.ErrInvalidRepayment, payment.String())
	}

	// send payment to leverage module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, borrowerAddr,
		types.ModuleName,
		sdk.NewCoins(payment),
	); err != nil {
		return err
	}

	// Subtract repaid amount from borrowed amount
	owed.Amount = owed.Amount.Sub(payment.Amount)
	// Store the new total borrowed amount in keeper
	store := ctx.KVStore(k.storeKey)
	key := types.CreateLoanKey(borrowerAddr, payment.Denom)
	if owed.IsZero() {
		store.Delete(key) // Completely repaid
	} else {
		bz, err := owed.Amount.Marshal()
		if err != nil {
			return err
		}
		store.Set(key, bz) // Partially repaid
	}
	return nil
}

// SetCollateralSetting enables or disables a uToken denom for use as collateral by a single borrower.
func (k Keeper) SetCollateralSetting(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string, enable bool) error {
	if !k.IsAcceptedUToken(ctx, denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	// TODO #213: If enable=false, use oracle to compute current borrowed value and borrow limit after disable.
	// Prevent disabling collateral when it would bring user borrow limit below current borrowed value.

	// Enable sets to true; disable removes from KVstore rather than setting false
	store := ctx.KVStore(k.storeKey)
	key := types.CreateCollateralSettingKey(borrowerAddr, denom)
	if enable {
		store.Set(key, []byte{0x01})
	} else {
		store.Delete(key)
	}
	return nil
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
func (k Keeper) LiquidateBorrow(
	ctx sdk.Context, liquidatorAddr, borrowerAddr sdk.AccAddress, repayment sdk.Coin, rewardDenom string,
) error {
	if !repayment.IsValid() {
		return sdkerrors.Wrap(types.ErrInvalidAsset, repayment.String())
	}
	if !k.IsAcceptedUToken(ctx, rewardDenom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, rewardDenom)
	}

	// Get total borrowed by borrower (all denoms)
	borrowed, err := k.GetBorrowerBorrows(ctx, borrowerAddr)
	if err != nil {
		return err
	}

	// Get borrower uToken balances, for all uToken denoms enabled as collateral
	collateral := k.GetBorrowerCollateral(ctx, borrowerAddr)
	if err != nil {
		return err
	}

	// Use oracle helper functions to find total borrowed value in USD
	borrowValue, err := k.TotalPrice(ctx, borrowed)
	if err != nil {
		return err
	}

	// Use collateral weights to compute borrow limit from enabled collateral
	borrowLimit, err := k.CalculateBorrowLimit(ctx, collateral)
	if err != nil {
		return err
	}

	// Confirm borrower's eligibility for liquidation
	if borrowLimit.GTE(borrowValue) {
		return sdkerrors.Wrap(types.ErrLiquidationIneligible, borrowerAddr.String())
	}

	// Get reward-specific incentive and dynamic close factor
	baseRewardDenom := k.FromUTokenToTokenDenom(ctx, rewardDenom)
	liquidationIncentive, closeFactor, err := k.LiquidationParams(ctx, baseRewardDenom, borrowValue, borrowLimit)
	if err != nil {
		return err
	}

	// Repayment cannot exceed liquidator's available balance
	liquidatorBalance := k.bankKeeper.GetBalance(ctx, liquidatorAddr, repayment.Denom)
	if repayment.Amount.GTE(liquidatorBalance.Amount) {
		repayment.Amount = liquidatorBalance.Amount
	}

	// Repayment cannot exceed borrower's borrowed amount of selected denom
	if repayment.Amount.GTE(borrowed.AmountOf(repayment.Denom)) {
		repayment.Amount = borrowed.AmountOf(repayment.Denom)
	}

	// Repayment cannot exceed borrowed value * close factor
	repayValue, err := k.Price(ctx, repayment)
	if err != nil {
		return err
	}
	if repayValue.GTE(borrowValue.Mul(closeFactor)) {
		maxRepayValue := borrowValue.Mul(closeFactor)
		// repayment *= (maxRepayValue / repayValue)
		repayment.Amount = repayment.Amount.ToDec().Mul(maxRepayValue).Quo(repayValue).TruncateInt()
		repayValue = maxRepayValue
	}

	// Given repay denom and amount, use oracle to find equivalent amount of rewardDenom's base asset
	baseReward, err := k.EquivalentValue(ctx, repayment, baseRewardDenom)
	if err != nil {
		return err
	}

	// Convert reward tokens back to uTokens
	reward, err := k.ExchangeToken(ctx, baseReward)
	if err != nil {
		return err
	}

	// Apply liquidation incentive
	reward.Amount = reward.Amount.ToDec().Mul(sdk.OneDec().Add(liquidationIncentive)).TruncateInt()

	// Reward amount cannot exceed available collateral
	if reward.Amount.GTE(collateral.AmountOf(rewardDenom)) {
		// only pay what can be correctly compensated
		repayment.Amount = repayment.Amount.Mul(collateral.AmountOf(rewardDenom)).Quo(reward.Amount)
		// use all collateral of selected denom
		reward.Amount = collateral.AmountOf(rewardDenom)
	}

	// Send repayment to leverage module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, liquidatorAddr,
		types.ModuleName,
		sdk.NewCoins(repayment),
	); err != nil {
		return err
	}

	// Store the remaining borrowed amount in keeper
	owed := borrowed.AmountOf(repayment.Denom).Sub(repayment.Amount)
	store := ctx.KVStore(k.storeKey)
	key := types.CreateLoanKey(borrowerAddr, repayment.Denom)
	if owed.IsZero() {
		store.Delete(key) // Completely repaid
	} else {
		bz, err := owed.Marshal()
		if err != nil {
			return err
		}
		store.Set(key, bz) // Partially repaid
	}

	// Transfer uToken collateral reward from borrower to liquidator
	if err := k.bankKeeper.SendCoins(ctx, borrowerAddr, liquidatorAddr, sdk.NewCoins(reward)); err != nil {
		return err
	}

	return nil
}

// LiquidationParams computes dynamic liquidation parameters based on collateral denomination,
// borrowed value, and borrow limit
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

	// close factor scales linearly between MinimumCloseFactor and 1.0,
	// reaching max value when (borrowed / limit) = CompleteLiquidationThreshold
	closeFactor := k.GetParams(ctx).MinimumCloseFactor
	completeLiquidationThreshold := k.GetParams(ctx).CompleteLiquidationThreshold
	closeFactor = Interpolate(
		borrowed.Quo(limit),          // x
		sdk.ZeroDec(),                // xMin
		closeFactor,                  // yMin
		completeLiquidationThreshold, // xMax
		sdk.OneDec(),                 // yMax
	)
	if closeFactor.GTE(sdk.OneDec()) {
		closeFactor = sdk.OneDec()
	}

	return liquidationIncentive, closeFactor, nil
}
