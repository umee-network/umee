package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/x/oracle/types"
)

// IterateAllHistoricPrices iterates over all historic prices.
// Iterator stops when exhausting the source, or when the handler returns `true`.
func (k Keeper) IterateAllHistoricPrices(
	ctx sdk.Context,
	handler func(types.HistoricPrice) bool,
) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixHistoricPrice)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var decProto sdk.DecProto
		k.cdc.MustUnmarshal(iter.Value(), &decProto)
		denom := types.ParseDenomFromKey(iter.Key(), types.KeyPrefixHistoricPrice)
		blockNum := types.ParseBlockFromKey(iter.Key())
		historicPrice := types.HistoricPrice{
			ExchangeRateTuple: types.ExchangeRateTuple{ExchangeRate: decProto.Dec, Denom: denom},
			BlockNum:          blockNum,
		}
		if handler(historicPrice) {
			break
		}
	}
}

// IterateAllMedianPrices iterates over all median prices.
// Iterator stops when exhausting the source, or when the handler returns `true`.
func (k Keeper) IterateAllMedianPrices(
	ctx sdk.Context,
	handler func(types.HistoricPrice) bool,
) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixMedian)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var decProto sdk.DecProto
		k.cdc.MustUnmarshal(iter.Value(), &decProto)
		denom := types.ParseDenomFromKey(iter.Key(), types.KeyPrefixMedian)
		blockNum := types.ParseBlockFromKey(iter.Key())
		median := types.HistoricPrice{
			ExchangeRateTuple: types.ExchangeRateTuple{ExchangeRate: decProto.Dec, Denom: denom},
			BlockNum:          blockNum,
		}

		if handler(median) {
			break
		}
	}
}

// IterateAllMedianDeviationPrices iterates over all median deviation prices.
// Iterator stops when exhausting the source, or when the handler returns `true`.
func (k Keeper) IterateAllMedianDeviationPrices(
	ctx sdk.Context,
	handler func(types.HistoricPrice) bool,
) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixMedianDeviation)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var decProto sdk.DecProto
		k.cdc.MustUnmarshal(iter.Value(), &decProto)
		denom := types.ParseDenomFromKey(iter.Key(), types.KeyPrefixMedianDeviation)
		blockNum := types.ParseBlockFromKey(iter.Key())
		medianDeviation := types.HistoricPrice{
			ExchangeRateTuple: types.ExchangeRateTuple{ExchangeRate: decProto.Dec, Denom: denom},
			BlockNum:          blockNum,
		}

		if handler(medianDeviation) {
			break
		}
	}
}
