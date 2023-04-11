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
		Asset:   coin.New("u/"+umeeDenom, 10),
	}
	_, err := srv.Bond(ctx, msg)
	require.ErrorContains(err, "empty address", "empty address")

	// attempt to bond 10 u/uumee out of 50 available
	msg = &incentive.MsgBond{
		Account: umeeSupplier.String(),
		Asset:   coin.New("u/"+umeeDenom, 10),
	}
	_, err = srv.Bond(ctx, msg)
	require.Nil(err, "bond 10")

	// attempt to bond 40 u/uumee out of the remaining 40 available
	msg = &incentive.MsgBond{
		Account: umeeSupplier.String(),
		Asset:   coin.New("u/"+umeeDenom, 40),
	}
	_, err = srv.Bond(ctx, msg)
	require.Nil(err, "bond 40")

	// attempt to bond 10 u/uumee, but all 50 is already bonded
	msg = &incentive.MsgBond{
		Account: umeeSupplier.String(),
		Asset:   coin.New("u/"+umeeDenom, 40),
	}
	_, err = srv.Bond(ctx, msg)
	require.ErrorIs(err, incentive.ErrInsufficientCollateral, "bond 10 #2")

	// attempt to bond 10 u/unknown, which should work
	msg = &incentive.MsgBond{
		Account: unknownSupplier.String(),
		Asset:   coin.New("u/unknown", 10),
	}
	_, err = srv.Bond(ctx, msg)
	require.Nil(err, "bond 10 unknown uToken")

	// attempt to bond 10 u/uumee, from an account which has zero
	msg = &incentive.MsgBond{
		Account: unknownSupplier.String(),
		Asset:   coin.New("u/"+umeeDenom, 10),
	}
	_, err = srv.Bond(ctx, msg)
	require.ErrorIs(err, incentive.ErrInsufficientCollateral, "bond 10 #3")

	// attempt to bond 10 uumee, which should fail
	msg = &incentive.MsgBond{
		Account: errorSupplier.String(),
		Asset:   coin.New(umeeDenom, 10),
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
		Asset:   coin.New("u/"+umeeDenom, 10),
	}
	_, err := srv.BeginUnbonding(ctx, msg)
	require.ErrorContains(err, "empty address", "empty address")

	// base token
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		Asset:   coin.New(umeeDenom, 10),
	}
	_, err = srv.BeginUnbonding(ctx, msg)
	require.ErrorIs(err, leveragetypes.ErrNotUToken)

	// attempt to begin unbonding 10 u/uumee out of 50 available
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		Asset:   coin.New("u/"+umeeDenom, 10),
	}
	_, err = srv.BeginUnbonding(ctx, msg)
	require.Nil(err, "begin unbonding 10")

	// attempt to begin unbonding 50 u/uumee more (only 40 available)
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		Asset:   coin.New("u/"+umeeDenom, 40),
	}
	_, err = srv.BeginUnbonding(ctx, msg)
	require.ErrorIs(err, incentive.ErrInsufficientBonded, "begin unbonding 40")

	// attempt to begin unbonding 50 u/unknown but from the wrong account
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		Asset:   coin.New("u/unknown", 50),
	}
	_, err = srv.BeginUnbonding(ctx, msg)
	require.ErrorIs(err, incentive.ErrInsufficientBonded, "begin unbonding 50 unknown (wrong account)")

	// attempt to begin unbonding 50 u/unknown but from the correct account
	msg = &incentive.MsgBeginUnbonding{
		Account: unknownSupplier.String(),
		Asset:   coin.New("u/unknown", 50),
	}
	_, err = srv.BeginUnbonding(ctx, msg)
	require.Nil(err, "begin unbonding 50 unknown")

	// attempt a large number of unbondings to hit MaxUnbondings
	msg = &incentive.MsgBeginUnbonding{
		Account: umeeSupplier.String(),
		Asset:   coin.New("u/"+umeeDenom, 1),
	}
	// create 4 more unbondings of u/uumee on this account, to hit the default maximum of 5
	for i := 1; i < 5; i++ {
		_, err = srv.BeginUnbonding(ctx, msg)
		require.Nil(err, "repeat begin unbonding 1")
	}
	// exceed max unbondings
	_, err = srv.BeginUnbonding(ctx, msg)
	require.ErrorIs(err, incentive.ErrMaxUnbondings, "max unbondings")
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
