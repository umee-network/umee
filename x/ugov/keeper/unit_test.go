package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/tests/tsdk"
	"github.com/umee-network/umee/v4/x/ugov"
)

// creates keeper without external dependencies (app, leverage etc...)
func initKeeper(t *testing.T) TestKeeper {
	ir := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(ir)
	storeKey := storetypes.NewMemoryStoreKey(ugov.StoreKey)
	kb := NewKeeperBuilder(cdc, storeKey)
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)
	return TestKeeper{kb.Keeper(&ctx), t, &ctx}
}

type TestKeeper struct {
	Keeper
	t   *testing.T
	ctx *sdk.Context
}
