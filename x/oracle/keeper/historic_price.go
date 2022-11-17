package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v3/util"
	"github.com/umee-network/umee/v3/x/oracle/types"
)

// median returns the median of a list of historic prices.
func median(prices []sdk.Dec) sdk.Dec {
	lenPrices := len(prices)
	if lenPrices == 0 {
		return sdk.ZeroDec()
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].BigInt().
			Cmp(prices[j].BigInt()) > 0
	})

	if lenPrices%2 == 0 {
		return prices[lenPrices/2-1].
			Add(prices[lenPrices/2]).
			QuoInt64(2)
	}
	return prices[lenPrices/2]
}

// medianDeviation returns the standard deviation around the
// median of a list of prices.
// medianDeviation = ∑((price - median)^2 / len(prices))
func medianDeviation(median sdk.Dec, prices []sdk.Dec) sdk.Dec {
	lenPrices := len(prices)
	medianDeviation := sdk.ZeroDec()

	for _, price := range prices {
		medianDeviation = medianDeviation.Add(price.
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
) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyMedian(denom))
	if bz == nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrNoMedian, fmt.Sprintf("denom: %s", denom))
	}

	median := sdk.DecProto{}
	k.cdc.MustUnmarshal(bz, &median)
	return median.Dec, nil
}

// SetMedian uses all the historic prices of a given denom to calculate
// its median price in the last prune period since the current block and
// set it to the store. It will also call setMedianDeviation with the
// calculated median.
func (k Keeper) CalcAndSetMedian(
	ctx sdk.Context,
	denom string,
) {
	historicPrices := k.getHistoricPrices(ctx, denom)
	median := median(historicPrices)
	k.SetMedian(ctx, denom, median)
	k.calcAndSetMedianDeviation(ctx, denom, median, historicPrices)
}

func (k Keeper) SetMedian(
	ctx sdk.Context,
	denom string,
	median sdk.Dec,
) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: median})
	store.Set(types.KeyMedian(denom), bz)
}

// GetMedianDeviation returns a given denom's standard deviation around
// its median price in the last prune period since a given block.
func (k Keeper) GetMedianDeviation(
	ctx sdk.Context,
	denom string,
) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyMedianDeviation(denom))
	if bz == nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrNoMedianDeviation, fmt.Sprintf("denom: %s", denom))
	}

	decProto := sdk.DecProto{}
	k.cdc.MustUnmarshal(bz, &decProto)

	return decProto.Dec, nil
}

// setMedianDeviation sets a given denom's standard deviation around
// its median price in the last prune period since the current block.
func (k Keeper) calcAndSetMedianDeviation(
	ctx sdk.Context,
	denom string,
	median sdk.Dec,
	prices []sdk.Dec,
) {
	medianDeviation := medianDeviation(median, prices)
	k.SetMedianDeviation(ctx, denom, medianDeviation)
}

func (k Keeper) SetMedianDeviation(
	ctx sdk.Context,
	denom string,
	medianDeviation sdk.Dec,
) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: medianDeviation})
	store.Set(types.KeyMedianDeviation(denom), bz)
}

// getHistoricPrices returns all the historic prices of a given denom.
func (k Keeper) getHistoricPrices(
	ctx sdk.Context,
	denom string,
) []sdk.Dec {
	historicPrices := []sdk.Dec{}

	k.IterateHistoricPrices(ctx, denom, func(exchangeRate sdk.Dec) bool {
		historicPrices = append(historicPrices, exchangeRate)
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
	handler func(sdk.Dec) bool,
) {
	store := ctx.KVStore(k.storeKey)

	// make sure we have one zero byte to correctly separate denoms
	prefix := util.ConcatBytes(1, types.KeyPrefixHistoricPrice, []byte(denom))
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		decProto := sdk.DecProto{}
		k.cdc.MustUnmarshal(iter.Value(), &decProto)
		if handler(decProto.Dec) {
			break
		}
	}
}

// AddHistoricPrice adds the historic price of a denom at the current
// block height.
func (k Keeper) AddHistoricPrice(
	ctx sdk.Context,
	denom string,
	exchangeRate sdk.Dec,
) {
	block := uint64(ctx.BlockHeight())
	k.SetHistoricPrice(ctx, denom, block, exchangeRate)
}

func (k Keeper) SetHistoricPrice(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
	exchangeRate sdk.Dec,
) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: exchangeRate})
	store.Set(types.KeyHistoricPrice(denom, blockNum), bz)
}

// DeleteHistoricPriceStats deletes the historic price, median price, and
// standard deviation around its median price of a denom at a given block.
func (k Keeper) DeleteHistoricPrice(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyHistoricPrice(denom, blockNum))
}

// DeleteMedian deletes a given denom's median price in the last prune
// period since a given block.
func (k Keeper) DeleteMedian(
	ctx sdk.Context,
	denom string,
) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyMedian(denom))
}

// DeleteMedianDeviation deletes a given denom's standard deviation around
// its median price in the last prune period since a given block.
func (k Keeper) DeleteMedianDeviation(
	ctx sdk.Context,
	denom string,
) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyMedianDeviation(denom))
}
