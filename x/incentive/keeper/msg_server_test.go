package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/incentive"
	"github.com/umee-network/umee/v5/x/leverage/fixtures"
	leveragetypes "github.com/umee-network/umee/v5/x/leverage/types"
)

func TestMsgBond(t *testing.T) {
	k := newTestKeeper(t)

	const (
		umee  = fixtures.UmeeDenom
		atom  = fixtures.AtomDenom
		uumee = leveragetypes.UTokenPrefix + fixtures.UmeeDenom
		uatom = leveragetypes.UTokenPrefix + fixtures.AtomDenom
	)

	// create an account which the mock leverage keeper will report as
	// having 50u/uumee collateral. No tokens ot uTokens are actually minted.
	umeeSupplier := k.newAccount()
	k.lk.setCollateral(umeeSupplier, uumee, 50)

	// create an additional account which has supplied an unregistered denom
	// which nonetheless has a uToken prefix. The incentive module will allow
	// this to be bonded (though u/atom wounldn't be eligible for rewards unless
	// a program was passed to incentivize it)
	atomSupplier := k.newAccount()
	k.lk.setCollateral(atomSupplier, uatom, 50)

	// create an account which has somehow managed to collateralize tokens (not uTokens).
	// bonding these should not be possible
	errorSupplier := k.newAccount()
	k.lk.setCollateral(errorSupplier, umee, 50)

	// empty address
	msg := &incentive.MsgBond{
		Account: "",
		UToken:  coin.New(uumee, 10),
	}
	_, err := k.msrv.Bond(k.ctx, msg)
	require.ErrorContains(t, err, "empty address", "empty address")

	// attempt to bond 10 u/uumee out of 50 available
	msg = &incentive.MsgBond{
		Account: umeeSupplier.String(),
		UToken:  coin.New(uumee, 10),
	}
	_, err = k.msrv.Bond(k.ctx, msg)
	require.Nil(t, err, "bond 10")

	// attempt to bond 40 u/uumee out of the remaining 40 available
	msg = &incentive.MsgBond{
		Account: umeeSupplier.String(),
		UToken:  coin.New(uumee, 40),
	}
	_, err = k.msrv.Bond(k.ctx, msg)
	require.Nil(t, err, "bond 40")

	// attempt to bond 10 u/uumee, but all 50 is already bonded
	msg = &incentive.MsgBond{
		Account: umeeSupplier.String(),
		UToken:  coin.New(uumee, 40),
	}
	_, err = k.msrv.Bond(k.ctx, msg)
	require.ErrorIs(t, err, incentive.ErrInsufficientCollateral, "bond 10 #2")

	// attempt to bond 10 u/atom, which should work
	msg = &incentive.MsgBond{
		Account: atomSupplier.String(),
		UToken:  coin.New(uatom, 10),
	}
	_, err = k.msrv.Bond(k.ctx, msg)
	require.Nil(t, err, "bond 10 unregistered uToken")

	// attempt to bond 10 u/uumee, from an account which has zero
	msg = &incentive.MsgBond{
		Account: atomSupplier.String(),
		UToken:  coin.New(uumee, 10),
	}
	_, err = k.msrv.Bond(k.ctx, msg)
	require.ErrorIs(t, err, incentive.ErrInsufficientCollateral, "bond 10 #3")

	// attempt to bond 10 uumee, which should fail
	msg = &incentive.MsgBond{
		Account: errorSupplier.String(),
		UToken:  coin.New(umee, 10),
	}
	_, err = k.msrv.Bond(k.ctx, msg)
	require.ErrorIs(t, err, leveragetypes.ErrNotUToken, "bond non-uToken")
}

func TestMsgBeginUnbonding(t *testing.T) {
	k := newTestKeeper(t)

	const (
		umee  = fixtures.UmeeDenom
		atom  = fixtures.AtomDenom
		uumee = leveragetypes.UTokenPrefix + fixtures.UmeeDenom
		uatom = leveragetypes.UTokenPrefix + fixtures.AtomDenom
	)

	// create an account which the mock leverage keeper will report as
	// having 50u/uumee collateral. No tokens ot uTokens are actually minted.
	// bond those uTokens.
	umeeSupplier := k.newAccount()
	k.lk.setCollateral(umeeSupplier, uumee, 50)
	k.mustBond(umeeSupplier, coin.New(uumee, 50))

	// create an additional account which has supplied an unregistered denom
	// which nonetheless has a uToken prefix. Bond those utokens.
	atomSupplier := k.newAccount()
	k.lk.setCollateral(atomSupplier, uatom, 50)
	k.mustBond(atomSupplier, coin.New(uatom, 50))

	// empty address
	msg := &incentive.MsgBeginUnbonding{
		Account: "",
		UToken:  coin.New(uumee, 10),
	}
	_, err := k.msrv.BeginUnbonding(k.ctx, msg)
	require.ErrorContains(t, err, "empty address", "empty address")

	// base token
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		UToken:  coin.New(umee, 10),
	}
	_, err = k.msrv.BeginUnbonding(k.ctx, msg)
	require.ErrorIs(t, err, leveragetypes.ErrNotUToken)

	// attempt to begin unbonding 10 u/uumee out of 50 available
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		UToken:  coin.New(uumee, 10),
	}
	_, err = k.msrv.BeginUnbonding(k.ctx, msg)
	require.Nil(t, err, "begin unbonding 10")

	// attempt to begin unbonding 50 u/uumee more (only 40 available)
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		UToken:  coin.New(uumee, 50),
	}
	_, err = k.msrv.BeginUnbonding(k.ctx, msg)
	require.ErrorIs(t, err, incentive.ErrInsufficientBonded, "begin unbonding 50")

	// attempt to begin unbonding 50 u/atom but from the wrong account
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		UToken:  coin.New(uatom, 50),
	}
	_, err = k.msrv.BeginUnbonding(k.ctx, msg)
	require.ErrorIs(t, err, incentive.ErrInsufficientBonded, "begin unbonding 50 unknown (wrong account)")

	// attempt to begin unbonding 50 u/atom but from the correct account
	msg = &incentive.MsgBeginUnbonding{
		Account: atomSupplier.String(),
		UToken:  coin.New(uatom, 50),
	}
	_, err = k.msrv.BeginUnbonding(k.ctx, msg)
	require.Nil(t, err, "begin unbonding 50 unknown")

	// attempt a large number of unbondings to hit MaxUnbondings
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		UToken:  coin.New(uumee, 1),
	}
	// create 9 more unbondings of u/uumee on this account, to hit the default maximum of 10
	for i := 1; i < 10; i++ {
		_, err = k.msrv.BeginUnbonding(k.ctx, msg)
		require.Nil(t, err, "repeat begin unbonding 1")
	}
	// exceed max unbondings
	_, err = k.msrv.BeginUnbonding(k.ctx, msg)
	require.ErrorIs(t, err, incentive.ErrMaxUnbondings, "max unbondings")

	// forcefully advance time, but not enough to finish any unbondings
	k.advanceTime(1)
	_, err = k.msrv.BeginUnbonding(k.ctx, msg)
	require.ErrorIs(t, err, incentive.ErrMaxUnbondings, "max unbondings")

	// forcefully advance time, enough to finish all unbondings
	k.advanceTime(k.GetParams(k.ctx).UnbondingDuration)
	_, err = k.msrv.BeginUnbonding(k.ctx, msg)
	require.Nil(t, err, "unbonding available after max unbondings finish")
}

func TestMsgEmergencyUnbond(t *testing.T) {
	k := newTestKeeper(t)

	const (
		umee  = fixtures.UmeeDenom
		atom  = fixtures.AtomDenom
		uumee = leveragetypes.UTokenPrefix + fixtures.UmeeDenom
		uatom = leveragetypes.UTokenPrefix + fixtures.AtomDenom
	)

	// create an account which the mock leverage keeper will report as
	// having 50u/uumee collateral. No tokens ot uTokens are actually minted.
	// bond those uTokens.
	umeeSupplier := k.newAccount()
	k.lk.setCollateral(umeeSupplier, uumee, 50)
	k.mustBond(umeeSupplier, coin.New(uumee, 50))

	// create an additional account which has supplied an unregistered denom
	// which nonetheless has a uToken prefix. Bond those utokens.
	atomSupplier := k.newAccount()
	k.lk.setCollateral(atomSupplier, uatom, 50)
	k.mustBond(atomSupplier, coin.New(uatom, 50))

	// empty address
	msg := &incentive.MsgEmergencyUnbond{
		Account: "",
		UToken:  coin.New(uumee, 10),
	}
	_, err := k.msrv.EmergencyUnbond(k.ctx, msg)
	require.ErrorContains(t, err, "empty address", "empty address")

	// base token
	msg = &incentive.MsgEmergencyUnbond{
		Account: umeeSupplier.String(),
		UToken:  coin.New(umee, 10),
	}
	_, err = k.msrv.EmergencyUnbond(k.ctx, msg)
	require.ErrorIs(t, err, leveragetypes.ErrNotUToken)

	// attempt to emergency unbond 10 u/uumee out of 50 available
	msg = &incentive.MsgEmergencyUnbond{
		Account: umeeSupplier.String(),
		UToken:  coin.New(uumee, 10),
	}
	_, err = k.msrv.EmergencyUnbond(k.ctx, msg)
	require.Nil(t, err, "emergency unbond 10")

	// attempt to emergency unbond 50 u/uumee more (only 40 available)
	msg = &incentive.MsgEmergencyUnbond{
		Account: umeeSupplier.String(),
		UToken:  coin.New(uumee, 50),
	}
	_, err = k.msrv.EmergencyUnbond(k.ctx, msg)
	require.ErrorIs(t, err, incentive.ErrInsufficientBonded, "emergency unbond 50")

	// attempt to emergency unbond 50 u/atom but from the wrong account
	msg = &incentive.MsgEmergencyUnbond{
		Account: umeeSupplier.String(),
		UToken:  coin.New(uatom, 50),
	}
	_, err = k.msrv.EmergencyUnbond(k.ctx, msg)
	require.ErrorIs(t, err, incentive.ErrInsufficientBonded, "emergency unbond 50 unknown (wrong account)")

	// attempt to emergency unbond 50 u/atom but from the correct account
	msg = &incentive.MsgEmergencyUnbond{
		Account: atomSupplier.String(),
		UToken:  coin.New(uatom, 50),
	}
	_, err = k.msrv.EmergencyUnbond(k.ctx, msg)
	require.Nil(t, err, "emergency unbond 50 unknown")

	// attempt a large number of emergency unbondings which would hit MaxUnbondings if they were not instant
	msg = &incentive.MsgEmergencyUnbond{
		Account: umeeSupplier.String(),
		UToken:  coin.New(uumee, 1),
	}
	// 9 more emergency unbondings of u/uumee on this account, which would reach the default maximum of 10 if not instant
	for i := 1; i < 10; i++ {
		_, err = k.msrv.EmergencyUnbond(k.ctx, msg)
		require.Nil(t, err, "repeat emergency unbond 1")
	}
	// this would exceed max unbondings, but because the unbondings are instant, it does not
	_, err = k.msrv.EmergencyUnbond(k.ctx, msg)
	require.Nil(t, err, "emergency unbond does is not restricted by max unbondings")

	// TODO: confirm donated collateral amounts using mock leverage keeper
}

func TestMsgSponsor(t *testing.T) {
	k := newTestKeeper(t)

	const (
		umee  = fixtures.UmeeDenom
		uumee = leveragetypes.UTokenPrefix + fixtures.UmeeDenom
	)

	sponsor := k.newAccount(sdk.NewInt64Coin(umee, 15_000000))

	govAccAddr := "govAcct"

	validProgram := incentive.IncentiveProgram{
		ID:               0,
		StartTime:        100,
		Duration:         100,
		UToken:           uumee,
		Funded:           false,
		TotalRewards:     sdk.NewInt64Coin(umee, 10_000000),
		RemainingRewards: coin.Zero(umee),
	}

	// require that NextProgramID starts at the correct value
	require.Equal(t, uint32(1), k.getNextProgramID(k.ctx), "initial next ID")

	// add program, with manual funding
	validMsg := &incentive.MsgGovCreatePrograms{
		Authority:         govAccAddr,
		Programs:          []incentive.IncentiveProgram{validProgram, validProgram},
		FromCommunityFund: true,
	}
	// pass but do not fund the programs
	_, err := k.msrv.GovCreatePrograms(k.ctx, validMsg)
	require.Nil(t, err, "set valid programs")
	require.Equal(t, uint32(3), k.getNextProgramID(k.ctx), "next Id after 2 programs passed")

	wrongProgramSponsorMsg := &incentive.MsgSponsor{
		Sponsor: sponsor.String(),
		Program: 3,
	}
	validSponsorMsg := &incentive.MsgSponsor{
		Sponsor: sponsor.String(),
		Program: 1,
	}
	failSponsorMsg := &incentive.MsgSponsor{
		Sponsor: sponsor.String(),
		Program: 2,
	}

	// test cases
	_, err = k.msrv.Sponsor(k.ctx, wrongProgramSponsorMsg)
	require.ErrorContains(t, err, "not found", "sponsor non-existing program")
	_, err = k.msrv.Sponsor(k.ctx, validSponsorMsg)
	require.Nil(t, err, "valid sponsor")
	_, err = k.msrv.Sponsor(k.ctx, validSponsorMsg)
	require.ErrorIs(t, err, incentive.ErrSponsorIneligible, "already funded program")
	_, err = k.msrv.Sponsor(k.ctx, failSponsorMsg)
	require.ErrorContains(t, err, "insufficient sponsor tokens", "sponsor with insufficient funds")
}

func TestMsgGovSetParams(t *testing.T) {
	k := newTestKeeper(t)

	govAccAddr := "govAcct"

	defaultParams := incentive.DefaultParams()

	// create new set of params which is different (in every field) from default
	newParams := incentive.Params{
		MaxUnbondings:      defaultParams.MaxUnbondings + 1,
		UnbondingDuration:  defaultParams.UnbondingDuration + 1,
		EmergencyUnbondFee: sdk.MustNewDecFromStr("0.99"),
	}

	// set params and expect no error
	validMsg := &incentive.MsgGovSetParams{
		Authority: govAccAddr,
		Params:    newParams,
	}
	_, err := k.msrv.GovSetParams(k.ctx, validMsg)
	require.Nil(t, err, "set valid params")

	// ensure params have changed
	require.Equal(t, newParams, k.GetParams(k.ctx))

	// create an invalid message
	invalidMsg := &incentive.MsgGovSetParams{
		Authority: "",
		Params:    incentive.Params{},
	}
	_, err = k.msrv.GovSetParams(k.ctx, invalidMsg)
	// error comes from params validate
	require.ErrorContains(t, err, "invalid emergency unbonding fee")
	// ensure params have not changed
	require.Equal(t, newParams, k.GetParams(k.ctx))
}

func TestMsgGovCreatePrograms(t *testing.T) {
	k := newTestKeeper(t)

	const (
		umee  = fixtures.UmeeDenom
		uumee = leveragetypes.UTokenPrefix + fixtures.UmeeDenom
	)

	// fund community fund with 15 UMEE
	k.initCommunityFund(
		sdk.NewInt64Coin(umee, 15_000000),
	)

	govAccAddr := "govAcct"

	validProgram := incentive.IncentiveProgram{
		ID:               0,
		StartTime:        100,
		Duration:         100,
		UToken:           uumee,
		Funded:           false,
		TotalRewards:     sdk.NewInt64Coin(umee, 10_000000),
		RemainingRewards: coin.Zero(umee),
	}

	// require that NextProgramID starts at the correct value
	require.Equal(t, uint32(1), k.getNextProgramID(k.ctx), "initial next ID")

	// Awards 10 UMEE to u/UMEE suppliers over 100 blocks"
	validMsg := &incentive.MsgGovCreatePrograms{
		Authority:         govAccAddr,
		Programs:          []incentive.IncentiveProgram{validProgram},
		FromCommunityFund: true,
	}
	// pass and fund the program using 10 UMEE from community fund
	_, err := k.msrv.GovCreatePrograms(k.ctx, validMsg)
	require.Nil(t, err, "set valid program")
	require.Equal(t, uint32(2), k.getNextProgramID(k.ctx), "next Id after 1 program passed")

	// pass and then attempt to fund the program again using 10 UMEE from community fund, but only 5 remains
	_, err = k.msrv.GovCreatePrograms(k.ctx, validMsg)
	require.Nil(t, err, "insufficient funds, but still passes and reverts to manual funding")
	require.Equal(t, uint32(3), k.getNextProgramID(k.ctx), "next Id after 2 programs passed")

	invalidProgram := validProgram
	invalidProgram.ID = 1
	invalidMsg := &incentive.MsgGovCreatePrograms{
		Authority:         "",
		Programs:          []incentive.IncentiveProgram{invalidProgram},
		FromCommunityFund: true,
	}
	// program should fail to be added, and nextID is unchanged
	_, err = k.msrv.GovCreatePrograms(k.ctx, invalidMsg)
	require.ErrorIs(t, err, incentive.ErrInvalidProgramID, "set invalid program")
	require.Equal(t, uint32(3), k.getNextProgramID(k.ctx), "next ID after 2 programs passed an 1 failed")

	// TODO: messages with multiple programs, including partially invalid
	// and checking exact equality with upcoming programs set
}
