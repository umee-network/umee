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
	paramStore paramtypes.Subspace
	bankKeeper types.BankKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey sdk.StoreKey,
	paramStore paramtypes.Subspace,
	bk types.BankKeeper,
) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramStore: paramStore,
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

	// mint uTokens
	// TODO: Use exchange rate instead of 1:1 redeeming
	uToken := sdk.NewCoin(k.FromTokenToUTokenDenom(ctx, loan.Denom), loan.Amount)
	uTokens := sdk.NewCoins(uToken)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, uTokens); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lenderAddr, uTokens); err != nil {
		return err
	}

	return nil
}

// WithdrawAsset attempts to deposit uTokens into the leverage module in exchange
// for the original tokens lent. If the uToken type is invalid or account balance
// insufficient on either side, we return an error.
func (k Keeper) WithdrawAsset(ctx sdk.Context, lenderAddr sdk.AccAddress, uToken sdk.Coin) error {
	if !k.bankKeeper.HasBalance(ctx, lenderAddr, uToken) {
		// Lender does not have the uTokens they intend to redeem
		return sdkerrors.Wrap(types.ErrInsufficientBalance, uToken.String())
	}

	// ensure the tokens exist in the leverage module account's balance
	// TODO: Use exchange rate instead of 1:1 redeeming
	tokenDenom := k.FromUTokenToTokenDenom(ctx, uToken.Denom)
	token := sdk.NewCoin(tokenDenom, uToken.Amount)
	if !k.bankKeeper.HasBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), token) {
		// TODO: We should never enter this case -- consider a panic
		return sdkerrors.Wrap(types.ErrInsufficientBalance, token.String())
	}

	// send the uTokens from the lender to the module account
	uTokens := sdk.NewCoins(uToken)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, lenderAddr, types.ModuleName, uTokens); err != nil {
		return err
	}

	// send the original lent tokens back to lender
	tokens := sdk.NewCoins(token)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lenderAddr, tokens); err != nil {
		return err
	}

	// burn the minted uTokens
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, uTokens); err != nil {
		return err
	}

	return nil
}

// SetCollateral enables or disables a uToken denom for use as collateral by a single borrower.
func (k Keeper) SetCollateral(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string, enable bool) error {
	if !k.IsAcceptedUToken(ctx, denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	// TODO: Enable sets to true; disable removes from KVstore rather than setting false

	return nil
}

// BorrowAsset attempts to borrow assets from the leverage module account using
// collateral uTokens. If asset type is invalid, collateral is insufficient,
// or module balance is insufficient, we return an error.
func (k Keeper) BorrowAsset(ctx sdk.Context, borrowerAddr sdk.AccAddress, loan sdk.Coin) error {
	if !k.IsAcceptedToken(ctx, loan.Denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, loan.String())
	}
	// ensure module account has sufficient assets to loan out
	if !k.bankKeeper.HasBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), loan) {
		return sdkerrors.Wrap(types.ErrLendingPoolInsufficient, loan.String())
	}

	loanTokens := sdk.NewCoins(loan)
	// TODO: Calculate loan value (oracle placeholder)

	// TODO: Calculate borrow limit (params + account + oracle placeholder)

	// TODO: Calculate borrow limit already used (keeper + oracle placeholder)

	// TODO: Loan is rejected if (borrow limit used + loan value > borrow limit)
	// use ErrBorrowLimitLow

	// send borrowed assets from module account to borrower
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrowerAddr, loanTokens); err != nil {
		return err
	}

	return nil
}

// RepayAsset attempts to repay an open borrow position with base assets. If asset type is invalid,
// account balance is insufficient, or no open borrow position exists, we return an error.
// Additionally, if the amount provided is greater than the full repayment amount, only the
// necessary amount is transferred.
func (k Keeper) RepayAsset(ctx sdk.Context, borrowerAddr sdk.AccAddress, payment sdk.Coin) error {
	if !k.IsAcceptedToken(ctx, payment.Denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, payment.String())
	}

	/*
		TODO: Detect nonexistent borrow case
		if // borrower has no open borrows in payment denom {
			// borrower has no open borrows in the denom presented as payment
			return sdkerrors.Wrap(types.ErrRepayNonexistentBorrow, payment.String())
		}
	*/

	// TODO: Determine borrower's full repayment amount in selected denomination

	// TODO: If full repayment amount < repayment offered, set payment = full repayment amount

	// send payment to leverage module account
	paymentTokens := sdk.NewCoins(payment)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, borrowerAddr, types.ModuleName, paymentTokens); err != nil {
		return err
	}

	return nil
}
