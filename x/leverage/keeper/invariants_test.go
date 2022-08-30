package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	umeeapp "github.com/umee-network/umee/v3/app"
	"github.com/umee-network/umee/v3/x/leverage/keeper"
	"github.com/umee-network/umee/v3/x/leverage/types"
)

func (s *IntegrationTestSuite) TestReserveAmountInvariant() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// artificially set reserves
	err := s.tk.SetReserveAmount(ctx, sdk.NewInt64Coin(umeeapp.BondDenom, 300000000)) // 300 umee
	require.NoError(err)

	// check invariants
	_, broken := keeper.ReserveAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)
}

func (s *IntegrationTestSuite) TestCollateralAmountInvariant() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(coin(umeeDenom, 1000_000000))
	s.supply(addr, coin(umeeDenom, 1000_000000))
	s.collateralize(addr, coin("u/"+umeeDenom, 1000_000000))

	// check invariant
	_, broken := keeper.CollateralAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)

	uTokenDenom := types.ToUTokenDenom(umeeapp.BondDenom)

	// withdraw the supplied umee in the initBorrowScenario
	_, err := app.LeverageKeeper.Withdraw(ctx, addr, sdk.NewInt64Coin(uTokenDenom, 1000000000))
	require.NoError(err)

	// check invariant
	_, broken = keeper.CollateralAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)
}

func (s *IntegrationTestSuite) TestBorrowAmountInvariant() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(coin(umeeDenom, 1000_000000))
	s.supply(addr, coin(umeeDenom, 1000_000000))
	s.collateralize(addr, coin("u/"+umeeDenom, 1000_000000))

	// user borrows 20 umee
	err := app.LeverageKeeper.Borrow(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000))
	require.NoError(err)

	// check invariant
	_, broken := keeper.BorrowAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)

	// user repays 30 umee, actually only 20 because is the min between
	// the amount borrowed and the amount repaid
	_, err = app.LeverageKeeper.Repay(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 30000000))
	require.NoError(err)

	// check invariant
	_, broken = keeper.BorrowAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)
}
