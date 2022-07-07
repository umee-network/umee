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

// WithdrawAsset attempts to deposit uTokens into the leverage module in exchange
// for the original tokens supplied. Accepts a uToken amount to exchange for base tokens.
// If the uToken denom is invalid or account or module balance insufficient, returns error.
func (k Keeper) WithdrawAsset(ctx sdk.Context, supplierAddr sdk.AccAddress, coin sdk.Coin) error {
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

// BorrowAsset attempts to borrow tokens from the leverage module account using
// collateral uTokens. If asset type is invalid, collateral is insufficient,
// or module balance is insufficient, we return an error.
func (k Keeper) BorrowAsset(ctx sdk.Context, borrowerAddr sdk.AccAddress, borrow sdk.Coin) error {
	if err := k.validateBorrowAsset(ctx, borrow); err != nil {
		return err
	}

	// Ensure module account has sufficient unreserved tokens to loan out
	reservedAmount := k.GetReserveAmount(ctx, borrow.Denom)
	availableAmount := k.ModuleBalance(ctx, borrow.Denom)
	if borrow.Amount.GT(availableAmount.Sub(reservedAmount)) {
		return types.ErrLendingPoolInsufficient.Wrap(borrow.String())
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

// RepayAsset attempts to repay a borrow position. If asset type is invalid, account balance
// is insufficient, or borrower has no borrows in payment denom to repay, we return an error.
// Additionally, if the amount provided is greater than the full repayment amount, only the
// necessary amount is transferred. Because amount repaid may be less than the repayment attempted,
// RepayAsset returns the actual amount repaid.
func (k Keeper) RepayAsset(ctx sdk.Context, borrowerAddr sdk.AccAddress, payment sdk.Coin) (sdk.Int, error) {
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

// AddCollateral enables selected uTokens for use as collateral by a single borrower.
func (k Keeper) AddCollateral(ctx sdk.Context, borrowerAddr sdk.AccAddress, coin sdk.Coin) error {
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

// RemoveCollateral disables selected uTokens for use as collateral by a single borrower.
func (k Keeper) RemoveCollateral(ctx sdk.Context, borrowerAddr sdk.AccAddress, coin sdk.Coin) error {
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
	ctx sdk.Context, liquidatorAddr, borrowerAddr sdk.AccAddress, desiredRepay sdk.Coin, rewardDenom string,
) (sdk.Coin, sdk.Coin, sdk.Coin, error) {
	if err := k.validateAcceptedAsset(ctx, desiredRepay); err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	if err := k.validateAcceptedDenom(ctx, rewardDenom); err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// calculate Token repay, and uToken and Token reward amounts allowed by liquidation rules and available balances
	baseRepay, collateralReward, baseReward, err := k.liquidationOutcome(
		ctx,
		liquidatorAddr,
		borrowerAddr,
		desiredRepay,
		rewardDenom,
	)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// send repayment from liquidator to leverage module account
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, liquidatorAddr, types.ModuleName, sdk.NewCoins(baseRepay))
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	// update borrower's remaining borrowed amount
	newBorrow := k.GetBorrow(ctx, borrowerAddr, baseRepay.Denom).Amount.Sub(baseRepay.Amount)
	if err = k.setBorrow(ctx, borrowerAddr, sdk.NewCoin(baseRepay.Denom, newBorrow)); err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// reduce borrower's collateral by collateral reward amount
	oldCollateral := k.GetCollateralAmount(ctx, borrowerAddr, collateralReward.Denom)
	newCollateral := sdk.NewCoin(collateralReward.Denom, oldCollateral.Amount.Sub(collateralReward.Amount))
	if err = k.setCollateralAmount(ctx, borrowerAddr, newCollateral); err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	// burn the collateral reward uTokens and set the new total uToken supply
	if err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(collateralReward)); err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}
	if err = k.setUTokenSupply(ctx, k.GetUTokenSupply(ctx, collateralReward.Denom).Sub(collateralReward)); err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// send base rewards from module to liquidator's account
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidatorAddr, sdk.NewCoins(baseReward))
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// detect bad debt if collateral is completely exhausted
	if k.GetBorrowerCollateral(ctx, borrowerAddr).IsZero() {
		// TODO: exclude blacklisted collateral
		for _, coin := range k.GetBorrowerBorrows(ctx, borrowerAddr) {
			// set a bad debt flag for each borrowed denom
			if err := k.setBadDebtAddress(ctx, borrowerAddr, coin.Denom, true); err != nil {
				return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, err
			}
		}
	}

	return baseRepay, collateralReward, baseReward, nil
}
