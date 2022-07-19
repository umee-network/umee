package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) TestGetBorrow() {
	// get uumee borrow amount of empty account address (zero)
	borrowed := s.tk.GetBorrow(s.ctx, sdk.AccAddress{}, umeeDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 0), borrowed)

	// creates account which has borrowed 123 uumee
	addr := s.setupAccount(umeeDenom, 1000, 1000, 123, true)

	// confirm borrowed amount is 123 uumee
	borrowed = s.tk.GetBorrow(s.ctx, addr, umeeDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 123), borrowed)

	// unregistered denom (zero)
	borrowed = s.tk.GetBorrow(s.ctx, addr, "abcd")
	s.Require().Equal(sdk.NewInt64Coin("abcd", 0), borrowed)

	// we do not test empty denom, as that will cause a panic
}

func (s *IntegrationTestSuite) TestSetBorrow() {
	// empty account address
	err := s.tk.SetBorrow(s.ctx, sdk.AccAddress{}, sdk.NewInt64Coin(umeeDenom, 123))
	s.Require().EqualError(err, "empty address")

	addr := sdk.AccAddress{0x00}

	// set nonzero borrow, and confirm effect
	err = s.tk.SetBorrow(s.ctx, addr, sdk.NewInt64Coin(umeeDenom, 123))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 123), s.tk.GetBorrow(s.ctx, addr, umeeDenom))

	// set zero borrow, and confirm effect
	err = s.tk.SetBorrow(s.ctx, addr, sdk.NewInt64Coin(umeeDenom, 0))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 0), s.tk.GetBorrow(s.ctx, addr, umeeDenom))

	// unregistered (but valid) denom
	err = s.tk.SetBorrow(s.ctx, addr, sdk.NewInt64Coin("abcd", 123))
	s.Require().NoError(err)

	// interest scalar test - ensure borrowing smallest possible amount doesn't round to zero at scalar = 1.0001
	s.Require().NoError(s.tk.SetInterestScalar(s.ctx, umeeDenom, sdk.MustNewDecFromStr("1.0001")))
	s.Require().NoError(s.tk.SetBorrow(s.ctx, addr, sdk.NewInt64Coin(umeeDenom, 1)))
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 1), s.tk.GetBorrow(s.ctx, addr, umeeDenom))

	// interest scalar test - scalar changing after borrow (as it does when interest accrues)
	s.Require().NoError(s.tk.SetInterestScalar(s.ctx, umeeDenom, sdk.MustNewDecFromStr("1.0")))
	s.Require().NoError(s.tk.SetBorrow(s.ctx, addr, sdk.NewInt64Coin(umeeDenom, 200)))
	s.Require().NoError(s.tk.SetInterestScalar(s.ctx, umeeDenom, sdk.MustNewDecFromStr("2.33")))
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 466), s.tk.GetBorrow(s.ctx, addr, umeeDenom))

	// interest scalar extreme case - rounding up becomes apparent at high borrow amount
	s.Require().NoError(s.tk.SetInterestScalar(s.ctx, umeeDenom, sdk.MustNewDecFromStr("555444333222111")))
	s.Require().NoError(s.tk.SetBorrow(s.ctx, addr, sdk.NewInt64Coin(umeeDenom, 1)))
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 1), s.tk.GetBorrow(s.ctx, addr, umeeDenom))
	s.Require().NoError(s.tk.SetBorrow(s.ctx, addr, sdk.NewInt64Coin(umeeDenom, 20000)))
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 20001), s.tk.GetBorrow(s.ctx, addr, umeeDenom))

	// we do not test empty denom, as that will cause a panic
}

func (s *IntegrationTestSuite) TestGetTotalBorrowed() {
	// unregistered denom (zero)
	borrowed := s.tk.GetTotalBorrowed(s.ctx, "abcd")
	s.Require().Equal(sdk.NewInt64Coin("abcd", 0), borrowed)

	// creates account which has borrowed 123 uumee
	_ = s.setupAccount(umeeDenom, 1000, 1000, 123, true)

	// confirm total borrowed amount is 123 uumee
	borrowed = s.tk.GetTotalBorrowed(s.ctx, umeeDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 123), borrowed)

	// creates account which has borrowed 456 uumee
	_ = s.setupAccount(umeeDenom, 2000, 2000, 456, true)

	// confirm total borrowed amount is 579 uumee
	borrowed = s.tk.GetTotalBorrowed(s.ctx, umeeDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 579), borrowed)

	// interest scalar test - scalar changing after borrow (as it does when interest accrues)
	s.Require().NoError(s.tk.SetInterestScalar(s.ctx, umeeDenom, sdk.MustNewDecFromStr("2.00")))
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 1158), s.tk.GetTotalBorrowed(s.ctx, umeeDenom))

	// we do not test empty denom, as that will cause a panic
}

func (s *IntegrationTestSuite) TestGetAvailableToBorrow() {
	// unregistered denom (zero)
	available := s.tk.GetAvailableToBorrow(s.ctx, "abcd")
	s.Require().Equal(sdk.ZeroInt(), available)

	// creates account which has supplied 1000 uumee, and borrowed 0 uumee
	_ = s.setupAccount(umeeDenom, 1000, 1000, 0, true)

	// confirm lending pool is 1000 uumee
	available = s.tk.GetAvailableToBorrow(s.ctx, umeeDenom)
	s.Require().Equal(sdk.NewInt(1000), available)

	// creates account which has supplied 1000 uumee, and borrowed 123 uumee
	_ = s.setupAccount(umeeDenom, 1000, 1000, 123, true)

	// confirm lending pool is 1877 uumee
	available = s.tk.GetAvailableToBorrow(s.ctx, umeeDenom)
	s.Require().Equal(sdk.NewInt(1877), available)

	// artificially reserve 200 uumee, reducing available amount to 1677
	s.Require().NoError(s.tk.SetReserveAmount(s.ctx, sdk.NewInt64Coin(umeeDenom, 200)))
	available = s.tk.GetAvailableToBorrow(s.ctx, umeeDenom)
	s.Require().Equal(sdk.NewInt(1677), available)

	// we do not test empty denom, as that will cause a panic
}

func (s *IntegrationTestSuite) TestDeriveSupplyUtilization() {
	// unregistered denom (0% utilization)
	utilization := s.tk.SupplyUtilization(s.ctx, "abcd")
	s.Require().Equal(sdk.ZeroDec(), utilization)

	// creates account which has supplied 1000 uumee, and borrowed 0 uumee
	addr := s.setupAccount(umeeDenom, 1000, 1000, 0, true)

	// All tests below are commented with the following equation in mind:
	//   utilization = (Total Borrowed / (Total Borrowed + Module Balance - Reserved Amount))

	// 0% utilization (0 / 0+1000-0)
	utilization = s.tk.SupplyUtilization(s.ctx, umeeDenom)
	s.Require().Equal(sdk.ZeroDec(), utilization)

	// user borrows 200 uumee, reducing module account to 800 uumee
	s.Require().NoError(s.tk.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeDenom, 200)))

	// 20% utilization (200 / 200+800-0)
	utilization = s.tk.SupplyUtilization(s.ctx, umeeDenom)
	s.Require().Equal(sdk.MustNewDecFromStr("0.2"), utilization)

	// artificially reserve 200 uumee
	s.Require().NoError(s.tk.SetReserveAmount(s.ctx, sdk.NewInt64Coin(umeeDenom, 200)))

	// 25% utilization (200 / 200+800-200)
	utilization = s.tk.SupplyUtilization(s.ctx, umeeDenom)
	s.Require().Equal(sdk.MustNewDecFromStr("0.25"), utilization)

	// Setting umee collateral weight to 1.0 to allow user to borrow heavily
	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("1")
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("1")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	// user borrows 600 uumee, reducing module account to 0 uumee
	s.Require().NoError(s.tk.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeDenom, 600)))

	// 100% utilization (800 / 800+200-200))
	utilization = s.tk.SupplyUtilization(s.ctx, umeeDenom)
	s.Require().Equal(sdk.MustNewDecFromStr("1.0"), utilization)

	// artificially set user borrow to 1200 umee
	s.Require().NoError(s.tk.SetBorrow(s.ctx, addr, sdk.NewInt64Coin(umeeDenom, 1200)))

	// still 100% utilization (1200 / 1200+200-200)
	utilization = s.tk.SupplyUtilization(s.ctx, umeeDenom)
	s.Require().Equal(sdk.MustNewDecFromStr("1.0"), utilization)

	// artificially set reserves to 800 uumee
	s.Require().NoError(s.tk.SetReserveAmount(s.ctx, sdk.NewInt64Coin(umeeDenom, 800)))

	// edge case interpreted as 200% utilization (1200 / 1200+200-800)
	utilization = s.tk.SupplyUtilization(s.ctx, umeeDenom)
	s.Require().Equal(sdk.MustNewDecFromStr("2.0"), utilization)

	// artificially set reserves to 4000 uumee
	s.Require().NoError(s.tk.SetReserveAmount(s.ctx, sdk.NewInt64Coin(umeeDenom, 4000)))

	// impossible case interpreted as near-infinite utilization (1200 / 1200+200-4000)
	utilization = s.tk.SupplyUtilization(s.ctx, umeeDenom)
	s.Require().Equal(sdk.MaxSortableDec, utilization)
}

func (s *IntegrationTestSuite) TestCalculateBorrowLimit() {
	// Empty coins
	borrowLimit, err := s.app.LeverageKeeper.CalculateBorrowLimit(s.ctx, sdk.NewCoins())
	s.Require().NoError(err)
	s.Require().Equal(sdk.ZeroDec(), borrowLimit)

	// Unregistered asset
	invalidCoins := sdk.NewCoins(sdk.NewInt64Coin("abcd", 1000))
	_, err = s.app.LeverageKeeper.CalculateBorrowLimit(s.ctx, invalidCoins)
	s.Require().EqualError(err, "abcd: invalid asset")

	// Create collateral uTokens (1k u/umee)
	umeeCollatDenom := s.app.LeverageKeeper.FromTokenToUTokenDenom(s.ctx, umeeDenom)
	umeeCollateral := sdk.NewCoins(sdk.NewInt64Coin(umeeCollatDenom, 1000000000))

	// Manually compute borrow limit using collateral weight of 0.25
	// and placeholder of 1 umee = $4.21.
	expectedUmeeLimit := umeeCollateral[0].Amount.ToDec().
		Mul(sdk.MustNewDecFromStr("0.00000421")).
		Mul(sdk.MustNewDecFromStr("0.25"))

	// Check borrow limit vs. manually computed value
	borrowLimit, err = s.app.LeverageKeeper.CalculateBorrowLimit(s.ctx, umeeCollateral)
	s.Require().NoError(err)
	s.Require().Equal(expectedUmeeLimit, borrowLimit)

	// Create collateral atom uTokens (1k u/uatom)
	atomCollatDenom := s.app.LeverageKeeper.FromTokenToUTokenDenom(s.ctx, atomIBCDenom)
	atomCollateral := sdk.NewCoins(sdk.NewInt64Coin(atomCollatDenom, 1000000000))

	// Manually compute borrow limit using collateral weight of 0.25
	// and placeholder of 1 atom = $39.38
	expectedAtomLimit := atomCollateral[0].Amount.ToDec().
		Mul(sdk.MustNewDecFromStr("0.00003938")).
		Mul(sdk.MustNewDecFromStr("0.25"))

	// Check borrow limit vs. manually computed value
	borrowLimit, err = s.app.LeverageKeeper.CalculateBorrowLimit(s.ctx, atomCollateral)
	s.Require().NoError(err)
	s.Require().Equal(expectedAtomLimit, borrowLimit)

	// Compute the expected borrow limit of the two combined collateral coins
	expectedCombinedLimit := expectedUmeeLimit.Add(expectedAtomLimit)
	combinedCollateral := umeeCollateral.Add(atomCollateral...)

	// Check borrow limit vs. manually computed value
	borrowLimit, err = s.app.LeverageKeeper.CalculateBorrowLimit(s.ctx, combinedCollateral)
	s.Require().NoError(err)
	s.Require().Equal(expectedCombinedLimit, borrowLimit)
}
