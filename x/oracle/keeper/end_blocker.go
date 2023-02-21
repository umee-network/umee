package keeper

import (
	"github.com/armon/go-metrics"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PruneAllPrices deletes all historic prices, medians, and median deviations
// outside pruning period determined by the stamp period multiplied by the maximum stamps.
func (k *Keeper) PruneAllPrices(ctx sdk.Context) {
	params := k.GetParams(ctx)
	blockHeight := uint64(ctx.BlockHeight())

	if k.IsPeriodLastBlock(ctx, params.HistoricStampPeriod) {
		pruneHistoricPeriod := params.HistoricStampPeriod * params.MaximumPriceStamps
		if pruneHistoricPeriod < blockHeight {
			k.PruneHistoricPricesBeforeBlock(ctx, blockHeight-pruneHistoricPeriod)
		}

		if k.IsPeriodLastBlock(ctx, params.MedianStampPeriod) {
			pruneMedianPeriod := params.MedianStampPeriod * params.MaximumMedianStamps
			if pruneMedianPeriod < blockHeight {
				k.PruneMediansBeforeBlock(ctx, blockHeight-pruneMedianPeriod)
				k.PruneMedianDeviationsBeforeBlock(ctx, blockHeight-pruneMedianPeriod)
			}
		}
	}
}

// IsPeriodLastBlock returns true if we are at the last block of the period
func (k *Keeper) IsPeriodLastBlock(ctx sdk.Context, blocksPerPeriod uint64) bool {
	return (uint64(ctx.BlockHeight())+1)%blocksPerPeriod == 0
}

func (k *Keeper) RecordEndBlockMetrics(ctx sdk.Context) {
	if !k.telemetryEnabled {
		return
	}

	k.IterateMissCounters(ctx, func(operator sdk.ValAddress, missCounter uint64) bool {
		metrics.SetGaugeWithLabels(
			[]string{"miss_counter"},
			float32(missCounter),
			[]metrics.Label{{Name: "address", Value: operator.String()}},
		)
		return false
	})

	medians := k.AllMedianPrices(ctx)
	medians = *medians.FilterByBlock(medians.NewestBlockNum())
	for _, median := range medians {
		metrics.SetGaugeWithLabels(
			[]string{"median"},
			float32(median.ExchangeRateTuple.ExchangeRate.MustFloat64()),
			[]metrics.Label{{Name: "denom", Value: median.ExchangeRateTuple.Denom}},
		)
	}

	medianDeviations := k.AllMedianDeviationPrices(ctx)
	medianDeviations = *medianDeviations.FilterByBlock(medians.NewestBlockNum())
	for _, medianDeviation := range medianDeviations {
		metrics.SetGaugeWithLabels(
			[]string{"median_deviation"},
			float32(medianDeviation.ExchangeRateTuple.ExchangeRate.MustFloat64()),
			[]metrics.Label{{Name: "denom", Value: medianDeviation.ExchangeRateTuple.Denom}},
		)
	}
}
