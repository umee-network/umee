package migrations

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/x/oracle/keeper"
	"github.com/umee-network/umee/v6/x/oracle/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper *keeper.Keeper
}

// NewMigrator creates a Migrator.
func NewMigrator(keeper *keeper.Keeper) Migrator {
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
	p := types.DefaultParams()

	m.keeper.SetHistoricStampPeriod(ctx, p.HistoricStampPeriod)
	m.keeper.SetMedianStampPeriod(ctx, p.MedianStampPeriod)
	m.keeper.SetMaximumPriceStamps(ctx, p.MaximumPriceStamps)
	m.keeper.SetMaximumMedianStamps(ctx, p.MaximumMedianStamps)
	return nil
}
