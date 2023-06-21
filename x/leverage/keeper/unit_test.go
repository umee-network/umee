package keeper

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/umee-network/umee/v5/tests/tsdk"
	"github.com/umee-network/umee/v5/x/leverage/types"
)

// creates keeper with mock leverage keeper
func newTestKeeper(t *testing.T) testKeeper {
	// codec and store
	cdc := codec.NewProtoCodec(nil)
	storeKey := storetypes.NewMemoryStoreKey(types.StoreKey)
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)
	// keepers
	ok := newMockOracleKeeper()
	bk := newMockBankKeeper()
	// does not initialize a proper paramtypes.Subspace since test keeper overrides
	k := NewKeeper(cdc, storeKey, paramtypes.Subspace{}, &bk, ok, true)
	msrv := NewMsgServerImpl(k)
	// modify genesis
	gen := types.DefaultGenesis()
	gen.LastInterestTime = 1 // initializes last interest time
	k.InitGenesis(ctx, *gen)
	return testKeeper{k, bk, *ok, t, ctx, sdk.ZeroInt(), msrv, gen.Params}
}

type testKeeper struct {
	Keeper
	bk                  mockBankKeeper
	ok                  mockOracleKeeper
	t                   *testing.T
	ctx                 sdk.Context
	setupAccountCounter sdkmath.Int
	msrv                types.MsgServer
	// stores params directly to avoid using paramtypes.Subspace
	params types.Params
}

// SetParams overrides leverage keeper's SetParams
func (k testKeeper) SetParams(_ sdk.Context, params types.Params) {
	k.params = params
}

// GetParams overrides leverage keeper's GetParams
func (k testKeeper) GetParams(ctx sdk.Context) (params types.Params) {
	return k.params
}

// newAccount creates a new account for testing, and funds it with any input tokens.
func (k *testKeeper) newAccount(funds ...sdk.Coin) sdk.AccAddress {
	// create a unique address
	k.setupAccountCounter = k.setupAccountCounter.Add(sdk.OneInt())
	addrStr := fmt.Sprintf("%-20s", "addr"+k.setupAccountCounter.String()+"_______________")
	addr := sdk.AccAddress([]byte(addrStr))
	// we skip accountKeeper SetAccount, because we are using mock bank keeper
	k.bk.FundAccount(addr, funds)
	return addr
}
