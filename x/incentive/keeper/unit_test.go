package keeper

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

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
	// modify genesis if necessary
	gen := incentive.DefaultGenesis()
	gen.LastRewardsTime = 1 // initializes last reward time
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

// initCommunityFund creates and funds an account, then sets it as the module's community fund
// newAccount creates a new account for testing, and funds it with any input tokens.
func (k *testKeeper) initCommunityFund(funds ...sdk.Coin) sdk.AccAddress {
	// create and fund account
	addr := k.newAccount(funds...)

	// change only the community fund address in params
	params := k.GetParams(k.ctx)
	params.CommunityFundAddress = addr.String()

	// set params and expect no error
	validMsg := &incentive.MsgGovSetParams{
		Authority: "govAcct",
		// Authority:   app.GovKeeper.GetGovernanceAccount(k.ctx).GetAddress().String(),
		Title:       "Update Params",
		Description: "New valid values",
		Params:      params,
	}
	_, err := k.msrv.GovSetParams(k.ctx, validMsg)
	require.Nil(k.t, err, "init community fund")
	return addr
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

	// add program and expect no error
	validMsg := &incentive.MsgGovCreatePrograms{
		Authority:         govAccAddr,
		Title:             "Add valid program",
		Description:       "using addIncentiveProgram helper function",
		Programs:          []incentive.IncentiveProgram{program},
		FromCommunityFund: fromCommunity,
	}
	// pass and optionally fund the program from community fund
	_, err := k.msrv.GovCreatePrograms(k.ctx, validMsg)
	require.Nil(k.t, err, "addIncentiveProgram")
}

// sponsor a program with tokens and require no errors. Use when setting up incentive scenarios.
func (k *testKeeper) sponsor(addr sdk.AccAddress, programID uint32) {
	program, _, err := k.getIncentiveProgram(k.ctx, programID)
	require.NoError(k.t, err, "get program")

	msg := &incentive.MsgSponsor{
		Sponsor: addr.String(),
		Program: programID,
		Asset:   program.TotalRewards,
	}
	_, err = k.msrv.Sponsor(k.ctx, msg)
	require.NoError(k.t, err, "sponsor program", programID)
}

// advanceTime runs the functions normally contained in EndBlocker with a fixed time elapsed.
// requires nonzero lastRewardsTime and a positive duration. Requires no error.
func (k *testKeeper) advanceTime(duration int64) {
	if duration <= 0 {
		panic("advanceTime needs positive duration")
	}

	prevTime := k.GetLastRewardsTime(k.ctx)
	if prevTime <= 0 {
		panic("last rewards time not initialized")
	}

	// simulate new block time to target an exact time elapsed
	blockTime := prevTime + duration
	require.Nil(k.t, k.updateRewards(k.ctx, prevTime, blockTime), "update rewards")
	require.Nil(k.t, k.updatePrograms(k.ctx, blockTime), "update programs")
}

// advanceTimeTo runs the functions normally contained in EndBlocker from the current rewards time
// to a target time. Requires positive duration and an initialized lastRewardsTime.
func (k *testKeeper) advanceTimeTo(end int64) {
	prevTime := k.GetLastRewardsTime(k.ctx)
	k.advanceTime(end - prevTime)
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
