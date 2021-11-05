package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/umee-network/umee/x/leverage/types"
)

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
		moduleBalance := sdk.NewDecFromInt(
			k.bankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), denom).Amount,
		)
		adjustedTokenSupply := moduleBalance.Add(
			sdk.NewDecFromInt(totalBorrows.AmountOf(denom).Sub(k.GetReserveAmount(ctx, denom))),
		)
		derivedExchangeRate := adjustedTokenSupply.QuoInt(
			k.TotalUTokenSupply(ctx, k.FromTokenToUTokenDenom(ctx, denom)).Amount,
		)
		err = k.SetExchangeRate(ctx, denom, derivedExchangeRate)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetExchangeRate gets the token:uTokenexchange rate for a given base token denom
func (k Keeper) GetExchangeRate(ctx sdk.Context, denom string) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateExchangeRateKey(denom)
	amount := sdk.ZeroDec()
	bz := store.Get(key)
	if bz != nil {
		err := amount.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
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
		panic(err)
	}
	store.Set(rateKey, bz)
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
			panic(err)
		}
		store.Set(rateKey, bz)
	}
	return nil
}
