package keeper

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util"
	"github.com/umee-network/umee/v4/util/decmath"
	"github.com/umee-network/umee/v4/x/oracle/types"
)

// HistoricMedians returns a list of a given denom's last numStamps medians.

func (k Keeper) HistoricMedians(
	ctx sdk.Context,
	denom string,
	numStamps uint64,
) types.Prices {
	medians := types.Prices{}

	k.IterateHistoricMedians(ctx, denom, uint(numStamps), func(median types.Price) bool {
		medians = append(medians, median)
		return false
	})

	return medians
}

// CalcAndSetHistoricMedian uses all the historic prices of a given denom to
// calculate its median price at the current block and set it to the store.
// It will also call setMedianDeviation with the calculated median.
func (k Keeper) CalcAndSetHistoricMedian(
	ctx sdk.Context,
	denom string,
) error {
	historicPrices := k.historicPrices(ctx, denom, k.MaximumPriceStamps(ctx))
	median, err := decmath.Median(historicPrices)
	if err != nil {
		return errors.Wrap(err, "denom: "+denom)
	}

	block := uint64(ctx.BlockHeight())
	k.SetHistoricMedian(ctx, denom, block, median)
	return k.calcAndSetHistoricMedianDeviation(ctx, denom, median, historicPrices)
}

func (k Keeper) SetHistoricMedian(
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
func (k Keeper) HistoricMedianDeviation(
	ctx sdk.Context,
	denom string,
) (*types.Price, error) {
	store := ctx.KVStore(k.storeKey)
	blockDiff := uint64(ctx.BlockHeight())%k.MedianStampPeriod(ctx) + 1
	blockNum := uint64(ctx.BlockHeight()) - blockDiff
	bz := store.Get(types.KeyMedianDeviation(denom, blockNum))
	if bz == nil {
		return &types.Price{}, types.ErrNoMedianDeviation.Wrap("denom: " + denom)
	}

	decProto := sdk.DecProto{}
	k.cdc.MustUnmarshal(bz, &decProto)
	price := types.NewPrice(decProto.Dec, denom, blockNum)

	return price, nil
}

// WithinHistoricMedianDeviation returns whether or not the current price of a
// given denom is within the latest stamped Standard Deviation around
// the Median.
func (k Keeper) WithinHistoricMedianDeviation(
	ctx sdk.Context,
	denom string,
) (bool, error) {
	// get latest median
	medians := k.HistoricMedians(ctx, denom, 1)
	if len(medians) == 0 {
		return false, types.ErrNoMedian.Wrap("denom: " + denom)
	}
	median := medians[0].ExchangeRateTuple.ExchangeRate

	// get latest historic price
	prices := k.historicPrices(ctx, denom, 1)
	if len(prices) == 0 {
		return false, types.ErrNoHistoricPrice.Wrap("denom: " + denom)
	}
	price := prices[0]

	medianDeviation, err := k.HistoricMedianDeviation(ctx, denom)
	if err != nil {
		return false, err
	}

	return price.Sub(median).Abs().LTE(medianDeviation.ExchangeRateTuple.ExchangeRate), nil
}

// calcAndSetHistoricMedianDeviation calculates and sets a given denom's standard
// deviation around its median price in the current block.
func (k Keeper) calcAndSetHistoricMedianDeviation(
	ctx sdk.Context,
	denom string,
	median sdk.Dec,
	prices []sdk.Dec,
) error {
	medianDeviation, err := decmath.MedianDeviation(median, prices)
	if err != nil {
		return errors.Wrap(err, "denom: "+denom)
	}

	block := uint64(ctx.BlockHeight())
	k.SetHistoricMedianDeviation(ctx, denom, block, medianDeviation)
	return nil
}

func (k Keeper) SetHistoricMedianDeviation(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
	medianDeviation sdk.Dec,
) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: medianDeviation})
	store.Set(types.KeyMedianDeviation(denom, blockNum), bz)
}

// MedianOfHistoricMedians calculates and returns the median of the last stampNum
// historic medians as well as the amount of medians used to calculate that median.
// If no medians are available, all returns are zero and error is nil.
func (k Keeper) MedianOfHistoricMedians(
	ctx sdk.Context,
	denom string,
	numStamps uint64,
) (sdk.Dec, uint32, error) {
	medians := k.HistoricMedians(ctx, denom, numStamps)
	if len(medians) == 0 {
		return sdk.ZeroDec(), 0, nil
	}
	median, err := decmath.Median(medians.Decs())
	if err != nil {
		return sdk.ZeroDec(), 0, errors.Wrap(err, "denom: "+denom)
	}

	return median, uint32(len(medians)), nil
}

// AverageOfHistoricMedians calculates and returns the average of the last stampNum
// historic medians as well as the amount of medians used to calculate that average.
// If no medians are available, all returns are zero and error is nil.
func (k Keeper) AverageOfHistoricMedians(
	ctx sdk.Context,
	denom string,
	numStamps uint64,
) (sdk.Dec, uint32, error) {
	medians := k.HistoricMedians(ctx, denom, numStamps)
	if len(medians) == 0 {
		return sdk.ZeroDec(), 0, nil
	}
	average, err := decmath.Average(medians.Decs())
	if err != nil {
		return sdk.ZeroDec(), 0, errors.Wrap(err, "denom: "+denom)
	}

	return average, uint32(len(medians)), nil
}

// MaxOfHistoricMedian calculates and returns the maximum value of the last stampNum
// historic medians as well as the amount of medians used to calculate that maximum.
// If no medians are available, all returns are zero and error is nil.
func (k Keeper) MaxOfHistoricMedians(
	ctx sdk.Context,
	denom string,
	numStamps uint64,
) (sdk.Dec, uint32, error) {
	medians := k.HistoricMedians(ctx, denom, numStamps)
	if len(medians) == 0 {
		return sdk.ZeroDec(), 0, nil
	}
	max, err := decmath.Max(medians.Decs())
	if err != nil {
		return sdk.ZeroDec(), 0, errors.Wrap(err, "denom: "+denom)
	}

	return max, uint32(len(medians)), nil
}

// MinOfHistoricMedians calculates and returns the minimum value of the last stampNum
// historic medians as well as the amount of medians used to calculate that minimum.
// If no medians are available, all returns are zero and error is nil.
func (k Keeper) MinOfHistoricMedians(
	ctx sdk.Context,
	denom string,
	numStamps uint64,
) (sdk.Dec, uint32, error) {
	medians := k.HistoricMedians(ctx, denom, numStamps)
	if len(medians) == 0 {
		return sdk.ZeroDec(), 0, nil
	}
	min, err := decmath.Min(medians.Decs())
	if err != nil {
		return sdk.ZeroDec(), 0, errors.Wrap(err, "denom: "+denom)
	}

	return min, uint32(len(medians)), nil
}

// historicPrices returns all the historic prices of a given denom.
func (k Keeper) historicPrices(
	ctx sdk.Context,
	denom string,
	numStamps uint64,
) []sdk.Dec {
	// calculate start block to iterate from
	historicPrices := []sdk.Dec{}

	k.IterateHistoricPrices(ctx, denom, uint(numStamps), func(exchangeRate sdk.Dec) bool {
		historicPrices = append(historicPrices, exchangeRate)
		return false
	})

	return historicPrices
}

// IterateHistoricPrices iterates over historic prices of a given
// denom in the store in reverse.
// Iterator stops when exhausting the source, or when the handler returns `true`.
func (k Keeper) IterateHistoricPrices(
	ctx sdk.Context,
	denom string,
	numStamps uint,
	handler func(sdk.Dec) bool,
) {
	store := ctx.KVStore(k.storeKey)

	// make sure we have one zero byte to correctly separate denoms
	prefix := util.ConcatBytes(1, types.KeyPrefixHistoricPrice, []byte(denom))
	iter := sdk.KVStoreReversePrefixIteratorPaginated(store, prefix, 1, numStamps)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		decProto := sdk.DecProto{}
		k.cdc.MustUnmarshal(iter.Value(), &decProto)
		if handler(decProto.Dec) {
			break
		}
	}
}

// IterateHistoricMediansSinceBlock iterates over medians of a given
// denom in the store in reverse.
// Iterator stops when exhausting the source, or when the handler returns `true`.
func (k Keeper) IterateHistoricMedians(
	ctx sdk.Context,
	denom string,
	numStamps uint,
	handler func(types.Price) bool,
) {
	store := ctx.KVStore(k.storeKey)

	// make sure we have one zero byte to correctly separate denoms
	prefix := util.ConcatBytes(1, types.KeyPrefixMedian, []byte(denom))
	iter := sdk.KVStoreReversePrefixIteratorPaginated(store, prefix, 1, numStamps)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		denom, block := types.ParseDenomAndBlockFromKey(iter.Key(), types.KeyPrefixMedian)
		decProto := sdk.DecProto{}
		k.cdc.MustUnmarshal(iter.Value(), &decProto)
		price := types.NewPrice(decProto.Dec, denom, block)
		if handler(*price) {
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
	k.AvgKeeper(ctx).updateAvgCounter(denom, exchangeRate, ctx.BlockTime())
}

func (k Keeper) HistoricAvgPrice(ctx sdk.Context, denom string) (sdk.Dec, error) {
	return k.AvgKeeper(ctx).GetCurrentAvg(denom)
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

// DeleteHistoricMedian deletes a given denom's median price at a given block.
func (k Keeper) DeleteHistoricMedian(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyMedian(denom, blockNum))
}

// DeleteHistoricMedianDeviation deletes a given denom's standard deviation
// around its median price at a given block.
func (k Keeper) DeleteHistoricMedianDeviation(
	ctx sdk.Context,
	denom string,
	blockNum uint64,
) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyMedianDeviation(denom, blockNum))
}

func (k Keeper) PruneHistoricPricesBeforeBlock(ctx sdk.Context, blockNum uint64) {
	k.IterateAllHistoricPrices(ctx, func(price types.Price) (stop bool) {
		if price.BlockNum <= blockNum {
			k.DeleteHistoricPrice(ctx, price.ExchangeRateTuple.Denom, price.BlockNum)
		}
		return false
	})
}

func (k Keeper) PruneMediansBeforeBlock(ctx sdk.Context, blockNum uint64) {
	k.IterateAllMedianPrices(ctx, func(price types.Price) (stop bool) {
		if price.BlockNum <= blockNum {
			k.DeleteHistoricMedian(ctx, price.ExchangeRateTuple.Denom, price.BlockNum)
		}
		return false
	})
}

func (k Keeper) PruneMedianDeviationsBeforeBlock(ctx sdk.Context, blockNum uint64) {
	k.IterateAllMedianDeviationPrices(ctx, func(price types.Price) (stop bool) {
		if price.BlockNum <= blockNum {
			k.DeleteHistoricMedianDeviation(ctx, price.ExchangeRateTuple.Denom, price.BlockNum)
		}
		return false
	})
}
