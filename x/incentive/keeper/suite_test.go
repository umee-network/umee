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

	"github.com/umee-network/umee/v4/util/coin"
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

	// can override default genesis here if needed - in our case, we will set initial unix time to 1
	gen := incentive.DefaultGenesis()
	gen.LastRewardsTime = 1
	incentivemodule.InitGenesis(ctx, app.IncentiveKeeper, *gen)

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

// newBondedAccount creates a new account for testing, and bonds a uToken amount to it.
// For accuracy, it first sets the account's mock leverage collateral to that value.
// A MsgBond is used for the bonding step.
func (s *IntegrationTestSuite) newBondedAccount(funds ...sdk.Coin) sdk.AccAddress {
	app, ctx := s.app, s.ctx

	// create a unique address
	s.setupAccountCounter = s.setupAccountCounter.Add(sdk.OneInt())
	addrStr := fmt.Sprintf("%-20s", "addr"+s.setupAccountCounter.String()+"_______________")
	addr := sdk.AccAddress([]byte(addrStr))

	// register the account in AccountKeeper
	acct := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acct)

	// first set account's collateral
	for _, uToken := range funds {
		s.mockLeverage.setCollateral(addr, uToken.Denom, uToken.Amount.Int64())
	}
	// then bond uTokens, requring no error
	s.bond(addr, funds...)

	return addr
}

// advanceTime runs the functions normally contained in EndBlocker with a fixed time elapsed.
// requires nonzero lastRewardsTime and a positive duration. Requires no error.
func (s *IntegrationTestSuite) advanceTime(duration int64) {
	k, ctx, require := s.k, s.ctx, s.Require()
	if duration <= 0 {
		panic("advanceTime needs positive duration")
	}

	prevTime := k.GetLastRewardsTime(ctx)
	if prevTime <= 0 {
		panic("last rewards time not initialized")
	}

	// simulate new block time to target an exact time elapsed
	blockTime := prevTime + duration
	require.Nil(k.UpdateRewards(ctx, prevTime, blockTime), "update rewards")
	require.Nil(k.UpdatePrograms(ctx, blockTime), "update programs")
}

// advanceTimeTo runs the functions normally contained in EndBlocker from the current rewards time
// to a target time. Requires positive duration and an initialized lastRewardsTime.
func (s *IntegrationTestSuite) advanceTimeTo(end int64) {
	k, ctx := s.k, s.ctx
	prevTime := k.GetLastRewardsTime(ctx)
	s.advanceTime(end - prevTime)
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

// initCommunityFund creates and funds an account, then sets it as the module's community fund
// newAccount creates a new account for testing, and funds it with any input tokens.
func (s *IntegrationTestSuite) initCommunityFund(funds ...sdk.Coin) sdk.AccAddress {
	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// create and fund account
	addr := s.newAccount(funds...)

	// change only the community fund address in params
	params := s.k.GetParams(ctx)
	params.CommunityFundAddress = addr.String()

	// set params and expect no error
	validMsg := &incentive.MsgGovSetParams{
		Authority:   app.GovKeeper.GetGovernanceAccount(s.ctx).GetAddress().String(),
		Title:       "Update Params",
		Description: "New valid values",
		Params:      params,
	}
	_, err := srv.GovSetParams(ctx, validMsg)
	require.Nil(err, "init community fund")

	return addr
}

// addIncentiveProgram used MsgGovCreateProgram to create and fund (if community funded) an incentive program.
// requires no errors, and returns the ID of the program created.
func (s *IntegrationTestSuite) addIncentiveProgram(uDenom string, start, duration int64, reward sdk.Coin, fromCommunity bool) uint32 {
	ctx, require, srv := s.ctx, s.Require(), s.msgSrvr
	govAccAddr := s.app.GovKeeper.GetGovernanceAccount(ctx).GetAddress().String()

	program := incentive.IncentiveProgram{
		ID:               0,
		StartTime:        start,
		Duration:         duration,
		UToken:           uDenom,
		Funded:           false,
		TotalRewards:     reward,
		RemainingRewards: coin.Zero(reward.Denom),
	}

	id := s.k.GetNextProgramID(ctx)

	// add program and expect no error
	validMsg := &incentive.MsgGovCreatePrograms{
		Authority:         govAccAddr,
		Title:             "Add valid program",
		Description:       "using addIncentiveProgram helper function",
		Programs:          []incentive.IncentiveProgram{program},
		FromCommunityFund: fromCommunity,
	}
	// pass and optionally fund the program from community fund
	_, err := srv.GovCreatePrograms(ctx, validMsg)
	require.Nil(err, "addIncentiveProgram")
	return id
}

// bond utokens to an account and require no errors. Use when setting up incentive scenarios.
func (s *IntegrationTestSuite) bond(addr sdk.AccAddress, coins ...sdk.Coin) {
	srv, ctx, require := s.msgSrvr, s.ctx, s.Require()

	for _, coin := range coins {
		msg := &incentive.MsgBond{
			Account: addr.String(),
			UToken:  coin,
		}
		_, err := srv.Bond(ctx, msg)
		require.NoError(err, "bond")
	}
}

// sponsor a program with tokens and require no errors. Use when setting up incentive scenarios.
func (s *IntegrationTestSuite) sponsor(addr sdk.AccAddress, programID uint32) {
	srv, ctx, require := s.msgSrvr, s.ctx, s.Require()

	program, _, err := s.k.GetIncentiveProgram(ctx, programID)
	require.NoError(err, "get program")

	msg := &incentive.MsgSponsor{
		Sponsor: addr.String(),
		Program: programID,
		Asset:   program.TotalRewards,
	}
	_, err = srv.Sponsor(ctx, msg)
	require.NoError(err, "sponsor program", programID)
}

// getProgram gets an incentive program by ID and requires no error
func (s *IntegrationTestSuite) getProgram(programID uint32) incentive.IncentiveProgram {
	ctx, require := s.ctx, s.Require()
	program, _, err := s.k.GetIncentiveProgram(ctx, programID)
	require.NoError(err, "get program", programID)
	return program
}

// programStatus checks whether an incentive program's status. Also asserts that
// the program exists.
func (s *IntegrationTestSuite) programStatus(programID uint32) incentive.ProgramStatus {
	ctx, require := s.ctx, s.Require()
	_, status, err := s.k.GetIncentiveProgram(ctx, programID)
	require.NoError(err, "get program (programStatus)", programID)
	return status
}

// programFunded checks whether an incentive program is funded. Also asserts that
// the program exists and that its funded status is not contradictory with its rewards.
func (s *IntegrationTestSuite) programFunded(programID uint32) bool {
	ctx, require := s.ctx, s.Require()
	program, _, err := s.k.GetIncentiveProgram(ctx, programID)
	require.NoError(err, "get program (programFunded)", programID)
	if !program.Funded {
		require.Equal(program.RemainingRewards.Amount, sdk.ZeroInt(),
			"non-funded must have zero remaining rewards. program id", programID)
	}
	return program.Funded
}
