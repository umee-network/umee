package keeper

import (
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v6/tests/accs"
	"github.com/umee-network/umee/v6/tests/tsdk"
	"github.com/umee-network/umee/v6/x/auction"
	ugovmocks "github.com/umee-network/umee/v6/x/ugov/mocks"
)

// creates keeper without external dependencies (app, leverage etc...)
func initKeeper(t *testing.T) TestKeeper {
	cdc := tsdk.NewCodec(auction.RegisterInterfaces)
	storeKey := storetypes.NewMemoryStoreKey(auction.StoreKey)
	subaccs := SubAccounts{
		RewardsCollect: accs.GenerateAddr("x/auction/rewards"),
	}
	eg := ugovmocks.NewSimpleEmergencyGroupBuilder()
	kb := NewBuilder(cdc, storeKey, subaccs, nil, eg)
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)

	return TestKeeper{kb.Keeper(&ctx), t, &ctx}
}

func initKeeperWithGen(t *testing.T) TestKeeper {
	k := initKeeper(t)
	err := k.InitGenesis(auction.DefaultGenesis())
	require.NoError(t, err)
	return k
}

type TestKeeper struct {
	Keeper
	T   *testing.T
	ctx *sdk.Context
}
