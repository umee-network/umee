package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	umeeapp "github.com/umee-network/umee/v3/app"
)

func (s *IntegrationTestSuite) TestSetReserves() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// get initial reserves
	amount := app.LeverageKeeper.GetReserveAmount(ctx, umeeapp.BondDenom)
	require.Equal(amount, sdk.ZeroInt())

	// artifically reserve 200 umee
	err := s.tk.SetReserveAmount(ctx, coin(umeeapp.BondDenom, 200000000))
	require.NoError(err)

	// get new reserves
	amount = app.LeverageKeeper.GetReserveAmount(ctx, umeeapp.BondDenom)
	require.Equal(amount, sdk.NewInt(200000000))
}

func (s *IntegrationTestSuite) TestRepayBadDebt() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// Creating a supplier so module account has some uumee
	addr := s.newAccount(coin(umeeDenom, 200_000000))
	s.supply(addr, coin(umeeDenom, 200_000000))

	// Using an address with no assets
	addr2 := s.newAccount()

	// Create an uncollateralized debt position
	badDebt := coin(umeeDenom, 100000000) // 100 umee
	err := s.tk.SetBorrow(ctx, addr2, badDebt)
	require.NoError(err)

	// Manually mark the bad debt for repayment
	require.NoError(s.tk.SetBadDebtAddress(ctx, addr2, umeeDenom, true))

	// Manually set reserves to 60 umee
	reserve := coin(umeeDenom, 60000000)
	err = s.tk.SetReserveAmount(ctx, reserve)
	require.NoError(err)

	// Sweep all bad debts, which should repay 60 umee of the bad debt (partial repayment)
	err = app.LeverageKeeper.SweepBadDebts(ctx)
	require.NoError(err)

	// Confirm that a debt of 40 umee remains
	remainingDebt := app.LeverageKeeper.GetBorrow(ctx, addr2, umeeDenom)
	require.Equal(coin(umeeDenom, 40000000), remainingDebt)

	// Confirm that reserves are exhausted
	remainingReserve := app.LeverageKeeper.GetReserveAmount(ctx, umeeDenom)
	require.Equal(sdk.ZeroInt(), remainingReserve)

	// Manually set reserves to 70 umee
	reserve = coin(umeeDenom, 70000000)
	err = s.tk.SetReserveAmount(ctx, reserve)
	require.NoError(err)

	// Sweep all bad debts, which should fully repay the bad debt this time
	err = app.LeverageKeeper.SweepBadDebts(ctx)
	require.NoError(err)

	// Confirm that the debt is eliminated
	remainingDebt = app.LeverageKeeper.GetBorrow(ctx, addr2, umeeDenom)
	require.Equal(coin(umeeDenom, 0), remainingDebt)

	// Confirm that reserves are now at 30 umee
	remainingReserve = app.LeverageKeeper.GetReserveAmount(ctx, umeeDenom)
	require.Equal(sdk.NewInt(30000000), remainingReserve)

	// Sweep all bad debts - but there are none
	err = app.LeverageKeeper.SweepBadDebts(ctx)
	require.NoError(err)
}
