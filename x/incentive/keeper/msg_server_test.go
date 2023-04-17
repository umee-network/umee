package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/incentive"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

func (s *IntegrationTestSuite) TestMsgBond() {
	ctx, srv, require := s.ctx, s.msgSrvr, s.Require()
	lk := s.mockLeverage

	// create an account which the mock leverage keeper will report as
	// having 50u/uumee collateral. No tokens ot uTokens are actually minted.
	umeeSupplier := s.newAccount()
	lk.setCollateral(umeeSupplier, "u/"+umeeDenom, 50)

	// create an additional account which has supplied an unregistered denom
	// which nonetheless has a uToken prefix. The incentive module will allow
	// this to be bonded (though u/unknown wounldn't be eligible for rewards unless
	// a program was passed to incentivize it)
	unknownSupplier := s.newAccount()
	lk.setCollateral(unknownSupplier, "u/unknown", 50)

	// create an account which has somehow managed to collateralize tokens (not uTokens).
	// bonding these should not be possible
	errorSupplier := s.newAccount()
	lk.setCollateral(errorSupplier, "uumee", 50)

	// empty address
	msg := &incentive.MsgBond{
		Account: "",
		UToken:  coin.New("u/"+umeeDenom, 10),
	}
	_, err := srv.Bond(ctx, msg)
	require.ErrorContains(err, "empty address", "empty address")

	// attempt to bond 10 u/uumee out of 50 available
	msg = &incentive.MsgBond{
		Account: umeeSupplier.String(),
		UToken:  coin.New("u/"+umeeDenom, 10),
	}
	_, err = srv.Bond(ctx, msg)
	require.Nil(err, "bond 10")

	// attempt to bond 40 u/uumee out of the remaining 40 available
	msg = &incentive.MsgBond{
		Account: umeeSupplier.String(),
		UToken:  coin.New("u/"+umeeDenom, 40),
	}
	_, err = srv.Bond(ctx, msg)
	require.Nil(err, "bond 40")

	// attempt to bond 10 u/uumee, but all 50 is already bonded
	msg = &incentive.MsgBond{
		Account: umeeSupplier.String(),
		UToken:  coin.New("u/"+umeeDenom, 40),
	}
	_, err = srv.Bond(ctx, msg)
	require.ErrorIs(err, incentive.ErrInsufficientCollateral, "bond 10 #2")

	// attempt to bond 10 u/unknown, which should work
	msg = &incentive.MsgBond{
		Account: unknownSupplier.String(),
		UToken:  coin.New("u/unknown", 10),
	}
	_, err = srv.Bond(ctx, msg)
	require.Nil(err, "bond 10 unknown uToken")

	// attempt to bond 10 u/uumee, from an account which has zero
	msg = &incentive.MsgBond{
		Account: unknownSupplier.String(),
		UToken:  coin.New("u/"+umeeDenom, 10),
	}
	_, err = srv.Bond(ctx, msg)
	require.ErrorIs(err, incentive.ErrInsufficientCollateral, "bond 10 #3")

	// attempt to bond 10 uumee, which should fail
	msg = &incentive.MsgBond{
		Account: errorSupplier.String(),
		UToken:  coin.New(umeeDenom, 10),
	}
	_, err = srv.Bond(ctx, msg)
	require.ErrorIs(err, leveragetypes.ErrNotUToken, "bond non-uToken")
}

func (s *IntegrationTestSuite) TestMsgBeginUnbonding() {
	ctx, srv, require := s.ctx, s.msgSrvr, s.Require()
	lk := s.mockLeverage

	// create an account which the mock leverage keeper will report as
	// having 50u/uumee collateral. No tokens ot uTokens are actually minted.
	// bond those uTokens.
	umeeSupplier := s.newAccount()
	lk.setCollateral(umeeSupplier, "u/"+umeeDenom, 50)
	s.bond(umeeSupplier, coin.New("u/"+umeeDenom, 50))

	// create an additional account which has supplied an unregistered denom
	// which nonetheless has a uToken prefix. Bond those utokens.
	unknownSupplier := s.newAccount()
	lk.setCollateral(unknownSupplier, "u/unknown", 50)
	s.bond(unknownSupplier, coin.New("u/unknown", 50))

	// empty address
	msg := &incentive.MsgBeginUnbonding{
		Account: "",
		UToken:  coin.New("u/"+umeeDenom, 10),
	}
	_, err := srv.BeginUnbonding(ctx, msg)
	require.ErrorContains(err, "empty address", "empty address")

	// base token
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		UToken:  coin.New(umeeDenom, 10),
	}
	_, err = srv.BeginUnbonding(ctx, msg)
	require.ErrorIs(err, leveragetypes.ErrNotUToken)

	// attempt to begin unbonding 10 u/uumee out of 50 available
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		UToken:  coin.New("u/"+umeeDenom, 10),
	}
	_, err = srv.BeginUnbonding(ctx, msg)
	require.Nil(err, "begin unbonding 10")

	// attempt to begin unbonding 50 u/uumee more (only 40 available)
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		UToken:  coin.New("u/"+umeeDenom, 50),
	}
	_, err = srv.BeginUnbonding(ctx, msg)
	require.ErrorIs(err, incentive.ErrInsufficientBonded, "begin unbonding 50")
	// TODO: why was this passing at amount 40?

	// attempt to begin unbonding 50 u/unknown but from the wrong account
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		UToken:  coin.New("u/unknown", 50),
	}
	_, err = srv.BeginUnbonding(ctx, msg)
	require.ErrorIs(err, incentive.ErrInsufficientBonded, "begin unbonding 50 unknown (wrong account)")

	// attempt to begin unbonding 50 u/unknown but from the correct account
	msg = &incentive.MsgBeginUnbonding{
		Account: unknownSupplier.String(),
		UToken:  coin.New("u/unknown", 50),
	}
	_, err = srv.BeginUnbonding(ctx, msg)
	require.Nil(err, "begin unbonding 50 unknown")

	// attempt a large number of unbondings to hit MaxUnbondings
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		UToken:  coin.New("u/"+umeeDenom, 1),
	}
	// create 4 more unbondings of u/uumee on this account, to hit the default maximum of 5
	for i := 1; i < 5; i++ {
		_, err = srv.BeginUnbonding(ctx, msg)
		require.Nil(err, "repeat begin unbonding 1")
	}
	// exceed max unbondings
	_, err = srv.BeginUnbonding(ctx, msg)
	require.ErrorIs(err, incentive.ErrMaxUnbondings, "max unbondings")

	// forcefully advance time, but not enough to finish any unbondings
	s.advanceTime(1)
	_, err = srv.BeginUnbonding(ctx, msg)
	require.ErrorIs(err, incentive.ErrMaxUnbondings, "max unbondings")

	// forcefully advance time, enough to finish all unbondings
	s.advanceTime(s.k.GetParams(s.ctx).UnbondingDuration)
	_, err = srv.BeginUnbonding(ctx, msg)
	require.Nil(err, "unbonding available after max unbondings finish")
}

func (s *IntegrationTestSuite) TestMsgEmergencyUnbond() {
	ctx, srv, require := s.ctx, s.msgSrvr, s.Require()
	lk := s.mockLeverage

	// create an account which the mock leverage keeper will report as
	// having 50u/uumee collateral. No tokens ot uTokens are actually minted.
	// bond those uTokens.
	umeeSupplier := s.newAccount()
	lk.setCollateral(umeeSupplier, "u/"+umeeDenom, 50)
	s.bond(umeeSupplier, coin.New("u/"+umeeDenom, 50))

	// create an additional account which has supplied an unregistered denom
	// which nonetheless has a uToken prefix. Bond those utokens.
	unknownSupplier := s.newAccount()
	lk.setCollateral(unknownSupplier, "u/unknown", 50)
	s.bond(unknownSupplier, coin.New("u/unknown", 50))

	// empty address
	msg := &incentive.MsgEmergencyUnbond{
		Account: "",
		UToken:  coin.New("u/"+umeeDenom, 10),
	}
	_, err := srv.EmergencyUnbond(ctx, msg)
	require.ErrorContains(err, "empty address", "empty address")

	// base token
	msg = &incentive.MsgEmergencyUnbond{
		Account: umeeSupplier.String(),
		UToken:  coin.New(umeeDenom, 10),
	}
	_, err = srv.EmergencyUnbond(ctx, msg)
	require.ErrorIs(err, leveragetypes.ErrNotUToken)

	// attempt to emergency unbond 10 u/uumee out of 50 available
	msg = &incentive.MsgEmergencyUnbond{
		Account: umeeSupplier.String(),
		UToken:  coin.New("u/"+umeeDenom, 10),
	}
	_, err = srv.EmergencyUnbond(ctx, msg)
	require.Nil(err, "emergency unbond 10")

	// attempt to emergency unbond 50 u/uumee more (only 40 available)
	msg = &incentive.MsgEmergencyUnbond{
		Account: umeeSupplier.String(),
		UToken:  coin.New("u/"+umeeDenom, 50),
	}
	_, err = srv.EmergencyUnbond(ctx, msg)
	require.ErrorIs(err, incentive.ErrInsufficientBonded, "emergency unbond 50")

	// attempt to emergency unbond 50 u/unknown but from the wrong account
	msg = &incentive.MsgEmergencyUnbond{
		Account: umeeSupplier.String(),
		UToken:  coin.New("u/unknown", 50),
	}
	_, err = srv.EmergencyUnbond(ctx, msg)
	require.ErrorIs(err, incentive.ErrInsufficientBonded, "emergency unbond 50 unknown (wrong account)")

	// attempt to emergency unbond 50 u/unknown but from the correct account
	msg = &incentive.MsgEmergencyUnbond{
		Account: unknownSupplier.String(),
		UToken:  coin.New("u/unknown", 50),
	}
	_, err = srv.EmergencyUnbond(ctx, msg)
	require.Nil(err, "emergency unbond 50 unknown")

	// attempt a large number of emergency unbondings which would hit MaxUnbondings if they were not instant
	msg = &incentive.MsgEmergencyUnbond{
		Account: umeeSupplier.String(),
		UToken:  coin.New("u/"+umeeDenom, 1),
	}
	// 4 more emergency unbondings of u/uumee on this account, which would reach the default maximum of 5 if not instant
	for i := 1; i < 5; i++ {
		_, err = srv.EmergencyUnbond(ctx, msg)
		require.Nil(err, "repeat emergency unbond 1")
	}
	// this would exceed max unbondings, but because the unbondings are instant, it does not
	_, err = srv.EmergencyUnbond(ctx, msg)
	require.Nil(err, "emergency unbond does is not restricted by max unbondings")

	// TODO: confirm donated collateral amounts using mock leverage keeper
}

func (s *IntegrationTestSuite) TestMsgSponsor() {
	ctx, srv, require := s.ctx, s.msgSrvr, s.Require()

	sponsor := s.newAccount(sdk.NewInt64Coin(umeeDenom, 15_000000))

	govAccAddr := s.app.GovKeeper.GetGovernanceAccount(s.ctx).GetAddress().String()

	validProgram := incentive.IncentiveProgram{
		ID:               0,
		StartTime:        100,
		Duration:         100,
		UToken:           "u/" + umeeDenom,
		Funded:           false,
		TotalRewards:     sdk.NewInt64Coin(umeeDenom, 10_000000),
		RemainingRewards: coin.Zero(umeeDenom),
	}

	// require that NextProgramID starts at the correct value
	require.Equal(uint32(1), s.k.GetNextProgramID(ctx), "initial next ID")

	// add program and expect no error
	validMsg := &incentive.MsgGovCreatePrograms{
		Authority:         govAccAddr,
		Title:             "Add two valid program",
		Description:       "Both will require manual funding",
		Programs:          []incentive.IncentiveProgram{validProgram, validProgram},
		FromCommunityFund: true,
	}
	// pass but do not fund the programs
	_, err := srv.GovCreatePrograms(ctx, validMsg)
	require.Nil(err, "set valid programs")
	require.Equal(uint32(3), s.k.GetNextProgramID(ctx), "next Id after 2 programs passed")

	wrongAssetSponsorMsg := &incentive.MsgSponsor{
		Sponsor: sponsor.String(),
		Program: 1,
		Asset:   sdk.NewInt64Coin(umeeDenom, 5_000000),
	}
	wrongProgramSponsorMsg := &incentive.MsgSponsor{
		Sponsor: sponsor.String(),
		Program: 3,
		Asset:   sdk.NewInt64Coin(umeeDenom, 10_000000),
	}
	validSponsorMsg := &incentive.MsgSponsor{
		Sponsor: sponsor.String(),
		Program: 1,
		Asset:   sdk.NewInt64Coin(umeeDenom, 10_000000),
	}
	failSponsorMsg := &incentive.MsgSponsor{
		Sponsor: sponsor.String(),
		Program: 2,
		Asset:   sdk.NewInt64Coin(umeeDenom, 10_000000),
	}

	// test cases
	_, err = srv.Sponsor(ctx, wrongAssetSponsorMsg)
	require.ErrorIs(err, incentive.ErrSponsorInvalid, "sponsor without exact asset match")
	_, err = srv.Sponsor(ctx, wrongProgramSponsorMsg)
	require.ErrorContains(err, "not found", "sponsor non-existing program")
	_, err = srv.Sponsor(ctx, validSponsorMsg)
	require.Nil(err, "valid sponsor")
	_, err = srv.Sponsor(ctx, validSponsorMsg)
	require.ErrorIs(err, incentive.ErrSponsorIneligible, "already funded program")
	_, err = srv.Sponsor(ctx, failSponsorMsg)
	require.ErrorContains(err, "insufficient sponsor tokens", "sponsor with insufficient funds")
}

func (s *IntegrationTestSuite) TestMsgGovSetParams() {
	ctx, srv, require := s.ctx, s.msgSrvr, s.Require()

	govAccAddr := s.app.GovKeeper.GetGovernanceAccount(s.ctx).GetAddress().String()

	// ensure that module is starting with default params
	defaultParams := incentive.DefaultParams()
	require.Equal(defaultParams, s.k.GetParams(ctx))

	// create an account to use as community fund
	communityFund := s.newAccount()

	// create new set of params which is different (in every field) from default
	newParams := incentive.Params{
		MaxUnbondings:        defaultParams.MaxUnbondings + 1,
		UnbondingDuration:    defaultParams.UnbondingDuration + 1,
		EmergencyUnbondFee:   sdk.MustNewDecFromStr("0.99"),
		CommunityFundAddress: communityFund.String(),
	}

	// set params and expect no error
	validMsg := &incentive.MsgGovSetParams{
		Authority:   govAccAddr,
		Title:       "Update Params",
		Description: "New valid values",
		Params:      newParams,
	}
	_, err := srv.GovSetParams(ctx, validMsg)
	require.Nil(err, "set valid params")

	// ensure params have changed
	require.Equal(newParams, s.k.GetParams(ctx))

	// create an invalid message
	invalidMsg := &incentive.MsgGovSetParams{
		Authority:   "",
		Title:       "",
		Description: "",
		Params:      incentive.Params{},
	}
	_, err = srv.GovSetParams(ctx, invalidMsg)
	// error comes from params validate
	require.ErrorContains(err, "max unbondings cannot be zero")
	// ensure params have not changed
	require.Equal(newParams, s.k.GetParams(ctx))
}

func (s *IntegrationTestSuite) TestMsgGovCreatePrograms() {
	ctx, srv, require := s.ctx, s.msgSrvr, s.Require()

	// create an account to use as community fund, with 15 UMEE
	_ = s.initCommunityFund(
		sdk.NewInt64Coin(umeeDenom, 15_000000),
	)

	govAccAddr := s.app.GovKeeper.GetGovernanceAccount(s.ctx).GetAddress().String()

	validProgram := incentive.IncentiveProgram{
		ID:               0,
		StartTime:        100,
		Duration:         100,
		UToken:           "u/" + umeeDenom,
		Funded:           false,
		TotalRewards:     sdk.NewInt64Coin(umeeDenom, 10_000000),
		RemainingRewards: coin.Zero(umeeDenom),
	}

	// require that NextProgramID starts at the correct value
	require.Equal(uint32(1), s.k.GetNextProgramID(ctx), "initial next ID")

	// add program and expect no error
	validMsg := &incentive.MsgGovCreatePrograms{
		Authority:         govAccAddr,
		Title:             "Add valid program",
		Description:       "Awards 10 UMEE to u/UMEE suppliers over 100 blocks",
		Programs:          []incentive.IncentiveProgram{validProgram},
		FromCommunityFund: true,
	}
	// pass and fund the program using 10 UMEE from community fund
	_, err := srv.GovCreatePrograms(ctx, validMsg)
	require.Nil(err, "set valid program")
	require.Equal(uint32(2), s.k.GetNextProgramID(ctx), "next Id after 1 program passed")

	// pass and then attempt to fund the program again using 10 UMEE from community fund, but only 5 remains
	_, err = srv.GovCreatePrograms(ctx, validMsg)
	require.Nil(err, "insufficient funds, but still passes and reverts to manual funding")
	require.Equal(uint32(3), s.k.GetNextProgramID(ctx), "next Id after 2 programs passed")

	invalidProgram := validProgram
	invalidProgram.ID = 1
	invalidMsg := &incentive.MsgGovCreatePrograms{
		Authority:         govAccAddr,
		Title:             "Add invalid program",
		Description:       "",
		Programs:          []incentive.IncentiveProgram{invalidProgram},
		FromCommunityFund: true,
	}
	// program should fail to be added, and nextID is unchanged
	_, err = srv.GovCreatePrograms(ctx, invalidMsg)
	require.ErrorIs(err, incentive.ErrInvalidProgramID, "set invalid program")
	require.Equal(uint32(3), s.k.GetNextProgramID(ctx), "next ID after 2 programs passed an 1 failed")

	// TODO: messages with multiple programs, including partially invalid
	// and checking exact equality with upcoming programs set
}
