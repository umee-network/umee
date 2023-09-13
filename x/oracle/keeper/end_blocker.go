package keeper

import (
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

	// Deleting the old exchange rates of denoms and keep latest rates
	k.PruneExgRates(ctx, params.HistoricStampPeriod)
}

// IsPeriodLastBlock returns true if we are at the last block of the period
func (k *Keeper) IsPeriodLastBlock(ctx sdk.Context, blocksPerPeriod uint64) bool {
	return (uint64(ctx.BlockHeight())+1)%blocksPerPeriod == 0
}
