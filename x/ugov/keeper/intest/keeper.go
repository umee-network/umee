package intest

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/tests/tsdk"
	"github.com/umee-network/umee/v6/x/ugov"
	"github.com/umee-network/umee/v6/x/ugov/keeper"
)

// MkKeeper initializes ugov.Keeper with no other dependencies.
// Should be used only when no other module is required.
func MkKeeper(t *testing.T) (*sdk.Context, ugov.Keeper) {
	cdc := tsdk.NewCodec(ugov.RegisterInterfaces)
	storeKey := storetypes.NewMemoryStoreKey(ugov.StoreKey)
	kb := keeper.NewBuilder(cdc, storeKey)
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)
	return &ctx, kb.Keeper(&ctx)
}
