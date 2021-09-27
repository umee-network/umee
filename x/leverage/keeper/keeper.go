package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/x/leverage/types"
)

type Keeper struct {
	cdc        codec.Codec
	storeKey   sdk.StoreKey
	memKey     sdk.StoreKey
	bankKeeper types.BankKeeper
}

func NewKeeper(cdc codec.Codec, storeKey, memKey sdk.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		memKey:   memKey,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// FromUTokenToTokenDenom returns the associated token denom for the given uToken
// denom. If the uToken denom does not exist, we assume the association is
// invalid and we return an empty string.
func (k Keeper) FromUTokenToTokenDenom(ctx sdk.Context, uTokenDenom string) string {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateUTokenDenomKey(uTokenDenom)

	bz := store.Get(key)
	if len(bz) == 0 {
		return ""
	}

	return string(bz)
}

// FromTokenToUTokenDenom returns the associated uToken denom for the given token
// denom. If the token denom does not exist, we assume the association is invalid
// and we return an empty string.
func (k Keeper) FromTokenToUTokenDenom(ctx sdk.Context, tokenDenom string) string {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateTokenDenomKey(tokenDenom)

	bz := store.Get(key)
	if len(bz) == 0 {
		return ""
	}

	return string(bz)
}

// IsAcceptedToken returns true if a given (non-UToken) token denom is an
// accepted asset type.
func (k Keeper) IsAcceptedToken(ctx sdk.Context, tokenDenom string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateTokenDenomKey(tokenDenom)

	return store.Has(key)
}

// IsAcceptedUToken returns true if a given uToken denom is an accepted asset
// type.
func (k Keeper) IsAcceptedUToken(ctx sdk.Context, uTokenDenom string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateUTokenDenomKey(uTokenDenom)

	return store.Has(key)
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

// TotalTokenSupply returns an sdk.Coin representing the total balance of a
// given asset type if valid. If the denom is not an accepted asset type, we
// return a zero amount.
func (k Keeper) TotalTokenSupply(ctx sdk.Context, tokenDenom string) sdk.Coin {
	if k.IsAcceptedToken(ctx, tokenDenom) {
		return sdk.NewCoin(tokenDenom, sdk.ZeroInt())
	}

	return sdk.NewCoin(tokenDenom, sdk.ZeroInt())
}

// LendAsset attempts to deposit assets into the leverage module account in
// exchange for uTokens. If asset type is invalid or account balance is
// insufficient, we return an error.
func (k Keeper) LendAsset(ctx sdk.Context, msg types.MsgLendAsset) error {
	lenderAddr, err := sdk.AccAddressFromBech32(msg.GetLender())
	if err != nil {
		return err
	}

	loan := msg.GetAmount()
	if !k.IsAcceptedToken(ctx, loan.Denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, loan.String())
	}

	if !k.bankKeeper.HasBalance(ctx, lenderAddr, loan) {
		// lender does not have the assets they intend to lend
		return sdkerrors.Wrap(types.ErrInsufficientBalance, loan.String())
	}

	// send token balance to leverage module account
	loanTokens := sdk.NewCoins(loan)
	if k.bankKeeper.SendCoinsFromAccountToModule(ctx, lenderAddr, types.ModuleName, loanTokens) != nil {
		return err
	}

	// mint uTokens
	// TODO: Use exchange rate instead of 1:1 redeeming
	uToken := sdk.NewCoin(k.FromTokenToUTokenDenom(ctx, loan.Denom), loan.Amount)
	uTokens := sdk.NewCoins(uToken)
	if k.bankKeeper.MintCoins(ctx, types.ModuleName, uTokens) != nil {
		return err
	}

	if k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lenderAddr, uTokens) != nil {
		return err
	}

	return nil
}

// WithdrawAsset attempts to deposit uTokens into the leverage module in exchange for original assets.
// If utoken type is invalid or account balance insufficient on either side, does nothing and returns false.
// TODO: Panic if partially executed then fail?
func (k Keeper) WithdrawAsset(ctx sdk.Context, msg types.MsgWithdrawAsset) bool {
	lender, err := sdk.AccAddressFromBech32(msg.GetLender())
	if err != nil {
		// Could not unmarshal address (Bech32)
		return false
	}
	utokens := msg.GetAmount()
	if !k.IsAcceptedUtoken(ctx, utokens.Denom) {
		// Not an accepted utoken type
		return false
	}
	assetDenom := k.ToBaseAssetDenom(ctx, utokens.Denom)
	if !k.bankKeeper.HasBalance(ctx, lender, utokens) {
		// Lender does not have the uTokens they intend to redeem
		return false
	}
	assets := sdk.NewCoin(assetDenom, utokens.Amount) // TODO: Use exchange rate instead of 1:1 redeeming
	/*
		TODO: once we have leverage module's account address
		if k.bankKeeper.HasBalance(ctx,module_account_address,assets) == false {
			return false
		}
	*/
	// TODO: Add additional checks to be sure tokens/accounts are good to send from before doing the
	// irreversible
	if k.bankKeeper.SendCoinsFromAccountToModule(ctx, lender, "leverage", sdk.Coins{utokens}) != nil {
		// Error depositing utokens
		return false
	}
	if k.bankKeeper.SendCoinsFromModuleToAccount(ctx, "leverage", lender, sdk.Coins{assets}) != nil {
		// Error releasing assets
		// TODO: Either make this a panic, or reverse the sending of uTokens
		return false
	}
	if k.bankKeeper.BurnCoins(ctx, "leverage", sdk.Coins{utokens}) != nil {
		// Error burning utokens
		// TODO: Either make this a panic, or reverse the depositing or utokens and releasing of assets
		return false
	}
	return false
}
