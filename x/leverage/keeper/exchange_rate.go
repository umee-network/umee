package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// ExchangeToken converts an sdk.Coin containing a base asset to its value as a
// uToken.
func (k Keeper) ExchangeToken(ctx sdk.Context, token sdk.Coin) (sdk.Coin, error) {
	if !token.IsValid() {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, token.String())
	}

	uTokenDenom := k.FromTokenToUTokenDenom(ctx, token.Denom)
	if uTokenDenom == "" {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, token.Denom)
	}

	exchangeRate, err := k.GetExchangeRate(ctx, token.Denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	uTokenAmount := token.Amount.ToDec().Quo(exchangeRate).TruncateInt()
	return sdk.NewCoin(uTokenDenom, uTokenAmount), nil
}

// ExchangeUToken converts an sdk.Coin containing a uToken to its value in a base
// token.
func (k Keeper) ExchangeUToken(ctx sdk.Context, uToken sdk.Coin) (sdk.Coin, error) {
	if !uToken.IsValid() {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, uToken.String())
	}

	tokenDenom := k.FromUTokenToTokenDenom(ctx, uToken.Denom)
	if tokenDenom == "" {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, uToken.Denom)
	}

	exchangeRate, err := k.GetExchangeRate(ctx, tokenDenom)
	if err != nil {
		return sdk.Coin{}, err
	}

	tokenAmount := uToken.Amount.ToDec().Mul(exchangeRate).TruncateInt()
	return sdk.NewCoin(tokenDenom, tokenAmount), nil
}

// GetExchangeRate gets the token:uTokenexchange rate for a given base token
// denom.
func (k Keeper) GetExchangeRate(ctx sdk.Context, denom string) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateExchangeRateKey(denom)

	bz := store.Get(key)
	if bz == nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	amount := sdk.ZeroDec()
	if err := amount.Unmarshal(bz); err != nil {
		return sdk.ZeroDec(), err
	}

	return amount, nil
}

// SetExchangeRate sets the token:uTokenexchange rate for a given base token
// denom.
func (k Keeper) SetExchangeRate(ctx sdk.Context, denom string, rate sdk.Dec) error {
	store := ctx.KVStore(k.storeKey)
	if !k.IsAcceptedToken(ctx, denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	bz, err := rate.Marshal()
	if err != nil {
		return err
	}

	store.Set(types.CreateExchangeRateKey(denom), bz)
	return nil
}

// UpdateExchangeRates calculates sets the token:uToken exchange rates for all
// token denoms.
func (k Keeper) UpdateExchangeRates(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)
	totalBorrows, err := k.GetTotalBorrows(ctx)
	if err != nil {
		return err
	}

	exchangeRatePrefix := types.CreateExchangeRateKeyNoDenom()

	iter := sdk.KVStorePrefixIterator(store, exchangeRatePrefix)
	defer iter.Close()

	// Calculate exchange rates for all denoms which have an exchange rate stored
	// in the keeper. If a token is registered but it's exchange rate is never
	// initialized (set to 1.0), this iterator will fail to detect it, and its
	// exchange rate will remain undefined.
	for ; iter.Valid(); iter.Next() {
		// key is exchangeRatePrefix | denom | 0x00
		key, _ := iter.Key(), iter.Value()

		// remove exchangeRatePrefix and null-terminator
		denom := string(key[len(exchangeRatePrefix) : len(key)-1])

		// uToken exchange rate is equal to the token supply (including borrowed
		// tokens yet to be repaid and excluding tokens reserved) divided by total
		// uTokens in circulation.
		//
		// Mathematically:
		// tokens:uToken = (module token balance + tokens borrowed - reserved tokens) / uToken supply
		moduleBalance := k.ModuleBalance(ctx, denom).ToDec()
		tokenSupply := moduleBalance.Add(totalBorrows.AmountOf(denom).Sub(k.GetReserveAmount(ctx, denom)).ToDec())
		uTokenSupply := k.TotalUTokenSupply(ctx, k.FromTokenToUTokenDenom(ctx, denom)).Amount
		derivedExchangeRate := sdk.OneDec()

		if uTokenSupply.IsPositive() {
			derivedExchangeRate = tokenSupply.QuoInt(uTokenSupply)
		}

		if err := k.SetExchangeRate(ctx, denom, derivedExchangeRate); err != nil {
			return err
		}
	}

	return nil
}

// InitializeExchangeRate checks the token:uTokenexchange rate for a given base
// token denom and sets it to 1.0 if no rate has been registered. No-op if a
// rate already exists.
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
