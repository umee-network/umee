package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/umee-network/umee/x/leverage/types"
)

// ExchangeTokens converts an sdk.Coin containing base assets to its value in uTokens
func (k Keeper) ExchangeTokens(ctx sdk.Context, tokens sdk.Coin) (sdk.Coin, error) {
	uTokenDenom := k.FromTokenToUTokenDenom(ctx, tokens.Denom)
	if uTokenDenom == "" {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, tokens.Denom)
	}
	exchangeRate, err := k.GetExchangeRate(ctx, tokens.Denom)
	if err != nil {
		return sdk.Coin{}, err
	}
	uTokenAmount := tokens.Amount.ToDec().Quo(exchangeRate).TruncateInt()
	return sdk.NewCoin(uTokenDenom, uTokenAmount), nil
}

// ExchangeUTokens converts an sdk.Coin containing uTokens to its value in base assets
func (k Keeper) ExchangeUTokens(ctx sdk.Context, utokens sdk.Coin) (sdk.Coin, error) {
	tokenDenom := k.FromUTokenToTokenDenom(ctx, utokens.Denom)
	if tokenDenom == "" {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, utokens.Denom)
	}
	exchangeRate, err := k.GetExchangeRate(ctx, tokenDenom)
	if err != nil {
		return sdk.Coin{}, err
	}
	tokenAmount := utokens.Amount.ToDec().Mul(exchangeRate).TruncateInt()
	return sdk.NewCoin(tokenDenom, tokenAmount), nil
}

// GetExchangeRate gets the token:uTokenexchange rate for a given base token denom
func (k Keeper) GetExchangeRate(ctx sdk.Context, denom string) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateExchangeRateKey(denom)
	amount := sdk.ZeroDec()
	bz := store.Get(key)
	if bz == nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}
	err := amount.Unmarshal(bz)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return amount, nil
}

// SetExchangeRate sets the token:uTokenexchange rate for a given base token denom
func (k Keeper) SetExchangeRate(ctx sdk.Context, denom string, rate sdk.Dec) error {
	store := ctx.KVStore(k.storeKey)
	if !k.IsAcceptedToken(ctx, denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}
	rateKey := types.CreateExchangeRateKey(denom)
	bz, err := rate.Marshal()
	if err != nil {
		return err
	}
	store.Set(rateKey, bz)
	return nil
}

// UpdateExchangeRates calculates sets the token:uToken exchange rates for all token denoms.
func (k Keeper) UpdateExchangeRates(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)
	totalBorrows, err := k.GetTotalBorrows(ctx)
	if err != nil {
		return err
	}
	// Calculate exchange rates for all denoms which have an exchange rate stored in keeper.
	// If a token is registered but its exchange rate is never initialized (set to 1.0), this
	// iterator will fail to detect it, and its exchange rate will remain undefined.
	exchangeRatePrefix := types.CreateExchangeRateKeyNoDenom()
	iter := sdk.KVStorePrefixIterator(store, exchangeRatePrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		// key is exchangeRatePrefix | denom | 0x00
		key, _ := iter.Key(), iter.Value()
		// remove exchangeRatePrefix and null-terminator
		denom := string(key[len(exchangeRatePrefix) : len(key)-1])
		// Utoken exchange rate is equal to the token supply (including borrowed tokens yet to be repaid
		// and excluding tokens reserved) divided by total uTokens in circulation. Mathematically:
		// tokens:uToken = (module token balance + tokens borrowed - reserved tokens) / uToken supply
		moduleBalance := k.bankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), denom).Amount.ToDec()
		tokenSupply := moduleBalance.Add(totalBorrows.AmountOf(denom).Sub(k.GetReserveAmount(ctx, denom)).ToDec())
		uTokenSupply := k.TotalUTokenSupply(ctx, k.FromTokenToUTokenDenom(ctx, denom)).Amount
		derivedExchangeRate := sdk.OneDec()
		if uTokenSupply.IsPositive() {
			derivedExchangeRate = tokenSupply.QuoInt(uTokenSupply)
		}
		err = k.SetExchangeRate(ctx, denom, derivedExchangeRate)
		if err != nil {
			return err
		}
	}
	return nil
}

// InitializeExchangeRate checks the token:uTokenexchange rate for a given base token denom
// and sets it to 1.0 if no rate has been registered. No-op if a rate already exists.
func (k Keeper) InitializeExchangeRate(ctx sdk.Context, denom string) error {
	store := ctx.KVStore(k.storeKey)
	rateKey := types.CreateExchangeRateKey(denom)
	if !store.Has(rateKey) {
		bz, err := sdk.OneDec().Marshal()
		if err != nil {
			return err
		}
		store.Set(rateKey, bz)
	}
	return nil
}
