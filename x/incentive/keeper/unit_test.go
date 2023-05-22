package keeper

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/umee-network/umee/v4/tests/tsdk"
	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/incentive"
)

// creates keeper with mock leverage keeper
func newTestKeeper(t *testing.T) testKeeper {
	// codec and store
	cdc := codec.NewProtoCodec(nil)
	storeKey := storetypes.NewMemoryStoreKey(incentive.StoreKey)
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)
	// keepers
	lk := newMockLeverageKeeper()
	bk := newMockBankKeeper()
	k := NewKeeper(cdc, storeKey, &bk, &lk)
	msrv := NewMsgServerImpl(k)
	// modify genesis
	gen := incentive.DefaultGenesis()
	gen.LastRewardsTime = 1 // initializes last reward time
	gen.Params.MaxUnbondings = 10
	gen.Params.UnbondingDuration = 86400
	gen.Params.EmergencyUnbondFee = sdk.MustNewDecFromStr("0.01")
	k.InitGenesis(ctx, *gen)
	return testKeeper{k, bk, lk, t, ctx, sdk.ZeroInt(), msrv}
}

type testKeeper struct {
	Keeper
	bk                  mockBankKeeper
	lk                  mockLeverageKeeper
	t                   *testing.T
	ctx                 sdk.Context
	setupAccountCounter sdkmath.Int
	msrv                incentive.MsgServer
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

// newBondedAccount creates a new account for testing, and bonds a uToken amount to it.
// For accuracy, it first sets the account's mock leverage collateral to that value.
// A MsgBond is used for the bonding step.
func (k *testKeeper) newBondedAccount(funds ...sdk.Coin) sdk.AccAddress {
	// create a unique address
	k.setupAccountCounter = k.setupAccountCounter.Add(sdk.OneInt())
	addrStr := fmt.Sprintf("%-20s", "addr"+k.setupAccountCounter.String()+"_______________")
	addr := sdk.AccAddress([]byte(addrStr))
	// we skip accountKeeper SetAccount, because we are using mock bank keeper
	// first set account's collateral
	for _, uToken := range funds {
		k.lk.setCollateral(addr, uToken.Denom, uToken.Amount.Int64())
	}
	// then bond uTokens, requring no error
	k.mustBond(addr, funds...)
	return addr
}

// mustBond bonds utokens to an account and requires no errors. Use when setting up incentive scenarios.
func (k *testKeeper) mustBond(addr sdk.AccAddress, coins ...sdk.Coin) {
	for _, coin := range coins {
		msg := &incentive.MsgBond{
			Account: addr.String(),
			UToken:  coin,
		}
		_, err := k.msrv.Bond(k.ctx, msg)
		require.NoError(k.t, err, "bond")
	}
}

// mustClaim claims rewards for an account and requires no errors. Use when setting up incentive scenarios.
func (k *testKeeper) mustClaim(addr sdk.AccAddress) {
	msg := &incentive.MsgClaim{
		Account: addr.String(),
	}
	_, err := k.msrv.Claim(k.ctx, msg)
	require.NoError(k.t, err, "claim")
}

// mustBeginUnbond unbonds utokens from an account and requires no errors. Use when setting up incentive scenarios.
func (k *testKeeper) mustBeginUnbond(addr sdk.AccAddress, coins ...sdk.Coin) {
	for _, coin := range coins {
		msg := &incentive.MsgBeginUnbonding{
			Account: addr.String(),
			UToken:  coin,
		}
		_, err := k.msrv.BeginUnbonding(k.ctx, msg)
		require.NoError(k.t, err, "begin unbonding")
	}
}

// initCommunityFund funds the mock bank keeper's distribution module account with some tokens
func (k *testKeeper) initCommunityFund(funds ...sdk.Coin) {
	k.bk.FundModule(disttypes.ModuleName, funds)
}

// addIncentiveProgram used MsgGovCreateProgram to create and fund (if community funded) an incentive program.
// and requires no errors. Use when setting up incentive scenarios.
func (k *testKeeper) addIncentiveProgram(uDenom string, start, duration int64, reward sdk.Coin, fromCommunity bool,
) {
	// govAccAddr := s.app.GovKeeper.GetGovernanceAccount(ctx).GetAddress().String()
	govAccAddr := "govAcct"

	program := incentive.IncentiveProgram{
		ID:               0,
		StartTime:        start,
		Duration:         duration,
		UToken:           uDenom,
		Funded:           false,
		TotalRewards:     reward,
		RemainingRewards: coin.Zero(reward.Denom),
	}

	validMsg := &incentive.MsgGovCreatePrograms{
		Authority:         govAccAddr,
		Programs:          []incentive.IncentiveProgram{program},
		FromCommunityFund: fromCommunity,
	}
	_, err := k.msrv.GovCreatePrograms(k.ctx, validMsg)
	require.Nil(k.t, err, "addIncentiveProgram")
}

// sponsor a program with tokens and require no errors. Use when setting up incentive scenarios.
func (k *testKeeper) sponsor(addr sdk.AccAddress, programID uint32) {
	msg := &incentive.MsgSponsor{
		Sponsor: addr.String(),
		Program: programID,
	}
	_, err := k.msrv.Sponsor(k.ctx, msg)
	require.NoError(k.t, err, "sponsor program", programID)
}

// advanceTime runs the functions normally contained in EndBlocker from the current rewards time
// to a target time a fixed duration later. Requires non-negative duration and an initialized lastRewardsTime.
func (k *testKeeper) advanceTime(duration int64) {
	prevTime := k.GetLastRewardsTime(k.ctx)
	k.advanceTimeTo(prevTime + duration)
}

// advanceTimeTo runs the functions normally contained in EndBlocker from the current rewards time
// to a target time. Requires non-negative duration and an initialized lastRewardsTime.
func (k *testKeeper) advanceTimeTo(blockTime int64) {
	// ensure initialized
	prevTime := k.GetLastRewardsTime(k.ctx)
	if prevTime <= 0 {
		panic("last rewards time not initialized")
	}

	// update block time in testkeeper's context, so endBlock will read it
	utcTime := time.Unix(blockTime, 0)
	k.ctx = k.ctx.WithBlockTime(utcTime)

	// ensure that last rewards time has actually updated without errors
	skipped, err := k.EndBlock(k.ctx)
	require.Nil(k.t, err, "endBlock")
	require.False(k.t, skipped, "endBlock skipped")
	require.Equal(k.t, blockTime, k.GetLastRewardsTime(k.ctx))
}

// getProgram gets an incentive program by ID and requires no error
func (k *testKeeper) getProgram(programID uint32) incentive.IncentiveProgram {
	program, _, err := k.getIncentiveProgram(k.ctx, programID)
	require.NoError(k.t, err, "get program", programID)
	return program
}

// programStatus checks whether an incentive program's status. Also asserts that
// the program exists.
func (k *testKeeper) programStatus(programID uint32) incentive.ProgramStatus {
	_, status, err := k.getIncentiveProgram(k.ctx, programID)
	require.NoError(k.t, err, "get program (programStatus)", programID)
	return status
}

// programFunded checks whether an incentive program is funded. Also asserts that
// the program exists and that its funded status is not contradictory with its rewards.
func (k *testKeeper) programFunded(programID uint32) bool {
	program, _, err := k.getIncentiveProgram(k.ctx, programID)
	require.NoError(k.t, err, "get program (programFunded)", programID)
	if !program.Funded {
		require.Equal(k.t, program.RemainingRewards.Amount, sdk.ZeroInt(),
			"non-funded must have zero remaining rewards. program id", programID)
	}
	return program.Funded
}

// initScenario1 creates a scenario with upcoming, ongoing, and completed incentive
// programs as well as a bonded account with ongoing unbondings and both claimed
// and pending rewards. Creates a complex state for genesis and query tests.
func (k *testKeeper) initScenario1() sdk.AccAddress {
	// init a community fund with 1000 UMEE and 10 ATOM available for funding
	k.initCommunityFund(
		coin.New(umee, 1000_000000),
		coin.New(atom, 100_000000),
	)

	// init a supplier with bonded uTokens
	alice := k.newBondedAccount(
		coin.New(u_umee, 100_000000),
		coin.New(u_atom, 50_000000),
	)
	// create some in-progress unbondings
	k.advanceTimeTo(90)
	k.mustBeginUnbond(alice, coin.New(u_umee, 5_000000))
	k.mustBeginUnbond(alice, coin.New(u_umee, 5_000000))
	k.mustBeginUnbond(alice, coin.New(u_atom, 5_000000))

	// create three separate programs, designed to be upcoming, ongoing, and completed at t=100
	k.addIncentiveProgram(u_umee, 10, 20, sdk.NewInt64Coin(umee, 10_000000), true)
	k.addIncentiveProgram(u_umee, 90, 20, sdk.NewInt64Coin(umee, 10_000000), true)
	k.addIncentiveProgram(u_umee, 140, 20, sdk.NewInt64Coin(umee, 10_000000), true)

	// start programs and claim some rewards to set nonzero reward trackers
	k.advanceTimeTo(95)
	k.mustClaim(alice)
	k.advanceTimeTo(100)

	return alice
}
