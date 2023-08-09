package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/leverage/types"
	"github.com/umee-network/umee/v5/x/ugov"
)

type Keeper struct {
	cdc                    codec.Codec
	storeKey               storetypes.StoreKey
	paramSpace             paramtypes.Subspace
	bankKeeper             types.BankKeeper
	oracleKeeper           types.OracleKeeper
	ugov                   ugov.EmergencyGroupBuilder
	liquidatorQueryEnabled bool
	meTokenAddr            sdk.AccAddress

	tokenHooks []types.TokenHooks
	bondHooks  []types.BondHooks
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	b types.BankKeeper,
	o types.OracleKeeper,
	ugov ugov.EmergencyGroupBuilder,
	enableLiquidatorQuery bool,
	meTokenAddr sdk.AccAddress,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:                    cdc,
		storeKey:               storeKey,
		paramSpace:             paramSpace,
		bankKeeper:             b,
		oracleKeeper:           o,
		ugov:                   ugov,
		liquidatorQueryEnabled: enableLiquidatorQuery,
		meTokenAddr:            meTokenAddr,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetTokenHooks sets the module's token registry hooks. Token hooks can only be set once.
func (k *Keeper) SetTokenHooks(h ...types.TokenHooks) {
	if k.tokenHooks != nil {
		panic("leverage token hooks already set")
	}

	k.tokenHooks = h
}

// SetBondHooks sets the module's bonded amount and force unbonding hooks.
// Panics if Bond hooks have been already set.
func (k *Keeper) SetBondHooks(h ...types.BondHooks) {
	if k.bondHooks != nil {
		panic("leverage bond hooks already set")
	}

	k.bondHooks = h
}

// ModuleBalance returns the amount of a given token held in the x/leverage module account
func (k Keeper) ModuleBalance(ctx sdk.Context, denom string) sdk.Coin {
	amount := k.bankKeeper.SpendableCoins(ctx, authtypes.NewModuleAddress(types.ModuleName)).AmountOf(denom)
	return sdk.NewCoin(denom, amount)
}

// Supply attempts to deposit assets into the leverage module account in
// exchange for uTokens. If asset type is invalid or account balance is
// insufficient, we return an error. Returns the amount of uTokens minted.
// Note: For supplying from a module account instead of a user, use SupplyFromModule.
func (k Keeper) Supply(ctx sdk.Context, supplierAddr sdk.AccAddress, coin sdk.Coin) (sdk.Coin, error) {
	if err := k.validateSupply(ctx, coin); err != nil {
		return sdk.Coin{}, err
	}

	// determine uToken amount to mint
	uToken, err := k.ToUToken(ctx, coin)
	if err != nil {
		return sdk.Coin{}, err
	}

	// send token balance to leverage module account
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, supplierAddr, types.ModuleName, sdk.NewCoins(coin))
	if err != nil {
		return sdk.Coin{}, err
	}

	// mint uToken and set new total uToken supply
	uTokens := sdk.NewCoins(uToken)
	if err = k.bankKeeper.MintCoins(ctx, types.ModuleName, uTokens); err != nil {
		return sdk.Coin{}, err
	}
	if err = k.setUTokenSupply(ctx, k.GetUTokenSupply(ctx, uToken.Denom).Add(uToken)); err != nil {
		return sdk.Coin{}, err
	}

	// The uTokens are sent to supplier address
	if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, supplierAddr, uTokens); err != nil {
		return sdk.Coin{}, err
	}

	return uToken, nil
}

// SupplyFromModule attempts to deposit assets into the leverage module account in
// exchange for uTokens on behalf of another module. In addition to the regular error
// return, also returns a boolean which indicates whether the error was recoverable.
// A recoverable = true error means SupplyFromModule was aborted without harming state.
func (k Keeper) SupplyFromModule(ctx sdk.Context, fromModule string, coin sdk.Coin) (sdk.Coin, bool, error) {
	if err := k.validateSupply(ctx, coin); err != nil {
		return sdk.Coin{}, true, err
	}

	// determine uToken amount to mint
	uToken, err := k.ToUToken(ctx, coin)
	if err != nil {
		return sdk.Coin{}, true, err
	}

	// All errors past this point are considered non-recoverable

	// send token balance to leverage module account
	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, fromModule, types.ModuleName, sdk.NewCoins(coin))
	if err != nil {
		return sdk.Coin{}, false, err
	}

	// mint uToken and set new total uToken supply
	uTokens := sdk.NewCoins(uToken)
	if err = k.bankKeeper.MintCoins(ctx, types.ModuleName, uTokens); err != nil {
		return sdk.Coin{}, false, err
	}
	if err = k.setUTokenSupply(ctx, k.GetUTokenSupply(ctx, uToken.Denom).Add(uToken)); err != nil {
		return sdk.Coin{}, false, err
	}

	// The uTokens are sent to supplier module
	if err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, fromModule, uTokens); err != nil {
		return sdk.Coin{}, false, err
	}

	// On nil error, recoverable is set to true
	return uToken, true, nil
}

// Withdraw attempts to redeem uTokens from the leverage module in exchange for base tokens.
// If there are not enough uTokens in balance, Withdraw will attempt to withdraw uToken collateral
// to make up the difference. If the uToken denom is invalid or balances are insufficient to withdraw
// the amount requested, returns an error. Returns the amount of base tokens received.
// This function does NOT check that a borrower remains under their borrow limit or that
// collateral liquidity remains healthy - those assertions have been moved to MsgServer.
// Returns a boolean which is true if some or all of the withdrawn uTokens were from collateral.
// Note: For withdrawing to a module account instead of a user, use WithdrawToModule.
func (k Keeper) Withdraw(ctx sdk.Context, supplierAddr sdk.AccAddress, uToken sdk.Coin) (sdk.Coin, bool, error) {
	isFromCollateral := false

	if err := validateUToken(uToken); err != nil {
		return sdk.Coin{}, isFromCollateral, err
	}

	// calculate base asset amount to withdraw
	token, err := k.ToToken(ctx, uToken)
	if err != nil {
		return sdk.Coin{}, isFromCollateral, err
	}

	// Ensure module account has sufficient unreserved tokens to withdraw
	availableAmount := k.AvailableLiquidity(ctx, token.Denom)
	if token.Amount.GT(availableAmount) {
		return sdk.Coin{}, isFromCollateral, types.ErrLendingPoolInsufficient.Wrap(token.String())
	}

	// Withdraw will first attempt to use any uTokens in the supplier's wallet
	amountFromWallet := sdk.MinInt(k.bankKeeper.SpendableCoins(ctx, supplierAddr).AmountOf(uToken.Denom), uToken.Amount)
	// Any additional uTokens must come from the supplier's collateral
	amountFromCollateral := uToken.Amount.Sub(amountFromWallet)

	if amountFromCollateral.IsPositive() {
		// This indicates that borrower health check cannot be skipped after MsgWithdraw
		isFromCollateral = true

		// Check for sufficient collateral
		collateral := k.GetBorrowerCollateral(ctx, supplierAddr)
		collateralAmount := collateral.AmountOf(uToken.Denom)
		if collateralAmount.LT(amountFromCollateral) {
			return sdk.Coin{}, isFromCollateral, types.ErrInsufficientBalance.Wrapf(
				"%s uToken balance + %s from collateral is less than %s to withdraw",
				amountFromWallet, collateralAmount, uToken,
			)
		}

		unbondedCollateral := k.unbondedCollateral(ctx, supplierAddr, uToken.Denom)
		if unbondedCollateral.Amount.LT(amountFromCollateral) {
			return sdk.Coin{}, isFromCollateral, types.ErrBondedCollateral.Wrapf(
				"%s unbonded collateral is less than %s to withdraw from collateral",
				unbondedCollateral, amountFromCollateral,
			)
		}

		// reduce the supplier's collateral by amountFromCollateral
		newCollateral := sdk.NewCoin(uToken.Denom, collateralAmount.Sub(amountFromCollateral))
		if err = k.setCollateral(ctx, supplierAddr, newCollateral); err != nil {
			return sdk.Coin{}, isFromCollateral, err
		}
	}

	// transfer amountFromWallet uTokens to the module account
	uTokens := sdk.NewCoins(sdk.NewCoin(uToken.Denom, amountFromWallet))
	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, supplierAddr, types.ModuleName, uTokens); err != nil {
		return sdk.Coin{}, isFromCollateral, err
	}

	// send the base assets to supplier
	tokens := sdk.NewCoins(token)
	if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, supplierAddr, tokens); err != nil {
		return sdk.Coin{}, isFromCollateral, err
	}

	// burn the uTokens and set the new total uToken supply
	if err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(uToken)); err != nil {
		return sdk.Coin{}, isFromCollateral, err
	}
	if err = k.setUTokenSupply(ctx, k.GetUTokenSupply(ctx, uToken.Denom).Sub(uToken)); err != nil {
		return sdk.Coin{}, isFromCollateral, err
	}

	return token, isFromCollateral, nil
}

// WithdrawToModule attempts to redeem uTokens from the leverage module in exchange for base tokens.
// This is done on behalf of another module, not by a user account. Modules do not have collateral.
// If the uToken denom is invalid or balances are insufficient to withdraw the amount requested,
// returns an error. Returns the amount of base tokens received. In addition to the regular error
// return, also returns a boolean which indicates whether the error was recoverable.
// A recoverable = true error means WithdrawToModule was aborted without harming state.
func (k Keeper) WithdrawToModule(ctx sdk.Context, toModule string, uToken sdk.Coin) (sdk.Coin, bool, error) {
	if err := validateUToken(uToken); err != nil {
		return sdk.Coin{}, true, err
	}

	// calculate base asset amount to withdraw
	token, err := k.ToToken(ctx, uToken)
	if err != nil {
		return sdk.Coin{}, true, err
	}

	// Ensure leverage module account has sufficient unreserved tokens to withdraw
	availableAmount := k.AvailableLiquidity(ctx, token.Denom)
	if token.Amount.GT(availableAmount) {
		return sdk.Coin{}, true, types.ErrLendingPoolInsufficient.Wrap(token.String())
	}

	// All errors past this point are considered non-recoverable

	// transfer uTokens to the leverage module account
	if err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, toModule, types.ModuleName, sdk.NewCoins(uToken)); err != nil {
		return sdk.Coin{}, false, err
	}

	// send the base assets to withdrawing module
	tokens := sdk.NewCoins(token)
	if err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, toModule, tokens); err != nil {
		return sdk.Coin{}, false, err
	}

	// burn the uTokens and set the new total uToken supply
	if err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(uToken)); err != nil {
		return sdk.Coin{}, false, err
	}
	if err = k.setUTokenSupply(ctx, k.GetUTokenSupply(ctx, uToken.Denom).Sub(uToken)); err != nil {
		return sdk.Coin{}, false, err
	}

	// On nil error, recoverable is set to true
	return token, true, nil
}

// Borrow attempts to borrow tokens from the leverage module account using
// collateral uTokens. If asset type is invalid,  or module balance is insufficient,
// we return an error.
// This function does NOT check that a borrower remains under their borrow limit or that
// collateral liquidity remains healthy - those assertions have been moved to MsgServer.
func (k Keeper) Borrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, borrow sdk.Coin) error {
	if err := k.validateBorrow(ctx, borrow); err != nil {
		return err
	}

	// Ensure module account has sufficient unreserved tokens to loan out
	availableAmount := k.AvailableLiquidity(ctx, borrow.Denom)
	if borrow.Amount.GT(availableAmount) {
		return types.ErrLendingPoolInsufficient.Wrap(borrow.String())
	}

	// Determine amount of all tokens currently borrowed
	borrowed := k.GetBorrowerBorrows(ctx, borrowerAddr)

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx, types.ModuleName, borrowerAddr, sdk.NewCoins(borrow),
	); err != nil {
		return err
	}

	// Determine the total amount of denom borrowed (previously borrowed + newly borrowed)
	newBorrow := borrowed.AmountOf(borrow.Denom).Add(borrow.Amount)
	return k.setBorrow(ctx, borrowerAddr, sdk.NewCoin(borrow.Denom, newBorrow))
}

// Repay attempts to repay a borrow position. If asset type is invalid, account balance
// is insufficient, or borrower has no borrows in payment denom to repay, we return an error.
// Additionally, if the amount provided is greater than the full repayment amount, only the
// necessary amount is transferred. Because amount repaid may be less than the repayment attempted,
// Repay returns the actual amount repaid.
func (k Keeper) Repay(ctx sdk.Context, borrowerAddr sdk.AccAddress, payment sdk.Coin) (sdk.Coin, error) {
	if err := validateBaseToken(payment); err != nil {
		return sdk.Coin{}, err
	}

	// determine amount of selected denom currently owed
	owed := k.GetBorrow(ctx, borrowerAddr, payment.Denom)
	if owed.IsZero() {
		// no need to repay - everything is all right
		return coin.Zero(payment.Denom), nil
	}

	// prevent overpaying
	payment.Amount = sdk.MinInt(owed.Amount, payment.Amount)

	// send payment to leverage module account
	if err := k.repayBorrow(ctx, borrowerAddr, borrowerAddr, payment); err != nil {
		return sdk.Coin{}, err
	}
	return payment, nil
}

// Collateralize enables selected uTokens for use as collateral by a single borrower.
// This function does NOT check that collateral share and collateral liquidity remain healthy.
// Those assertions have been moved to MsgServer.
func (k Keeper) Collateralize(ctx sdk.Context, borrowerAddr sdk.AccAddress, uToken sdk.Coin) error {
	if err := k.validateCollateralize(ctx, uToken); err != nil {
		return err
	}

	currentCollateral := k.GetCollateral(ctx, borrowerAddr, uToken.Denom)
	if err := k.setCollateral(ctx, borrowerAddr, currentCollateral.Add(uToken)); err != nil {
		return err
	}

	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, borrowerAddr, types.ModuleName, sdk.NewCoins(uToken))
}

// Decollateralize disables selected uTokens for use as collateral by a single borrower.
// This function does NOT check that a borrower remains under their borrow limit.
// That assertion has been moved to MsgServer.
func (k Keeper) Decollateralize(ctx sdk.Context, borrowerAddr sdk.AccAddress, uToken sdk.Coin) error {
	if err := validateUToken(uToken); err != nil {
		return err
	}

	// Detect where sufficient collateral exists to disable
	collateral := k.GetBorrowerCollateral(ctx, borrowerAddr)
	collateralAmount := collateral.AmountOf(uToken.Denom)
	if collateralAmount.LT(uToken.Amount) {
		return types.ErrInsufficientCollateral
	}

	unbondedCollateral := k.unbondedCollateral(ctx, borrowerAddr, uToken.Denom)
	if unbondedCollateral.Amount.LT(uToken.Amount) {
		return types.ErrBondedCollateral.Wrapf(
			"%s unbonded collateral uTokens are less than %s to decollateralize",
			unbondedCollateral, uToken,
		)
	}

	// Decollateralizing uTokens withdraws them from the module account and returns them to the user
	newCollateralAmount := collateral.AmountOf(uToken.Denom).Sub(uToken.Amount)
	if err := k.setCollateral(ctx, borrowerAddr, sdk.NewCoin(uToken.Denom, newCollateralAmount)); err != nil {
		return err
	}
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrowerAddr, sdk.NewCoins(uToken))
}

// Liquidate attempts to repay one of an eligible borrower's borrows (in part or in full) in exchange for
// some of the borrower's uToken collateral or associated base tokens. If the borrower is not over their
// liquidation limit, or the repayment or reward denominations are invalid, an error is returned. If the
// attempted repayment is greater than the amount owed or the maximum that can be repaid due to parameters
// or available balances, then a partial liquidation, equal to the maximum valid amount, is performed.
// Because partial liquidation is possible and exchange rates vary, Liquidate returns the actual amount of
// tokens repaid, collateral liquidated, and base tokens or uTokens rewarded.
func (k Keeper) Liquidate(
	ctx sdk.Context, liquidatorAddr, borrowerAddr sdk.AccAddress, requestedRepay sdk.Coin, rewardDenom string,
) (repaid sdk.Coin, liquidated sdk.Coin, reward sdk.Coin, err error) {
	if err := k.validateAcceptedAsset(ctx, requestedRepay); err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// detect if the user selected a base token reward instead of a uToken
	directLiquidation := !coin.HasUTokenPrefix(rewardDenom)
	if !directLiquidation {
		// convert rewardDenom to base token
		rewardDenom = coin.StripUTokenDenom(rewardDenom)
	}
	// ensure that base reward is a registered token
	if err := k.validateAcceptedDenom(ctx, rewardDenom); err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}

	tokenRepay, uTokenLiquidate, tokenReward, err := k.getLiquidationAmounts(
		ctx,
		liquidatorAddr,
		borrowerAddr,
		requestedRepay,
		rewardDenom,
		directLiquidation,
		false,
	)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	if tokenRepay.IsZero() {
		// Zero repay amount returned from liquidation computation means the target was eligible for liquidation
		// but the proposed reward and repayment would have zero effect.
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, types.ErrLiquidationRepayZero
	}

	// repay some of the borrower's debt using the liquidator's balance
	if err = k.repayBorrow(ctx, liquidatorAddr, borrowerAddr, tokenRepay); err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}

	if directLiquidation {
		err = k.liquidateCollateral(ctx, borrowerAddr, liquidatorAddr, uTokenLiquidate, tokenReward)
	} else {
		// send uTokens from borrower collateral to liquidator's account
		err = k.decollateralize(ctx, borrowerAddr, liquidatorAddr, uTokenLiquidate)
	}
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// check for bad debt and trigger forced unbond hooks
	if err := k.postLiquidate(ctx, borrowerAddr, uTokenLiquidate.Denom); err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// the last return value is the liquidator's selected reward
	if directLiquidation {
		return tokenRepay, uTokenLiquidate, tokenReward, nil
	}
	return tokenRepay, uTokenLiquidate, uTokenLiquidate, nil
}

// LeveragedLiquidate
func (k Keeper) LeveragedLiquidate(
	ctx sdk.Context, liquidatorAddr, borrowerAddr sdk.AccAddress, repayDenom, rewardDenom string,
) (repaid sdk.Coin, reward sdk.Coin, err error) {
	// If the message did not specify repay or reward denoms, select one arbitrarily (first in
	// denom alphabetical order) from borrower position. Then proceed normally with the transaction.
	if repayDenom == "" {
		borrowed := k.GetBorrowerBorrows(ctx, borrowerAddr)
		if !borrowed.IsZero() {
			repayDenom = borrowed[0].Denom
		}
	}
	if rewardDenom == "" {
		collateral := k.GetBorrowerCollateral(ctx, borrowerAddr)
		if !collateral.IsZero() {
			rewardDenom = coin.StripUTokenDenom(collateral[0].Denom)
		}
	}

	if err := k.validateAcceptedDenom(ctx, repayDenom); err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	if err := k.validateAcceptedDenom(ctx, rewardDenom); err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	uRewardDenom := coin.ToUTokenDenom(rewardDenom)

	tokenRepay, uTokenReward, _, err := k.getLiquidationAmounts(
		ctx,
		liquidatorAddr,
		borrowerAddr,
		sdk.NewCoin(repayDenom, sdk.OneInt()), // amount is ignored for LeveragedLiquidate
		rewardDenom,
		false,
		true,
	)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	if tokenRepay.IsZero() || uTokenReward.IsZero() {
		return sdk.Coin{}, sdk.Coin{}, types.ErrLiquidationRepayZero
	}

	// directly move debt from borrower to liquidator without transferring any tokens between accounts
	if err := k.moveBorrow(ctx, borrowerAddr, liquidatorAddr, tokenRepay); err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// directly move collateral from borrower to liquidator while keeping it collateralized
	if err := k.moveCollateral(ctx, borrowerAddr, liquidatorAddr, uTokenReward); err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// check for bad debt and trigger forced unbond hooks
	return tokenRepay, uTokenReward, k.postLiquidate(ctx, borrowerAddr, uRewardDenom)
}
