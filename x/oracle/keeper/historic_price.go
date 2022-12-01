package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v3/util"
	"github.com/umee-network/umee/v3/x/oracle/types"
)

// median returns the median of a list of prices.
func median(prices []sdk.Dec) sdk.Dec {
	lenPrices := len(prices)
	if lenPrices == 0 {
		return sdk.ZeroDec()
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].BigInt().
			Cmp(prices[j].BigInt()) < 0
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

// average returns the average value of a list of prices
func average(prices []sdk.Dec) sdk.Dec {
	lenPrices := len(prices)
	sumPrices := sdk.ZeroDec()
	if lenPrices == 0 {
		return sdk.ZeroDec()
	}

	for _, price := range prices {
		sumPrices = sumPrices.Add(price)
	}

	return sumPrices.QuoInt64(int64(lenPrices))
}

// max returns the max value of a list of prices
func max(prices []sdk.Dec) sdk.Dec {
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].BigInt().
			Cmp(prices[j].BigInt()) < 0
	})

	return prices[len(prices)-1]
}

// min returns the min value of a list of prices
func min(prices []sdk.Dec) sdk.Dec {
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].BigInt().
			Cmp(prices[j].BigInt()) < 0
	})

	return prices[0]
}

// GetMedian returns a given denom's most recently stamped median price.
func (k Keeper) HistoricMedian(
	ctx sdk.Context,
	denom string,
) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	blockDiff := uint64(ctx.BlockHeight())%k.MedianStampPeriod(ctx) + 1
	bz := store.Get(types.KeyMedian(denom, uint64(ctx.BlockHeight())-blockDiff))
	if bz == nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrNoMedian, fmt.Sprintf("denom: %s", denom))
	}
	median := sdk.DecProto{}
	k.cdc.MustUnmarshal(bz, &median)
	return median.Dec, nil
}

// CalcAndSetMedian uses all the historic prices of a given denom to
// calculate its median price at the current block and set it to the store.
// It will also call setMedianDeviation with the calculated median.
func (k Keeper) CalcAndSetMedian(
	ctx sdk.Context,
	denom string,
) {
	historicPrices := k.getHistoricPrices(ctx, denom)
	median := median(historicPrices)
	block := uint64(ctx.BlockHeight())
	k.SetMedian(ctx, denom, block, median)
	k.calcAndSetMedianDeviation(ctx, denom, median, historicPrices)
}

func (k Keeper) SetMedian(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
	median sdk.Dec,
) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: median})
	store.Set(types.KeyMedian(denom, blockNum), bz)
}

// GetMedianDeviation returns a given denom's most recently stamped standard
// deviation around its median price at a given block.
func (k Keeper) HistoricMedianDeviation(
	ctx sdk.Context,
	denom string,
) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	blockDiff := uint64(ctx.BlockHeight())%k.MedianStampPeriod(ctx) + 1
	bz := store.Get(types.KeyMedianDeviation(denom, uint64(ctx.BlockHeight())-blockDiff))
	if bz == nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrNoMedianDeviation, fmt.Sprintf("denom: %s", denom))
	}

	decProto := sdk.DecProto{}
	k.cdc.MustUnmarshal(bz, &decProto)

	return decProto.Dec, nil
}

// WithinMedianDeviation returns whether or not a given price of a given
// denom is within the latest stamped Standard Deviation around the Median.
func (k Keeper) WithinMedianDeviation(
	ctx sdk.Context,
	denom string,
	price sdk.Dec,
) (bool, error) {
	median, err := k.HistoricMedian(ctx, denom)
	if err != nil {
		return false, err
	}

	medianDeviation, err := k.HistoricMedianDeviation(ctx, denom)
	if err != nil {
		return false, err
	}

	return price.Sub(median).Abs().LTE(medianDeviation), nil
}

// setMedianDeviation sets a given denom's standard deviation around
// its median price in the current block.
func (k Keeper) calcAndSetMedianDeviation(
	ctx sdk.Context,
	denom string,
	median sdk.Dec,
	prices []sdk.Dec,
) {
	medianDeviation := medianDeviation(median, prices)
	block := uint64(ctx.BlockHeight())
	k.SetMedianDeviation(ctx, denom, block, medianDeviation)
}

func (k Keeper) SetMedianDeviation(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
	medianDeviation sdk.Dec,
) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: medianDeviation})
	store.Set(types.KeyMedianDeviation(denom, blockNum), bz)
}

// MedianOfMedians calculates and returns the median of medians in a
// given block range for a given denom. It will not error out, and
// return sdk.ZeroDec if no medians exist in that range for the denom.
func (k Keeper) MedianOfMedians(
	ctx sdk.Context,
	denom string,
	startBlock uint64,
	endBlock uint64,
) sdk.Dec {
	medians := k.getMedians(ctx, denom, startBlock, endBlock)
	return median(medians)
}

// AverageOfMedians calculates and returns the average of medians in a
// given block range for a given denom. It will not error out, and
// return sdk.ZeroDec if no medians exist in that range for the denom.
func (k Keeper) AverageOfMedians(
	ctx sdk.Context,
	denom string,
	startBlock uint64,
	endBlock uint64,
) sdk.Dec {
	medians := k.getMedians(ctx, denom, startBlock, endBlock)
	return average(medians)
}

// MaxMedian calculates and returns the maximum value of medians in a
// given block range for a given denom. It will not error out, and
// return sdk.ZeroDec if no medians exist in that range for the denom.
func (k Keeper) MaxMedian(
	ctx sdk.Context,
	denom string,
	startBlock uint64,
	endBlock uint64,
) sdk.Dec {
	medians := k.getMedians(ctx, denom, startBlock, endBlock)
	return max(medians)
}

// MinMedian calculates and returns the minimum value of medians in a
// given block range for a given denom. It will not error out, and
// return sdk.ZeroDec if no medians exist in that range for the denom.
func (k Keeper) MinMedian(
	ctx sdk.Context,
	denom string,
	startBlock uint64,
	endBlock uint64,
) sdk.Dec {
	medians := k.getMedians(ctx, denom, startBlock, endBlock)
	return min(medians)
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

// getMedians returns all the medians of a given denom.
func (k Keeper) getMedians(
	ctx sdk.Context,
	denom string,
	startBlock uint64,
	endBlock uint64,
) []sdk.Dec {
	medians := []sdk.Dec{}
	k.IterateMedians(ctx, denom, func(exchangeRate sdk.Dec, block uint64) bool {
		if block >= startBlock && block <= endBlock {
			medians = append(medians, exchangeRate)
		}
		return false
	})

	return medians
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

// IterateMedians iterates over median prices of a given
// denom in the store and tracks their block stamp.
// Iterator stops when exhausting the source, or when the handler returns `true`.
func (k Keeper) IterateMedians(
	ctx sdk.Context,
	denom string,
	handler func(sdk.Dec, uint64) bool,
) {
	store := ctx.KVStore(k.storeKey)

	// make sure we have one zero byte to correctly separate denoms
	prefix := util.ConcatBytes(1, types.KeyPrefixMedian, []byte(denom))
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		decProto := sdk.DecProto{}
		k.cdc.MustUnmarshal(iter.Value(), &decProto)
		if handler(decProto.Dec, types.ParseBlockFromMedianKey(iter.Key())) {
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

// DeleteHistoricPrice deletes the historic price of a denom at a
// given block.
func (k Keeper) DeleteHistoricPrice(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyHistoricPrice(denom, blockNum))
}

// DeleteMedian deletes a given denom's median price at a given block.
func (k Keeper) DeleteMedian(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyMedian(denom, blockNum))
}

// DeleteMedianDeviation deletes a given denom's standard deviation around
// its median price at a given block.
func (k Keeper) DeleteMedianDeviation(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyMedianDeviation(denom, blockNum))
}

// ClearMedians iterates through all medians in the store and deletes them.
func (k Keeper) ClearMedians(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixMedian)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// ClearMedianDeviations iterates through all median deviations in the store
// and deletes them.
func (k Keeper) ClearMedianDeviations(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixMedianDeviation)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}
