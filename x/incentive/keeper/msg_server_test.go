package keeper_test

import (
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
