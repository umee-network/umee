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

// MigrateBNB fixes the BNB base denom for the 4.1 upgrade.
// Also returns a boolean representing whether the token was changed.
func (m Migrator) MigrateBNB(ctx sdk.Context) (bool, error) {
	// Bad BNB token denom
	badDenom := "ibc/77BCD42E49E5B7E0FC6B269FEBF0185B15044F13F6F38CA285DF0AF883459F40"
	// Ensure zero supply of the token being removed from leverage registry
	uSupply := m.keeper.GetUTokenSupply(ctx, types.ToUTokenDenom(badDenom))
	if !uSupply.IsZero() {
		ctx.Logger().Error("can't correctly migrate leverage with existing supply",
			"token", badDenom, "total_u_supply", uSupply)
		return false, nil
	}
	token, err := m.keeper.GetTokenSettings(ctx, badDenom)
	if err != nil {
		ctx.Logger().Error("leverage migration skipped due to missing token", "err", err.Error())
		// If it's not a registered token, then we don't need to run this migration
		return false, nil
	}
	// Delete previous entry in token registry
	store := ctx.KVStore(m.keeper.storeKey)
	store.Delete(types.KeyRegisteredToken(badDenom))
	// Modify base denom and add back to store, bypassing the hooks in SetRegisteredToken
	correctDenom := "ibc/8184469200C5E667794375F5B0EC3B9ABB6FF79082941BF5D0F8FF59FEBA862E"
	token.BaseDenom = correctDenom
	bz, err := m.keeper.cdc.Marshal(&token)
	if err != nil {
		return false, err
	}
	store.Set(types.KeyRegisteredToken(correctDenom), bz)
	return true, nil
}
