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

// MigrateBNB fixes the BNB base denom for the 4.1 upgrade without using leverage hooks
func (m Migrator) MigrateBNB(ctx sdk.Context) error {
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
	return nil
}
