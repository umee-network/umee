package keeper_test

import (
	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/x/leverage/keeper"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

func (s *IntegrationTestSuite) TestReserveAmountInvariant() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// artificially set reserves
	s.setReserves(mkCoin(appparams.BondDenom, 300_000000))

	// check invariants
	_, broken := keeper.ReserveAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)
}

func (s *IntegrationTestSuite) TestCollateralAmountInvariant() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(mkCoin(umeeDenom, 1000_000000))
	s.supply(addr, mkCoin(umeeDenom, 1000_000000))
	s.collateralize(addr, mkCoin("u/"+umeeDenom, 1000_000000))

	// check invariant
	_, broken := keeper.InefficientCollateralAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)

	uTokenDenom := types.ToUTokenDenom(appparams.BondDenom)

	// withdraw the supplied umee in the initBorrowScenario
	s.withdraw(addr, mkCoin(uTokenDenom, 1000_000000))

	// check invariant
	_, broken = keeper.InefficientCollateralAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)
}

func (s *IntegrationTestSuite) TestBorrowAmountInvariant() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(mkCoin(umeeDenom, 1000_000000))
	s.supply(addr, mkCoin(umeeDenom, 1000_000000))
	s.collateralize(addr, mkCoin("u/"+umeeDenom, 1000_000000))

	// user borrows 20 umee
	s.borrow(addr, mkCoin(appparams.BondDenom, 20_000000))

	// check invariant
	_, broken := keeper.InefficientBorrowAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)

	// user repays 30 umee, actually only 20 because is the min between
	// the amount borrowed and the amount repaid
	_, err := app.LeverageKeeper.Repay(ctx, addr, mkCoin(appparams.BondDenom, 30_000000))
	require.NoError(err)

	// check invariant
	_, broken = keeper.InefficientBorrowAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)
}
