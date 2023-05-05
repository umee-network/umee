package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/tests/tsdk"
)

// creates keeper with all external dependencies (app, leverage etc...)
func initFullKeeper(t *testing.T, cdc codec.Codec) (sdk.Context, Keeper) {
	storeKey := storetypes.NewMemoryStoreKey("metoken")
	k := NewKeeper(cdc, storeKey)
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)
	return ctx, k
}
