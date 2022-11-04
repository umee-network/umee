package keeper

import (
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v3/x/oracle/types"
)

// median returns the median of list of historic prices.
func median(prices []types.HistoricPrice) sdk.Dec {
	lenPrices := len(prices)
	if lenPrices == 0 {
		return sdk.ZeroDec()
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].ExchangeRates.ExchangeRate.BigInt().
			Cmp(prices[j].ExchangeRates.ExchangeRate.BigInt()) > 0
	})

	var medianPrice sdk.Dec
	if lenPrices%2 == 0 {
		medianPrice = prices[lenPrices/2-1].ExchangeRates.ExchangeRate.
			Add(prices[lenPrices/2].ExchangeRates.ExchangeRate).
			QuoInt64(2)
	} else {
		medianPrice = prices[lenPrices/2].ExchangeRates.ExchangeRate
	}

	return medianPrice
}

// GetMedian returns the median price of a given denom.
func (k Keeper) GetMedian(
	ctx sdk.Context,
	denom string,
) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	denom = strings.ToUpper(denom)
	bz := store.Get(types.GetMedianKey(denom))
	if bz == nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrUnknownDenom, denom)
	}

	decProto := sdk.DecProto{}
	k.cdc.MustUnmarshal(bz, &decProto)

	return decProto.Dec, nil
}

// setMedian uses all the historic prices of given denom to set the
// the median price of that denom in the store.
func (k Keeper) setMedian(
	ctx sdk.Context,
	denom string,
) {
	store := ctx.KVStore(k.storeKey)

	historicPrices := k.GetHistoricPrices(ctx, denom)

	median := median(historicPrices)

	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: median})
	denom = strings.ToUpper(denom)
	store.Set(types.GetMedianKey(denom), bz)
}

// GetHistoricPrice returns the historic price of a denom at a given
// block.
func (k Keeper) GetHistoricPrice(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
) (types.HistoricPrice, error) {
	store := ctx.KVStore(k.storeKey)
	denom = strings.ToUpper(denom)
	bz := store.Get(types.GetHistoricPriceKey(denom, blockNum))
	if bz == nil {
		return types.HistoricPrice{}, sdkerrors.Wrap(types.ErrUnknownDenom, denom)
	}

	var historicPrice types.HistoricPrice
	k.cdc.MustUnmarshal(bz, &historicPrice)

	return historicPrice, nil
}

// GetHistoricPrices returns all the historic prices of a given denom.
func (k Keeper) GetHistoricPrices(
	ctx sdk.Context,
	denom string,
) []types.HistoricPrice {
	store := ctx.KVStore(k.storeKey)
	denom = strings.ToUpper(denom)
	historicPrices := []types.HistoricPrice{}

	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixHistoricPrice)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		bz := store.Get(iter.Key())
		var historicPrice types.HistoricPrice
		k.cdc.MustUnmarshal(bz, &historicPrice)

		if historicPrice.ExchangeRates.Denom == denom {
			historicPrices = append(historicPrices, historicPrice)
		}
	}

	return historicPrices
}

// SetHistoricPrice sets the historic price of a denom at the current
// block height when called to the store. Afterwards it will call
// setMedian to update the median price of the denom in the store.
func (k Keeper) SetHistoricPrice(
	ctx sdk.Context,
	denom string,
) {
	store := ctx.KVStore(k.storeKey)
	exchangeRate, _ := k.GetExchangeRateBase(ctx, denom)
	exchangeRateTuple := types.ExchangeRateTuple{
		Denom:        denom,
		ExchangeRate: exchangeRate,
	}

	block := uint64(ctx.BlockHeight())
	bz := k.cdc.MustMarshal(&types.HistoricPrice{
		ExchangeRates: exchangeRateTuple,
		BlockNum:      block,
	})
	denom = strings.ToUpper(denom)
	store.Set(types.GetHistoricPriceKey(denom, block), bz)

	// update median after adding new historic price
	k.setMedian(ctx, denom)
}

// DeleteHistoricPrice deletes the historic price of a denom at a
// given block.
func (k Keeper) DeleteHistoricPrice(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetHistoricPriceKey(denom, blockNum))
}

func (k Keeper) ClearHistoricPrices(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixHistoricPrice)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}
