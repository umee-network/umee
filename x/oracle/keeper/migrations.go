package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/oracle/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper *Keeper
}

// NewMigrator creates a Migrator.
func NewMigrator(keeper *Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	m.keeper.SetHistoricStampPeriod(ctx, 1)
	m.keeper.SetMedianStampPeriod(ctx, 1)
	m.keeper.SetMaximumPriceStamps(ctx, 1)
	m.keeper.SetMaximumMedianStamps(ctx, 1)
	return nil
}

// HistoracleParams3x4 updates Historic Params to defaults for the v4.0 upgrade
func (m Migrator) HistoracleParams3x4(ctx sdk.Context) error {
	m.keeper.SetHistoricStampPeriod(ctx, types.DefaultHistoricStampPeriod)
	m.keeper.SetMedianStampPeriod(ctx, types.DefaultMedianStampPeriod)
	m.keeper.SetMaximumPriceStamps(ctx, types.DefaultMaximumPriceStamps)
	m.keeper.SetMaximumMedianStamps(ctx, types.DefaultMaximumMedianStamps)
	return nil
}

// Uppercase4_1 changes all symbol denoms in keeper keys and the accept list to uppercase
func (m Migrator) Uppercase4_1(ctx sdk.Context) error {
	m.keeper.IterateAllHistoricPrices(ctx,
		func(price types.Price) bool {
			// Delete at existing key
			m.keeper.DeleteHistoricPrice(ctx, price.ExchangeRateTuple.Denom, price.BlockNum)
			// Set at uppercase denom key (which may be the same)
			m.keeper.SetHistoricPrice(ctx,
				price.ExchangeRateTuple.Denom, price.BlockNum, price.ExchangeRateTuple.ExchangeRate)
			return false
		},
	)
	m.keeper.IterateAllMedianPrices(ctx,
		func(price types.Price) bool {
			// Delete at existing key
			m.keeper.DeleteHistoricMedian(ctx, price.ExchangeRateTuple.Denom, price.BlockNum)
			// Set at uppercase denom key (which may be the same)
			m.keeper.SetHistoricMedian(ctx,
				price.ExchangeRateTuple.Denom, price.BlockNum, price.ExchangeRateTuple.ExchangeRate)
			return false
		},
	)
	m.keeper.IterateAllMedianDeviationPrices(ctx,
		func(price types.Price) bool {
			// Delete at existing key
			m.keeper.DeleteHistoricMedianDeviation(ctx, price.ExchangeRateTuple.Denom, price.BlockNum)
			// Set at uppercase denom key (which may be the same)
			m.keeper.SetHistoricMedianDeviation(ctx,
				price.ExchangeRateTuple.Denom, price.BlockNum, price.ExchangeRateTuple.ExchangeRate)
			return false
		},
	)
	// Gets the existing accept list, but setter enforces uppercase symbol denoms
	m.keeper.SetAcceptList(ctx, m.keeper.AcceptList(ctx))
	// Note: denoms for exchange rates were already forced to uppercase, so no migration is needed
	//
	// TODO: iterate over AvgKeeper
	//
	return nil
}
