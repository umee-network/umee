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

// SetCollateralSetting enables or disables a uToken denom for use as collateral by a single borrower.
func (k Keeper) SetCollateralSetting(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string, enable bool) error {
	if !k.IsAcceptedUToken(ctx, denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

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

// GetLoan returns an sdk.Coin representing how much of a given denom a borrower currently owes.
func (k Keeper) GetLoan(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	owed := sdk.NewCoin(denom, sdk.ZeroInt())
	key := types.CreateLoanKey(borrowerAddr, denom)
	bz := store.Get(key)
	if bz != nil {
		err := owed.Amount.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
	}
	return owed
}

// GetAllBorrowerLoans returns an sdk.Coins object containing all open borrows associated with an address
func (k Keeper) GetAllBorrowerLoans(ctx sdk.Context, borrowerAddr sdk.AccAddress) (sdk.Coins, error) {
	totalBorrowed := sdk.NewCoins()
	store := ctx.KVStore(k.storeKey)
	prefix := types.CreateLoanKeyNoDenom(borrowerAddr)
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// k is prefix | denom | 0x00
		k, v := iter.Key(), iter.Value()
		denom := string(k[len(prefix) : len(k)-1]) // remove prefix and null-terminator
		var amount sdk.Int
		if err := amount.Unmarshal(v); err != nil {
			return sdk.NewCoins(), err // improperly marshaled loan amount should never happen
		}
		// For each loan found, add it to totalBorrowed
		totalBorrowed = totalBorrowed.Add(sdk.NewCoin(denom, amount))
	}
	totalBorrowed.Sort()
	return totalBorrowed, nil
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

	// Determine amount of all tokens currently borrowed
	currentlyBorrowed, err := k.GetAllBorrowerLoans(ctx, borrowerAddr)
	if err != nil {
		return err
	}

	// Retrieve borrower's account balance.
	// accountBalance := k.bankKeeper.GetAllBalances(ctx, borrowerAddr)

	// TODO (Oracle+Params+CollateralSettings): Calculate borrow limit from uTokens in borrower's account
	// TODO (Oracle): Calculate borrow limit already used (from currentlyBorrowed)
	// TODO (Oracle): Calculate loanTokens value
	// TODO: ErrBorrowLimitLow if (borrow limit used + loan value > borrow limit)

	// Note: Prior to oracle implementation, we cannot compare loan value to borrow limit
	loanTokens := sdk.NewCoins(loan)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrowerAddr, loanTokens); err != nil {
		return err
	}
	// Determine the total amount of denom borrowed (previously borrowed + newly borrowed)
	totalBorrowed := currentlyBorrowed.AmountOf(loan.Denom).Add(loan.Amount)
	store := ctx.KVStore(k.storeKey)
	bz, err := totalBorrowed.Marshal()
	if err != nil {
		return err
	}
	store.Set(types.CreateLoanKey(borrowerAddr, loan.Denom), bz)
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

	// Determine amount of selected denom currently owed
	owed := k.GetLoan(ctx, borrowerAddr, payment.Denom)
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
