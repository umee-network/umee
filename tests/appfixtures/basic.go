package appfixtures

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v4/app"
	appparams "github.com/umee-network/umee/v4/app/params"
)

type Val struct {
	ValPubKey cryptotypes.PubKey
	Addr      sdk.Address
	ValAddr   sdk.ValAddress
}

func NewVal() Val {
	pubKey := secp256k1.GenPrivKey().PubKey()
	addr := pubKey.Address()

	return Val{simapp.CreateTestPubKeys(1)[0],
		sdk.AccAddress(addr),
		sdk.ValAddress(addr)}
}

type Basic struct {
	BondDenom string
	App       *app.UmeeApp
	Ctx       sdk.Context
	Querier   *baseapp.QueryServiceTestHelper
}

// NewBasic cresats an UmeeApp with a basic setup of 2 validators. Each will receive
// 10000000000 Umee tokens.
func NewBasic(t *testing.T) Basic {
	const initialValPower = int64(10000000000)

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

		initTokens = sdk.TokensFromConsensusPower(initialValPower, sdk.DefaultPowerReduction)
		initUmee   = sdk.NewCoin(appparams.BondDenom, initTokens)
	)

	app := Setup(t)
	isCheckTx := false
	ctx := app.NewContext(isCheckTx, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  2,
	})
	suite := Basic{
		App:     app,
		Ctx:     ctx,
		Querier: baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry()),
	}

	suite.MintCoins(t, addr, initUmee)
	suite.MintCoins(t, addr2, initUmee)

	// create validators and stake tokens
	amt := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	sh := teststaking.NewHelper(t, ctx, *app.StakingKeeper)
	sh.Denom = appparams.BondDenom
	sh.CreateValidator(valAddr, valPubKey, amt, true)
	sh.CreateValidator(valAddr2, valPubKey2, amt, true)

	staking.EndBlocker(ctx, *app.StakingKeeper)

	return suite
}

// MintUmee creates new uumee from mint moudle and sends it to a given account.
func (suite Basic) MintUmee(t *testing.T, recipient sdk.AccAddress, amount int64) {
	MintUmee(t, suite.Ctx, suite.App, recipient, amount)
}

// MintCoins creates new coins from mint moudle and sends it to a given account.
// Uses internal Context.
func (suite Basic) MintCoins(t *testing.T, recipient sdk.AccAddress, coins ...sdk.Coin) {
	assert.NilError(t, suite.App.BankKeeper.MintCoins(suite.Ctx, minttypes.ModuleName, coins))
	assert.NilError(t, suite.App.BankKeeper.SendCoinsFromModuleToAccount(
		suite.Ctx, minttypes.ModuleName, recipient, coins))
}
