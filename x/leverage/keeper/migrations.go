package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper *Keeper
}

// NewMigrator creates a Migrator.
func NewMigrator(keeper *Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// MigrateBNB fixes the BNB base denom
func (m Migrator) MigrateBNB(ctx sdk.Context) error {
	badDenom := "ibc/77BCD42E49E5B7E0FC6B269FEBF0185B15044F13F6F38CA285DF0AF883459F40"
	correctDenom := "ibc/8184469200C5E667794375F5B0EC3B9ABB6FF79082941BF5D0F8FF59FEBA862E"
	token, err := m.keeper.GetTokenSettings(ctx, badDenom)
	if err != nil {
		return err
	}
	// Modify base denom
	token.BaseDenom = correctDenom
	// Delete initial entry in token registry
	store := ctx.KVStore(m.keeper.storeKey)
	badKey := types.KeyRegisteredToken(badDenom)
	store.Delete(badKey)
	// Add back to store, but bypass the hooks in SetRegisteredToken
	trueKey := types.KeyRegisteredToken(correctDenom)
	bz, err := m.keeper.cdc.Marshal(&token)
	if err != nil {
		return err
	}
	store.Set(trueKey, bz)
	return nil
}
