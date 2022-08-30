package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	umeeapp "github.com/umee-network/umee/v3/app"
)

func (s *IntegrationTestSuite) TestAccrueZeroInterest() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(coin(umeeDenom, 1000_000000))
	s.supply(addr, coin(umeeDenom, 1000_000000))
	s.collateralize(addr, coin("u/"+umeeDenom, 1000_000000))

	// user borrows 40 umee
	err := app.LeverageKeeper.Borrow(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 40000000))
	require.NoError(err)

	// verify user's loan amount (40 umee)
	loanBalance := app.LeverageKeeper.GetBorrow(ctx, addr, umeeapp.BondDenom)
	require.Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 40000000))

	// Because no time has passed since genesis (due to test setup) this will not
	// increase borrowed amount.
	err = app.LeverageKeeper.AccrueAllInterest(ctx)
	require.NoError(err)

	// verify user's loan amount (40 umee)
	loanBalance = app.LeverageKeeper.GetBorrow(ctx, addr, umeeapp.BondDenom)
	require.Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 40000000))

	// borrow APY at utilization = 4%
	// when kink utilization = 80%, and base/kink APY are 0.02 and 0.22
	borrowAPY := app.LeverageKeeper.DeriveBorrowAPY(ctx, umeeapp.BondDenom)
	require.Equal(sdk.MustNewDecFromStr("0.03"), borrowAPY)

	// supply APY when borrow APY is 3%
	// and utilization is 4%, and reservefactor is 20%, and OracleRewardFactor is 1%
	// 0.03 * 0.04 * (1 - 0.21) = 0.000948
	supplyAPY := app.LeverageKeeper.DeriveSupplyAPY(ctx, umeeapp.BondDenom)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("0.000948"), supplyAPY)
}

func (s *IntegrationTestSuite) TestDynamicInterest() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(coin(umeeDenom, 1000_000000))
	s.supply(addr, coin(umeeDenom, 1000_000000))
	s.collateralize(addr, coin("u/"+umeeDenom, 1000_000000))

	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("1.0")     // to allow high utilization
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("1.0") // to allow high utilization

	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, umeeToken))

	// Base interest rate (0% utilization)
	rate := app.LeverageKeeper.DeriveBorrowAPY(ctx, umeeapp.BondDenom)
	require.Equal(rate, sdk.MustNewDecFromStr("0.02"))

	// user borrows 200 umee, utilization 200/1000
	err := app.LeverageKeeper.Borrow(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	require.NoError(err)

	// Between base interest and kink (20% utilization)
	rate = app.LeverageKeeper.DeriveBorrowAPY(ctx, umeeapp.BondDenom)
	require.Equal(rate, sdk.MustNewDecFromStr("0.07"))

	// user borrows 600 more umee, utilization 800/1000
	err = app.LeverageKeeper.Borrow(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 600000000))
	require.NoError(err)

	// Kink interest rate (80% utilization)
	rate = app.LeverageKeeper.DeriveBorrowAPY(ctx, umeeapp.BondDenom)
	require.NoError(err)
	require.Equal(rate, sdk.MustNewDecFromStr("0.22"))

	// user borrows 100 more umee, utilization 900/1000
	err = app.LeverageKeeper.Borrow(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	require.NoError(err)

	// Between kink interest and max (90% utilization)
	rate = app.LeverageKeeper.DeriveBorrowAPY(ctx, umeeapp.BondDenom)
	require.NoError(err)
	require.Equal(rate, sdk.MustNewDecFromStr("0.87"))

	// user borrows 100 more umee, utilization 1000/1000
	err = app.LeverageKeeper.Borrow(ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	require.NoError(err)

	// Max interest rate (100% utilization)
	rate = app.LeverageKeeper.DeriveBorrowAPY(ctx, umeeapp.BondDenom)
	require.NoError(err)
	require.Equal(rate, sdk.MustNewDecFromStr("1.52"))
}

func (s *IntegrationTestSuite) TestDynamicInterest_InvalidAsset() {
	app, ctx, require := s.app, s.ctx, s.Require()

	rate := app.LeverageKeeper.DeriveBorrowAPY(ctx, "uabc")
	require.Equal(rate, sdk.ZeroDec())
}
