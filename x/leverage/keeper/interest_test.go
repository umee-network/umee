package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/util/coin"
)

func (s *IntegrationTestSuite) TestAccrueZeroInterest() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(coin.New(umeeDenom, 1000_000000))
	s.supply(addr, coin.New(umeeDenom, 1000_000000))
	s.collateralize(addr, coin.New("u/"+umeeDenom, 1000_000000))

	// user borrows 40 umee
	s.borrow(addr, coin.New(appparams.BondDenom, 40_000000))

	// verify user's loan amount (40 umee)
	loanBalance := app.LeverageKeeper.GetBorrow(ctx, addr, appparams.BondDenom)
	require.Equal(coin.New(appparams.BondDenom, 40_000000), loanBalance)

	// Because no time has passed since genesis (due to test setup) this will not
	// increase borrowed amount.
	err := app.LeverageKeeper.AccrueAllInterest(ctx)
	require.NoError(err)

	// verify user's loan amount (40 umee)
	loanBalance = app.LeverageKeeper.GetBorrow(ctx, addr, appparams.BondDenom)
	require.Equal(coin.New(appparams.BondDenom, 40_000000), loanBalance)

	// borrow APY at utilization = 4%
	// when kink utilization = 80%, and base/kink APY are 0.02 and 0.22
	borrowAPY := app.LeverageKeeper.DeriveBorrowAPY(ctx, appparams.BondDenom)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("0.03"), borrowAPY)

	// supply APY when borrow APY is 3%
	// and utilization is 4%, and reservefactor is 20%, and OracleRewardFactor is 1%
	// 0.03 * 0.04 * (1 - 0.21) = 0.000948
	supplyAPY := app.LeverageKeeper.DeriveSupplyAPY(ctx, appparams.BondDenom)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("0.000948"), supplyAPY)
}

func (s *IntegrationTestSuite) TestDynamicInterest() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(coin.New(umeeDenom, 1000_000000))
	s.supply(addr, coin.New(umeeDenom, 1000_000000))
	s.collateralize(addr, coin.New("u/"+umeeDenom, 1000_000000))

	// Base interest rate (0% utilization)
	rate := app.LeverageKeeper.DeriveBorrowAPY(ctx, appparams.BondDenom)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("0.02"), rate)

	// user borrows 200 umee, utilization 200/1000
	s.borrow(addr, coin.New(appparams.BondDenom, 200_000000))

	// Between base interest and kink (20% utilization)
	rate = app.LeverageKeeper.DeriveBorrowAPY(ctx, appparams.BondDenom)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("0.07"), rate)

	// user borrows 600 more umee (ignores collateral), utilization 800/1000
	s.forceBorrow(addr, coin.New(appparams.BondDenom, 600_000000))

	// Kink interest rate (80% utilization)
	rate = app.LeverageKeeper.DeriveBorrowAPY(ctx, appparams.BondDenom)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("0.22"), rate)

	// user borrows 100 more umee (ignores collateral), utilization 900/1000
	s.forceBorrow(addr, coin.New(appparams.BondDenom, 100_000000))

	// Between kink interest and max (90% utilization)
	rate = app.LeverageKeeper.DeriveBorrowAPY(ctx, appparams.BondDenom)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("0.87"), rate)

	// user borrows 100 more umee (ignores collateral), utilization 1000/1000
	s.forceBorrow(addr, coin.New(appparams.BondDenom, 100_000000))

	// Max interest rate (100% utilization)
	rate = app.LeverageKeeper.DeriveBorrowAPY(ctx, appparams.BondDenom)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("1.52"), rate)
}

func (s *IntegrationTestSuite) TestDynamicInterest_InvalidAsset() {
	app, ctx, require := s.app, s.ctx, s.Require()

	rate := app.LeverageKeeper.DeriveBorrowAPY(ctx, "uabc")
	require.Equal(sdkmath.LegacyZeroDec(), rate)
}
