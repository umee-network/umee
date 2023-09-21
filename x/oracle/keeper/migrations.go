package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/x/oracle/types"
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
	// NOTE: call to m.SetAvgPeriodAndShift is missing here, and caused a chain halt
	// related to v6.0.1.
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

// SetAvgPeriodAndShift updates the avg shift and period params
func (m Migrator) SetAvgPeriodAndShift(ctx sdk.Context) error {
	p := types.DefaultAvgCounterParams()
	return m.keeper.SetHistoricAvgCounterParams(ctx, p)
}

// MigrateBNB fixes the BNB base denom for the 4.1 upgrade without using leverage hooks
func (m Migrator) MigrateBNB(ctx sdk.Context) {
	badDenom := "ibc/77BCD42E49E5B7E0FC6B269FEBF0185B15044F13F6F38CA285DF0AF883459F40"
	correctDenom := "ibc/8184469200C5E667794375F5B0EC3B9ABB6FF79082941BF5D0F8FF59FEBA862E"
	acceptList := m.keeper.AcceptList(ctx)
	for index := range acceptList {
		// Switch the base denom of the token with changing anything else
		if acceptList[index].BaseDenom == badDenom {
			acceptList[index].BaseDenom = correctDenom
		}
	}
	// Overwrite previous accept list
	m.keeper.SetAcceptList(ctx, acceptList)
}
