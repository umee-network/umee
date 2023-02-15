package keeper_test

import (
	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

// TestBorrowedPriceOutage tests price outage scenarios where a borrowed token
// has unknown price
func (s *IntegrationTestSuite) TestBorrowedPriceOutage() {
	ctx, srv, require := s.ctx, s.msgSrvr, s.Require()

	// create an ATOM supplier
	atomSupplier := s.newAccount(coin.New(atomDenom, 100_000000))
	s.supply(atomSupplier, coin.New(atomDenom, 100_000000))

	// create a supplier to supply 150 UMEE, and collateralize 100 UMEE
	umeeSupplier := s.newAccount(coin.New(umeeDenom, 200_000000))
	s.supply(umeeSupplier, coin.New(umeeDenom, 150_000000))
	s.collateralize(umeeSupplier, coin.New("u/"+umeeDenom, 100_000000))
	// additionally borrow 0.000001 ATOM and 0.000001 UMEE with the UMEE supplier
	s.borrow(umeeSupplier, coin.New(atomDenom, 1))
	s.borrow(umeeSupplier, coin.New(umeeDenom, 1))

	// Create an ATOM price outage
	s.mockOracle.Clear("ATOM")

	// UMEE can still be supplied
	msg1 := &types.MsgSupply{
		Supplier: umeeSupplier.String(),
		Asset:    coin.New(umeeDenom, 50_000000),
	}
	_, err := srv.Supply(ctx, msg1)
	require.NoError(err, "supply umee")

	// Non-collateral UMEE can still be withdrawn
	msg2 := &types.MsgWithdraw{
		Supplier: umeeSupplier.String(),
		Asset:    coin.New("u/"+umeeDenom, 50_000000),
	}
	_, err = srv.Withdraw(ctx, msg2)
	require.NoError(err, "withdraw non-collateral umee")

	// Non-collateral UMEE can still be withdrawn using max withdraw
	msg3 := &types.MsgMaxWithdraw{
		Supplier: umeeSupplier.String(),
		Denom:    umeeDenom,
	}
	_, err = srv.MaxWithdraw(ctx, msg3)
	require.NoError(err, "max withdraw non-collateral umee")

	// Collateral UMEE cannot be withdrawn since borrowed ATOM value is unknown
	msg4 := &types.MsgWithdraw{
		Supplier: umeeSupplier.String(),
		Asset:    coin.New("u/"+umeeDenom, 1),
	}
	_, err = srv.Withdraw(ctx, msg4)
	require.ErrorIs(err, types.ErrUndercollaterized, "withdraw collateral umee")

	// UMEE can still be collateralized
	s.supply(umeeSupplier, coin.New(umeeDenom, 50_000000))
	msg5 := &types.MsgCollateralize{
		Borrower: umeeSupplier.String(),
		Asset:    coin.New("u/"+umeeDenom, 50_000000),
	}
	_, err = srv.Collateralize(ctx, msg5)
	require.NoError(err, "collateralize umee")

	// SupplyCollateral still works for UMEE
	msg6 := &types.MsgSupplyCollateral{
		Supplier: umeeSupplier.String(),
		Asset:    coin.New(umeeDenom, 50_000000),
	}
	_, err = srv.SupplyCollateral(ctx, msg6)
	require.NoError(err, "supply+collateralize umee")

	// Collateral UMEE cannot be decollateralized since borrowed ATOM value is unknown
	msg7 := &types.MsgDecollateralize{
		Borrower: umeeSupplier.String(),
		Asset:    coin.New("u/"+umeeDenom, 1),
	}
	_, err = srv.Decollateralize(ctx, msg7)
	require.ErrorIs(err, types.ErrUndercollaterized, "decollateralize collateral umee")

	// UMEE cannot be borrowed since ATOM borrowed value is unknown
	msg8 := &types.MsgBorrow{
		Borrower: umeeSupplier.String(),
		Asset:    coin.New(umeeDenom, 1),
	}
	_, err = srv.Borrow(ctx, msg8)
	require.ErrorIs(err, types.ErrUndercollaterized, "borrow umee")

	// UMEE max-borrow succeeds with amount = zero since UMEE cannot be borrowed
	msg9 := &types.MsgMaxBorrow{
		Borrower: umeeSupplier.String(),
		Denom:    umeeDenom,
	}
	resp9, err := srv.MaxBorrow(ctx, msg9)
	require.NoError(err, "max-borrow umee")
	require.Equal(int64(0), resp9.Borrowed.Amount.Int64(), "max borrow umee")

	// UMEE repay succeeds
	msg10 := &types.MsgRepay{
		Borrower: umeeSupplier.String(),
		Asset:    coin.New(umeeDenom, 1),
	}
	_, err = srv.Repay(ctx, msg10)
	require.NoError(err, "repay umee")

	// Liquidation is ineligible because known borrowed value does not exceed borrow limit
	msg11 := &types.MsgLiquidate{
		Liquidator:  umeeSupplier.String(),
		Borrower:    umeeSupplier.String(),
		Repayment:   coin.New(umeeDenom, 1),
		RewardDenom: umeeDenom,
	}
	_, err = srv.Liquidate(ctx, msg11)
	require.ErrorIs(err, types.ErrLiquidationIneligible, "liquidate umee")

	// ATOM repay succeeds
	msg12 := &types.MsgRepay{
		Borrower: umeeSupplier.String(),
		Asset:    coin.New(atomDenom, 1),
	}
	_, err = srv.Repay(ctx, msg12)
	require.NoError(err, "repay atom")
}

// TestCollateralPriceOutage tests price outage scenarios where a collateral token
// has unknown price
func (s *IntegrationTestSuite) TestCollateralPriceOutage() {
	// ctx, srv, require := s.ctx, s.msgSrvr, s.Require()

	// create a supplier
	umeeSupplier := s.newAccount(coin.New(umeeDenom, 200_000000))
	s.supply(umeeSupplier, coin.New(umeeDenom, 200_000000))

	// create an ATOM supplier and borrow 0.000001 UMEE
	atomSupplier := s.newAccount(coin.New(atomDenom, 100_000000))
	s.supply(atomSupplier, coin.New(atomDenom, 75_000000))
	s.collateralize(atomSupplier, coin.New("u/"+atomDenom, 50_000000))
	s.borrow(atomSupplier, coin.New(umeeDenom, 1))

	// Create an ATOM price outage
	s.mockOracle.Clear("ATOM")

	// TODO: Test every message type
}

// TODO: Test complex (3-asset) scenarios
