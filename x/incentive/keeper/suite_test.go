package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	k                   keeper.Keeper // maybe use TestKeeper?
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
	k := keeper.NewKeeper(
		app.AppCodec(),
		app.GetKey(incentive.ModuleName),
		app.BankKeeper,
		s.mockLeverage,
	)

	s.k = k
	app.IncentiveKeeper = k
	// TODO: if I need to add hooks
	// app.IncentiveKeeper = *app.IncentiveKeeper.SetHooks(types.NewMultiHooks())

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
