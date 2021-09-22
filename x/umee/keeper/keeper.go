package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/x/umee/types"
)

type (
	Keeper struct {
		cdc        codec.Codec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		authKeeper types.AuthKeeper
		bankKeeper types.BankKeeper
	}
)

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

// IsAcceptedAsset returns true if a given denom is an accepted asset (not uToken) type.
func (k Keeper) IsAcceptedAsset(ctx sdk.Context, denom string) bool {
	// Has an associated utoken iff it's an accepted asset
	return k.GetAssociatedUtoken(ctx, denom) != ""
}

// IsAcceptedUtoken returns true if a given denom is an accepted uToken (not asset) type.
func (k Keeper) IsAcceptedUtoken(ctx sdk.Context, denom string) bool {
	// Has an associated asset iff it's an accepted utoken
	return k.GetAssociatedAsset(ctx, denom) != ""
}

// TotalUtokenSupply returns an sdk.Coin representing the total balance of a given uToken type if valid.
// If the denom is not an accepted uToken type, the sdk.Coin returned will have a zero amount and the
// input denom.
func (k Keeper) TotalUtokenSupply(ctx sdk.Context, denom string) sdk.Coin {
	if k.IsAcceptedUtoken(ctx, denom) {
		return k.bankKeeper.GetSupply(ctx, denom)
		// Question: Does bank module still track balances sent (locked) via IBC? If it doesn't
		// then the balance returned here would decrease when the tokens are sent off, which is not
		// what we want. In that case, the keeper should keep an sdk.Int total supply for each uToken type.
	}
	return sdk.Coin{
		// Return zero amount on not existing uToken type
		Denom:  denom,
		Amount: sdk.ZeroInt(),
	}
}

// TotalAssetBalance returns an sdk.Coin representing the total balance of a given asset type if valid.
// If the denom is not an accepted asset type, the sdk.Coin returned will have a zero amount and the input
// denom.
func (k Keeper) TotalAssetBalance(ctx sdk.Context, denom string) sdk.Coin {
	if k.IsAcceptedAsset(ctx, denom) {
		return sdk.Coin{
			Denom:  denom,
			Amount: sdk.ZeroInt(), // TODO: Amount = module account total balance of this token type
		}
	}
	return sdk.Coin{
		// Return zero amount on not accepted asset
		Denom:  denom,
		Amount: sdk.ZeroInt(),
	}
}

// GetAssociatedAsset returns the asset type associated with a given uToken. Empty string on non-uToken
// input.
func (k Keeper) GetAssociatedAsset(ctx sdk.Context, denom string) string {
	// TODO: Remove hard-coding, and store this for each uToken denom using keeper
	// 	under UtokenAssociatedAssetPrefix+denom
	if denom == "u/uumee" {
		return "uumee"
	}
	return ""
}

// GetAssociatedUtoken returns the asset type associated with a given asset. Empty string on
// non-accepted-asset input.
func (k Keeper) GetAssociatedUtoken(ctx sdk.Context, denom string) string {
	// TODO: Remove hard-coding, and store this for each accepted asset denom using keeper
	// 	under AssetAssociatedUtokenPrefix+denom
	if denom == "uumee" {
		return "u/uumee"
	}
	return ""

}

// DepositAssets attempts to deposit assets into the leverage module in exchange for uTokens.
// If asset type is invalid or account balance is insufficient, does nothing and returns false.
// TODO: Panic if partially executed then fail?
func (k Keeper) DepositAssets(ctx sdk.Context, lender sdk.AccAddress, assets sdk.Coin) bool {
	if !k.IsAcceptedAsset(ctx, assets.Denom) {
		// Not an accepted asset type for lending
		return false
	}
	utokenDenom := k.GetAssociatedUtoken(ctx, assets.Denom)
	if !k.bankKeeper.HasBalance(ctx, lender, assets) {
		// Lender does not have the assets they intend to lend
		return false
	}
	utokens := sdk.Coin{
		Denom:  utokenDenom,
		Amount: assets.Amount, // TODO: Use exchange rate instead of 1:1 redeeming
	}
	// TODO: Add additional checks to be sure tokens/accounts are good to send from before doing the
	// irreversible
	if k.bankKeeper.MintCoins(ctx, "leverage", sdk.Coins{utokens}) != nil {
		// Error minting utokens
		return false
	}
	if k.bankKeeper.SendCoinsFromAccountToModule(ctx, lender, "leverage", sdk.Coins{assets}) != nil {
		// Error depositing assets
		// TODO: Either make this a panic, or reverse the minting of uTokens above with a burn
		return false
	}
	if k.bankKeeper.SendCoinsFromModuleToAccount(ctx, "leverage", lender, sdk.Coins{utokens}) != nil {
		// Error granting utokens
		// TODO: Either make this a panic, or reverse both the minting of uTokens and the transfer of assets
		return false
	}
	return false

}

// RedeemUtokens attempts to deposit uTokens into the leverage module in exchange for original assets.
// If utoken type is invalid or account balance insufficient on either side, does nothing and returns false.
// TODO: Panic if partially executed then fail?
func (k Keeper) RedeemUtokens(ctx sdk.Context, lender sdk.AccAddress, utokens sdk.Coin) bool {
	if !k.IsAcceptedUtoken(ctx, utokens.Denom) {
		// Not an accepted utoken type
		return false
	}
	assetDenom := k.GetAssociatedAsset(ctx, utokens.Denom)
	if !k.bankKeeper.HasBalance(ctx, lender, utokens) {
		// Lender does not have the uTokens they intend to redeem
		return false
	}
	assets := sdk.Coin{
		Denom:  assetDenom,
		Amount: utokens.Amount, // TODO: Use exchange rate instead of 1:1 redeeming
	}
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
