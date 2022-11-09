package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v3/x/oracle/types"
)

// median returns the median of a list of historic prices.
func median(prices []types.HistoricPrice) sdk.Dec {
	lenPrices := len(prices)
	if lenPrices == 0 {
		return sdk.ZeroDec()
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].ExchangeRates.ExchangeRate.BigInt().
			Cmp(prices[j].ExchangeRates.ExchangeRate.BigInt()) > 0
	})

	if lenPrices%2 == 0 {
		return prices[lenPrices/2-1].ExchangeRates.ExchangeRate.
			Add(prices[lenPrices/2].ExchangeRates.ExchangeRate).
			QuoInt64(2)
	}
	return prices[lenPrices/2].ExchangeRates.ExchangeRate
}

// medianDeviation returns the standard deviation around the
// median of a list of prices.
// medianDeviation = ∑((price - median)^2 / len(prices))
func medianDeviation(median sdk.Dec, prices []types.HistoricPrice) sdk.Dec {
	lenPrices := len(prices)
	medianDeviation := sdk.ZeroDec()

	for _, price := range prices {
		medianDeviation = medianDeviation.Add(price.ExchangeRates.ExchangeRate.
			Sub(median).Abs().Power(2).
			QuoInt64(int64(lenPrices)))
	}

	return medianDeviation
}

// GetMedian returns a given denom's median price in the last prune
// period since a given block.
func (k Keeper) GetMedian(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetMedianKey(denom, blockNum))
	if bz == nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrUnknownDenom, denom)
	}

	decProto := sdk.DecProto{}
	k.cdc.MustUnmarshal(bz, &decProto)

	return decProto.Dec, nil
}

// SetMedian uses all the historic prices of a given denom to calculate
// its median price in the last prune period since the current block and
// set it to the store. It will also call setMedianDeviation with the
// calculated median.
func (k Keeper) SetMedian(
	ctx sdk.Context,
	denom string,
) {
	store := ctx.KVStore(k.storeKey)
	historicPrices := k.GetHistoricPrices(ctx, denom)
	median := median(historicPrices)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: median})
	store.Set(types.GetMedianKey(denom, uint64(ctx.BlockHeight())), bz)
	k.setMedianDeviation(ctx, denom, median, historicPrices)
}

// GetMedianDeviation returns a given denom's standard deviation around
// its median price in the last prune period since a given block.
func (k Keeper) GetMedianDeviation(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetMedianDeviationKey(denom, blockNum))
	if bz == nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrUnknownDenom, denom)
	}

	decProto := sdk.DecProto{}
	k.cdc.MustUnmarshal(bz, &decProto)

	return decProto.Dec, nil
}

// setMedianDeviation sets a given denom's standard deviation around
// its median price in the last prune period since the current block.
func (k Keeper) setMedianDeviation(
	ctx sdk.Context,
	denom string,
	median sdk.Dec,
	prices []types.HistoricPrice,
) {
	store := ctx.KVStore(k.storeKey)
	medianDeviation := medianDeviation(median, prices)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: medianDeviation})
	store.Set(types.GetMedianDeviationKey(denom, uint64(ctx.BlockHeight())), bz)
}

// GetHistoricPrice returns the historic price of a denom at a given
// block.
func (k Keeper) GetHistoricPrice(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
) (types.HistoricPrice, error) {
	store := ctx.KVStore(k.storeKey)
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
	historicPrices := []types.HistoricPrice{}

	k.IterateHistoricPrices(ctx, denom, func(exchangeRate sdk.Dec, blockNum uint64) bool {
		historicPrices = append(historicPrices, types.HistoricPrice{
			ExchangeRates: types.ExchangeRateTuple{
				Denom:        denom,
				ExchangeRate: exchangeRate,
			},
			BlockNum: blockNum,
		})

		return false
	})

	return historicPrices
}

// IterateHistoricPrices iterates over historic prices of a given
// denom in the store.
// Iterator stops when exhausting the source, or when the handler returns `true`.
func (k Keeper) IterateHistoricPrices(
	ctx sdk.Context,
	denom string,
	handler func(sdk.Dec, uint64) bool,
) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, append(types.KeyPrefixHistoricPrice, []byte(denom)...))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var historicPrice types.HistoricPrice
		k.cdc.MustUnmarshal(iter.Value(), &historicPrice)
		if handler(historicPrice.ExchangeRates.ExchangeRate, historicPrice.BlockNum) {
			break
		}
	}
}

// AddHistoricPrice adds the historic price of a denom at the current
// block height when called to the store. Afterwards it will call
// setMedian to update the median price of the denom in the store.
func (k Keeper) AddHistoricPrice(
	ctx sdk.Context,
	denom string,
	exchangeRate sdk.Dec,
) {
	store := ctx.KVStore(k.storeKey)
	exchangeRateTuple := types.ExchangeRateTuple{
		Denom:        denom,
		ExchangeRate: exchangeRate,
	}

	block := uint64(ctx.BlockHeight())
	bz := k.cdc.MustMarshal(&types.HistoricPrice{
		ExchangeRates: exchangeRateTuple,
		BlockNum:      block,
	})
	store.Set(types.GetHistoricPriceKey(denom, block), bz)
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
