package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v3/util"
	"github.com/umee-network/umee/v3/x/oracle/types"
)

// HistoricMedians returns a list of a given denom's last numStamps medians.
func (k Keeper) GetHistoricMedians(
	ctx sdk.Context,
	denom string,
	numStamps uint64,
) []sdk.Dec {
	// calculate start block to iterate from
	blockPeriod := (numStamps - 1) * k.MedianStampPeriod(ctx)
	lastStampBlock := uint64(ctx.BlockHeight()) - (uint64(ctx.BlockHeight())%k.MedianStampPeriod(ctx) + 1)
	startBlock := lastStampBlock - blockPeriod

	medians := []sdk.Dec{}

	k.IterateMediansSinceBlock(ctx, denom, startBlock, func(median sdk.Dec) bool {
		medians = append(medians, median)
		return false
	})

	return medians
}

// CalcAndSetMedian uses all the historic prices of a given denom to
// calculate its median price at the current block and set it to the store.
// It will also call setMedianDeviation with the calculated median.
func (k Keeper) CalcAndSetMedian(
	ctx sdk.Context,
	denom string,
) error {
	historicPrices := k.getHistoricPrices(ctx, denom)
	median, err := util.Median(historicPrices)
	if err != nil {
		return sdkerrors.Wrap(err, fmt.Sprintf("denom: %s", denom))
	}

	block := uint64(ctx.BlockHeight())
	k.SetMedian(ctx, denom, block, median)
	return k.calcAndSetMedianDeviation(ctx, denom, median, historicPrices)
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

// HistoricMedianDeviation returns a given denom's most recently stamped
// standard deviation around its median price at a given block.
func (k Keeper) GetHistoricMedianDeviation(
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
	// get latest median
	medians := k.GetHistoricMedians(
		ctx,
		denom,
		1,
	)
	if len(medians) == 0 {
		return false, sdkerrors.Wrap(types.ErrNoMedian, fmt.Sprintf("denom: %s", denom))
	}

	medianDeviation, err := k.GetHistoricMedianDeviation(ctx, denom)
	if err != nil {
		return false, err
	}

	return price.Sub(medians[0]).Abs().LTE(medianDeviation), nil
}

// setMedianDeviation sets a given denom's standard deviation around
// its median price in the current block.
func (k Keeper) calcAndSetMedianDeviation(
	ctx sdk.Context,
	denom string,
	median sdk.Dec,
	prices []sdk.Dec,
) error {
	medianDeviation, err := util.MedianDeviation(median, prices)
	if err != nil {
		return sdkerrors.Wrap(err, fmt.Sprintf("denom: %s", denom))
	}

	block := uint64(ctx.BlockHeight())
	k.SetMedianDeviation(ctx, denom, block, medianDeviation)
	return nil
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

// MedianOfMedians calculates and returns the median of the last stampNum
// medians.
func (k Keeper) GetMedianOfMedians(
	ctx sdk.Context,
	denom string,
	numStamps uint64,
) (sdk.Dec, error) {
	medians := k.GetHistoricMedians(ctx, denom, numStamps)
	median, err := util.Median(medians)
	if err != nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(err, fmt.Sprintf("denom: %s", denom))
	}
	return median, nil
}

// AverageOfMedians calculates and returns the average of the last stampNum
// medians.
func (k Keeper) GetAverageOfMedians(
	ctx sdk.Context,
	denom string,
	numStamps uint64,
) (sdk.Dec, error) {
	medians := k.GetHistoricMedians(ctx, denom, numStamps)
	average, err := util.Average(medians)
	if err != nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(err, fmt.Sprintf("denom: %s", denom))
	}
	return average, nil
}

// MaxMedian calculates and returns the maximum value of the last stampNum
// medians.
func (k Keeper) GetMaxOfMedians(
	ctx sdk.Context,
	denom string,
	numStamps uint64,
) (sdk.Dec, error) {
	medians := k.GetHistoricMedians(ctx, denom, numStamps)
	max, err := util.Max(medians)
	if err != nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(err, fmt.Sprintf("denom: %s", denom))
	}
	return max, nil
}

// MinMedian calculates and returns the minimum value of the last stampNum
// medians.
func (k Keeper) GetMinOfMedians(
	ctx sdk.Context,
	denom string,
	numStamps uint64,
) (sdk.Dec, error) {
	medians := k.GetHistoricMedians(ctx, denom, numStamps)
	min, err := util.Min(medians)
	if err != nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(err, fmt.Sprintf("denom: %s", denom))
	}
	return min, nil
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

// IterateMediansSinceBlock iterates over medians of a given
// denom in the store in reverse starting from the latest median
// until startBlock.
// Iterator stops when exhausting the source, or when the handler returns `true`.
func (k Keeper) IterateMediansSinceBlock(
	ctx sdk.Context,
	denom string,
	startBlock uint64,
	handler func(sdk.Dec) bool,
) {
	store := ctx.KVStore(k.storeKey)

	// make sure we have one zero byte to correctly separate denoms
	prefix := util.ConcatBytes(1, types.KeyPrefixMedian, []byte(denom))
	iter := sdk.KVStoreReversePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		decProto := sdk.DecProto{}
		k.cdc.MustUnmarshal(iter.Value(), &decProto)
		block := types.ParseBlockFromKey(iter.Key())
		if handler(decProto.Dec) || block <= startBlock {
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
