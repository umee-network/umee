package keeper_test

import (
	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"
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
	require.ErrorIs(err, oracletypes.ErrUnknownDenom, "withdraw collateral umee")

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
	require.ErrorIs(err, oracletypes.ErrUnknownDenom, "decollateralize collateral umee")

	// UMEE cannot be borrowed since ATOM borrowed value is unknown
	msg8 := &types.MsgBorrow{
		Borrower: umeeSupplier.String(),
		Asset:    coin.New(umeeDenom, 1),
	}
	_, err = srv.Borrow(ctx, msg8)
	require.ErrorIs(err, oracletypes.ErrUnknownDenom, "borrow umee")

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

	s.mockOracle.Reset()
}

// TestCollateralPriceOutage tests price outage scenarios where a collateral token
// has unknown price
func (s *IntegrationTestSuite) TestCollateralPriceOutage() {
	ctx, srv, require := s.ctx, s.msgSrvr, s.Require()

	// create an ATOM supplier
	atomSupplier := s.newAccount(coin.New(atomDenom, 100_000000))
	s.supply(atomSupplier, coin.New(atomDenom, 100_000000))

	// create a supplier to supply 150 UMEE, and collateralize 100 UMEE
	umeeSupplier := s.newAccount(coin.New(umeeDenom, 200_000000))
	s.supply(umeeSupplier, coin.New(umeeDenom, 150_000000))
	s.collateralize(umeeSupplier, coin.New("u/"+umeeDenom, 100_000000))
	// additionally borrow 0.000001 ATOM
	s.borrow(umeeSupplier, coin.New(atomDenom, 1))

	// Create an UMEE price outage
	s.mockOracle.Clear("UMEE")

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

	// Collateral UMEE cannot be withdrawn since collateral UMEE value is unknown
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

	// Collateral UMEE cannot be decollateralized since collateral UMEE value is unknown
	msg7 := &types.MsgDecollateralize{
		Borrower: umeeSupplier.String(),
		Asset:    coin.New("u/"+umeeDenom, 1),
	}
	_, err = srv.Decollateralize(ctx, msg7)
	require.ErrorIs(err, types.ErrUndercollaterized, "decollateralize collateral umee")

	// UMEE cannot be borrowed since UMEE value is unknown
	msg8 := &types.MsgBorrow{
		Borrower: umeeSupplier.String(),
		Asset:    coin.New(umeeDenom, 1),
	}
	_, err = srv.Borrow(ctx, msg8)
	require.ErrorIs(err, oracletypes.ErrUnknownDenom, "borrow umee")

	// UMEE max-borrow succeeds with amount = zero since UMEE cannot be borrowed
	msg9 := &types.MsgMaxBorrow{
		Borrower: umeeSupplier.String(),
		Denom:    umeeDenom,
	}
	resp9, err := srv.MaxBorrow(ctx, msg9)
	require.NoError(err, "max-borrow umee")
	require.Equal(int64(0), resp9.Borrowed.Amount.Int64(), "max borrow umee")

	// Liquidation fails because collateral value cannot be calculated
	msg11 := &types.MsgLiquidate{
		Liquidator:  umeeSupplier.String(),
		Borrower:    umeeSupplier.String(),
		Repayment:   coin.New(umeeDenom, 1),
		RewardDenom: umeeDenom,
	}
	_, err = srv.Liquidate(ctx, msg11)
	require.ErrorIs(err, oracletypes.ErrUnknownDenom, "liquidate umee")

	// ATOM repay succeeds
	msg12 := &types.MsgRepay{
		Borrower: umeeSupplier.String(),
		Asset:    coin.New(atomDenom, 1),
	}
	_, err = srv.Repay(ctx, msg12)
	require.NoError(err, "repay atom")

	s.mockOracle.Reset()
}

// TestPartialCollateralPriceOutage tests price outage scenarios where two collateral
// tokens are used and only one collateral token has unknown price
func (s *IntegrationTestSuite) TestCollateralPartialPriceOutage() {
	app, ctx, srv, require := s.app, s.ctx, s.msgSrvr, s.Require()

	// create an ATOM supplier
	atomSupplier := s.newAccount(coin.New(atomDenom, 100_000000))
	s.supply(atomSupplier, coin.New(atomDenom, 100_000000))

	// create a supplier to supply 150 UMEE, and collateralize 100 UMEE
	// plus the same amounts of ATOM
	bothSupplier := s.newAccount(coin.New(umeeDenom, 200_000000), coin.New(atomDenom, 200_000000))
	s.supply(bothSupplier, coin.New(umeeDenom, 150_000000))
	s.collateralize(bothSupplier, coin.New("u/"+umeeDenom, 100_000000))
	s.supply(bothSupplier, coin.New(atomDenom, 150_000000))
	s.collateralize(bothSupplier, coin.New("u/"+atomDenom, 100_000000))
	// additionally borrow 0.000001 ATOM
	s.borrow(bothSupplier, coin.New(atomDenom, 1))

	// Create an UMEE price outage
	s.mockOracle.Clear("UMEE")

	// UMEE can still be supplied
	msg1 := &types.MsgSupply{
		Supplier: bothSupplier.String(),
		Asset:    coin.New(umeeDenom, 50_000000),
	}
	_, err := srv.Supply(ctx, msg1)
	require.NoError(err, "supply umee")

	// ATOM can still be supplied
	msg2 := &types.MsgSupply{
		Supplier: bothSupplier.String(),
		Asset:    coin.New(atomDenom, 50_000000),
	}
	_, err = srv.Supply(ctx, msg2)
	require.NoError(err, "supply atom")

	// Non-collateral UMEE can still be withdrawn
	msg3 := &types.MsgWithdraw{
		Supplier: bothSupplier.String(),
		Asset:    coin.New("u/"+umeeDenom, 50_000000),
	}
	_, err = srv.Withdraw(ctx, msg3)
	require.NoError(err, "withdraw non-collateral umee")

	// Non-collateral ATOM can still be withdrawn
	msg4 := &types.MsgWithdraw{
		Supplier: bothSupplier.String(),
		Asset:    coin.New("u/"+atomDenom, 50_000000),
	}
	_, err = srv.Withdraw(ctx, msg4)
	require.NoError(err, "withdraw non-collateral atom")

	// Non-collateral and Collateral UMEE can still be withdrawn using max withdraw
	msg5 := &types.MsgMaxWithdraw{
		Supplier: bothSupplier.String(),
		Denom:    umeeDenom,
	}
	_, err = srv.MaxWithdraw(ctx, msg5)
	require.NoError(err, "max withdraw umee")
	supplied, err := app.LeverageKeeper.GetSupplied(ctx, bothSupplier, umeeDenom)
	require.NoError(err, "measure umee supplied")
	require.Equal(int64(0), supplied.Amount.Int64(), "all umee withdrawn by maxwithdraw")

	// Non-collateral ATOM and some collateral ATOM can still be withdrawn using max withdraw
	msg6 := &types.MsgMaxWithdraw{
		Supplier: bothSupplier.String(),
		Denom:    atomDenom,
	}
	_, err = srv.MaxWithdraw(ctx, msg6)
	require.NoError(err, "max withdraw non-collateral atom")
	supplied, err = app.LeverageKeeper.GetSupplied(ctx, bothSupplier, atomDenom)
	require.NoError(err, "measure atom supplied")
	require.Equal(int64(4), supplied.Amount.Int64(), "some atom collateral withdrawn by maxwithdraw")

	// UMEE can still be collateralized
	s.supply(bothSupplier, coin.New(umeeDenom, 50_000000))
	msg7 := &types.MsgCollateralize{
		Borrower: bothSupplier.String(),
		Asset:    coin.New("u/"+umeeDenom, 50_000000),
	}
	_, err = srv.Collateralize(ctx, msg7)
	require.NoError(err, "collateralize umee")

	// ATOM can still be collateralized
	s.supply(bothSupplier, coin.New(atomDenom, 50_000000))
	msg8 := &types.MsgCollateralize{
		Borrower: bothSupplier.String(),
		Asset:    coin.New("u/"+atomDenom, 50_000000),
	}
	_, err = srv.Collateralize(ctx, msg8)
	require.NoError(err, "collateralize atom")

	// SupplyCollateral still works for UMEE
	msg9 := &types.MsgSupplyCollateral{
		Supplier: bothSupplier.String(),
		Asset:    coin.New(umeeDenom, 50_000000),
	}
	_, err = srv.SupplyCollateral(ctx, msg9)
	require.NoError(err, "supply+collateralize umee")

	// SupplyCollateral still works for ATOM
	msg10 := &types.MsgSupplyCollateral{
		Supplier: bothSupplier.String(),
		Asset:    coin.New(umeeDenom, 50_000000),
	}
	_, err = srv.SupplyCollateral(ctx, msg10)
	require.NoError(err, "supply+collateralize atom")

	// Collateral UMEE can still decollateralized even though collateral UMEE value is unknown
	// because ATOM collateral is sufficient to cover borrows
	msg11 := &types.MsgDecollateralize{
		Borrower: bothSupplier.String(),
		Asset:    coin.New("u/"+umeeDenom, 1),
	}
	_, err = srv.Decollateralize(ctx, msg11)
	require.NoError(err, "decollateralize collateral umee")

	// Collateral ATOM can still decollateralized even though collateral UMEE value is unknown
	// because remaining ATOM collateral is sufficient to cover borrows
	msg12 := &types.MsgDecollateralize{
		Borrower: bothSupplier.String(),
		Asset:    coin.New("u/"+atomDenom, 1),
	}
	_, err = srv.Decollateralize(ctx, msg12)
	require.NoError(err, "decollateralize collateral atom")

	// ATOM can be borrowed even though UMEE collateral value is unknown
	// because because ATOM collateral is sufficient to cover borrows
	msg13 := &types.MsgBorrow{
		Borrower: bothSupplier.String(),
		Asset:    coin.New(atomDenom, 1),
	}
	_, err = srv.Borrow(ctx, msg13)
	require.NoError(err, "borrow atom")

	// UMEE max-borrow succeeds with amount = zero since UMEE cannot be borrowed
	msg14 := &types.MsgMaxBorrow{
		Borrower: bothSupplier.String(),
		Denom:    umeeDenom,
	}
	resp14, err := srv.MaxBorrow(ctx, msg14)
	require.NoError(err, "max-borrow umee")
	require.Equal(int64(0), resp14.Borrowed.Amount.Int64(), "max borrow umee")

	// ATOM max-borrow succeeds with NONZERO amount since ATOM can still borrowed
	msg15 := &types.MsgMaxBorrow{
		Borrower: bothSupplier.String(),
		Denom:    atomDenom,
	}
	resp15, err := srv.MaxBorrow(ctx, msg15)
	require.NoError(err, "max-borrow atom")
	require.Greater(resp15.Borrowed.Amount.Int64(), int64(0), "max borrow atom")

	// Liquidation fails because UMEE collateral value cannot be calculated
	msg16 := &types.MsgLiquidate{
		Liquidator:  bothSupplier.String(),
		Borrower:    bothSupplier.String(),
		Repayment:   coin.New(atomDenom, 1),
		RewardDenom: atomDenom,
	}
	_, err = srv.Liquidate(ctx, msg16)
	require.ErrorIs(err, oracletypes.ErrUnknownDenom, "liquidate atom")

	// ATOM repay succeeds
	msg17 := &types.MsgRepay{
		Borrower: bothSupplier.String(),
		Asset:    coin.New(atomDenom, 1),
	}
	_, err = srv.Repay(ctx, msg17)
	require.NoError(err, "repay atom")

	s.mockOracle.Reset()
}
