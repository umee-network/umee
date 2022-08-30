package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	umeeapp "github.com/umee-network/umee/v3/app"
)

func (s *IntegrationTestSuite) TestAccrueZeroInterest() {
	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(coin(umeeDenom, 1000_000000))
	s.supply(addr, coin(umeeDenom, 1000_000000))
	s.collateralize(addr, coin("u/"+umeeDenom, 1000_000000))

	// user borrows 40 umee
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 40000000))
	s.Require().NoError(err)

	// verify user's loan amount (40 umee)
	loanBalance := s.app.LeverageKeeper.GetBorrow(s.ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 40000000))

	// Because no time has passed since genesis (due to test setup) this will not
	// increase borrowed amount.
	err = s.app.LeverageKeeper.AccrueAllInterest(s.ctx)
	s.Require().NoError(err)

	// verify user's loan amount (40 umee)
	loanBalance = s.app.LeverageKeeper.GetBorrow(s.ctx, addr, umeeapp.BondDenom)
	s.Require().Equal(loanBalance, sdk.NewInt64Coin(umeeapp.BondDenom, 40000000))

	// borrow APY at utilization = 4%
	// when kink utilization = 80%, and base/kink APY are 0.02 and 0.22
	borrowAPY := s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(sdk.MustNewDecFromStr("0.03"), borrowAPY)

	// supply APY when borrow APY is 3%
	// and utilization is 4%, and reservefactor is 20%, and OracleRewardFactor is 1%
	// 0.03 * 0.04 * (1 - 0.21) = 0.000948
	supplyAPY := s.app.LeverageKeeper.DeriveSupplyAPY(s.ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdk.MustNewDecFromStr("0.000948"), supplyAPY)
}

func (s *IntegrationTestSuite) TestDynamicInterest() {
	// creates account which has supplied and collateralized 1000 UMEE
	addr := s.newAccount(coin(umeeDenom, 1000_000000))
	s.supply(addr, coin(umeeDenom, 1000_000000))
	s.collateralize(addr, coin("u/"+umeeDenom, 1000_000000))

	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("1.0")     // to allow high utilization
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("1.0") // to allow high utilization

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	// Base interest rate (0% utilization)
	rate := s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.02"))

	// user borrows 200 umee, utilization 200/1000
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	s.Require().NoError(err)

	// Between base interest and kink (20% utilization)
	rate = s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.07"))

	// user borrows 600 more umee, utilization 800/1000
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 600000000))
	s.Require().NoError(err)

	// Kink interest rate (80% utilization)
	rate = s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.22"))

	// user borrows 100 more umee, utilization 900/1000
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

	// Between kink interest and max (90% utilization)
	rate = s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("0.87"))

	// user borrows 100 more umee, utilization 1000/1000
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

	// Max interest rate (100% utilization)
	rate = s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.MustNewDecFromStr("1.52"))
}

func (s *IntegrationTestSuite) TestDynamicInterest_InvalidAsset() {
	rate := s.app.LeverageKeeper.DeriveBorrowAPY(s.ctx, "uabc")
	s.Require().Equal(rate, sdk.ZeroDec())
}
