package intest

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/umee-network/umee/v6/tests/tsdk"
	"github.com/umee-network/umee/v6/x/ugov"
	"github.com/umee-network/umee/v6/x/ugov/keeper"
)

// MkKeeper initializes ugov.Keeper with no other dependencies.
// Should be used only when no other module is required.
func MkKeeper(t *testing.T) (*sdk.Context, ugov.Keeper) {
	ir := cdctypes.NewInterfaceRegistry()
	ugov.RegisterInterfaces(ir)
	cdc := codec.NewProtoCodec(ir)
	storeKey := storetypes.NewMemoryStoreKey(ugov.StoreKey)
	kb := keeper.NewBuilder(cdc, storeKey, bkeeper.BaseKeeper{}, true)
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)
	return &ctx, kb.Keeper(&ctx)
}
