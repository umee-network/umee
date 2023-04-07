package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	umeeapp "github.com/umee-network/umee/v4/app"
	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/x/incentive"
	"github.com/umee-network/umee/v4/x/incentive/keeper"
	incentivemodule "github.com/umee-network/umee/v4/x/incentive/module"

	"github.com/umee-network/umee/v4/x/leverage/fixtures"
)

const (
	umeeDenom = appparams.BondDenom
	atomDenom = fixtures.AtomDenom
	daiDenom  = fixtures.DaiDenom
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx                 sdk.Context
	app                 *umeeapp.UmeeApp
	k                   keeper.TestKeeper
	queryClient         incentive.QueryClient
	setupAccountCounter sdkmath.Int
	addrs               []sdk.AccAddress
	msgSrvr             incentive.MsgServer

	mockLeverage *mockLeverageKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	// require := s.Require()
	app := umeeapp.Setup(s.T())
	ctx := app.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
		Time:    time.Unix(0, 0),
	})

	s.mockLeverage = newMockLeverageKeeper()

	// override the Incentive keeper so we can supply a custom mock leverage keeper
	k, tk := keeper.NewTestKeeper(
		app.AppCodec(),
		app.GetKey(incentive.ModuleName),
		app.BankKeeper,
		s.mockLeverage,
	)

	s.k = tk
	app.IncentiveKeeper = k

	// can override default genesis here if needed
	incentivemodule.InitGenesis(ctx, app.IncentiveKeeper, *incentive.DefaultGenesis())

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	incentive.RegisterQueryServer(queryHelper, keeper.NewQuerier(app.IncentiveKeeper))

	s.app = app
	s.ctx = ctx
	s.setupAccountCounter = sdkmath.ZeroInt()
	s.queryClient = incentive.NewQueryClient(queryHelper)
	s.addrs = umeeapp.AddTestAddrsIncremental(app, s.ctx, 1, sdk.NewInt(3000000))
	s.msgSrvr = keeper.NewMsgServerImpl(s.app.IncentiveKeeper)
}

// newAccount creates a new account for testing, and funds it with any input tokens.
func (s *IntegrationTestSuite) newAccount(funds ...sdk.Coin) sdk.AccAddress {
	app, ctx := s.app, s.ctx

	// create a unique address
	s.setupAccountCounter = s.setupAccountCounter.Add(sdk.OneInt())
	addrStr := fmt.Sprintf("%-20s", "addr"+s.setupAccountCounter.String()+"_______________")
	addr := sdk.AccAddress([]byte(addrStr))

	// register the account in AccountKeeper
	acct := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acct)

	s.fundAccount(addr, funds...)

	return addr
}

// fundAccount mints and sends tokens to an account for testing.
func (s *IntegrationTestSuite) fundAccount(addr sdk.AccAddress, funds ...sdk.Coin) {
	app, ctx, require := s.app, s.ctx, s.Require()

	coins := sdk.NewCoins(funds...)
	if !coins.IsZero() {
		// mint and send tokens to account
		require.NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
		require.NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins))
	}
}
