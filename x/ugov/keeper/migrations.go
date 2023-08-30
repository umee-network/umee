package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v2 "github.com/umee-network/umee/v6/x/ugov/migrations/v2"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	kb Builder
}

// NewMigrator returns a new Migrator instance.
func NewMigrator(kb Builder) Migrator {
	return Migrator{
		kb: kb,
	}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateStore(m.kb.Keeper(&ctx))
}
