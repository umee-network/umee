package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gotest.tools/v3/assert"

	umeeapp "github.com/umee-network/umee/v4/app"
	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/tests/tsdk"
	"github.com/umee-network/umee/v4/x/uibc"
	"github.com/umee-network/umee/v4/x/uibc/quota/keeper"
)

const (
	displayDenom string = appparams.DisplayDenom
	bondDenom    string = appparams.BondDenom
	initialPower        = int64(10000000000)
)

// Test addresses
var (
	valPubKeys = simapp.CreateTestPubKeys(2)

	valPubKey = valPubKeys[0]
	pubKey    = secp256k1.GenPrivKey().PubKey()
	addr      = sdk.AccAddress(pubKey.Address())
	valAddr   = sdk.ValAddress(pubKey.Address())

	valPubKey2 = valPubKeys[1]
	pubKey2    = secp256k1.GenPrivKey().PubKey()
	addr2      = sdk.AccAddress(pubKey2.Address())
	valAddr2   = sdk.ValAddress(pubKey2.Address())

	initTokens = sdk.TokensFromConsensusPower(initialPower, sdk.DefaultPowerReduction)
	initCoins  = sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, initTokens))

	sampleOutflow = sdk.NewDecCoin("utest", sdk.NewInt(1111))
)

type IntTestSuite struct {
	ctx         sdk.Context
	app         *umeeapp.UmeeApp
	queryClient uibc.QueryClient
	msgServer   uibc.MsgServer
}

func initTestSuite(t *testing.T) *IntTestSuite {
	t.Parallel()
	s := &IntTestSuite{}
	isCheckTx := false
	app := umeeapp.Setup(t)
	ctx := app.NewContext(isCheckTx, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  9,
	})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	uibc.RegisterQueryServer(queryHelper, keeper.NewQuerier(app.UIbcQuotaKeeperB))

	sh := teststaking.NewHelper(t, ctx, *app.StakingKeeper)
	sh.Denom = bondDenom
	amt := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)

	// mint and send coins to validators
	assert.NilError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	assert.NilError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, initCoins))
	assert.NilError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	assert.NilError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr2, initCoins))
	k := app.UIbcQuotaKeeperB.Keeper(&ctx)
	k.SetTokenOutflow(sampleOutflow)

	sh.CreateValidator(valAddr, valPubKey, amt, true)
	sh.CreateValidator(valAddr2, valPubKey2, amt, true)

	staking.EndBlocker(ctx, *app.StakingKeeper)

	s.app = app
	s.ctx = ctx
	s.queryClient = uibc.NewQueryClient(queryHelper)
	s.msgServer = keeper.NewMsgServerImpl(app.UIbcQuotaKeeperB)

	return s
}

// creates keeper with all external dependencies (app, leverage etc...)
func initKeeper(
	t *testing.T,
	cdc codec.BinaryCodec,
	ics4Wrapper porttypes.ICS4Wrapper,
	leverage uibc.Leverage,
	oracle uibc.Oracle,
) (sdk.Context, keeper.Keeper) {
	storeKey := storetypes.NewMemoryStoreKey("quota")
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)
	kb := keeper.NewKeeperBuilder(cdc, storeKey, ics4Wrapper, leverage, oracle)
	return ctx, kb.Keeper(&ctx)
}
