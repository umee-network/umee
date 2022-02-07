package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) TestGetCollateralAmount() {
	uDenom := s.tk.FromTokenToUTokenDenom(s.ctx, umeeDenom)

	// get u/umee collateral amount of empty account address (zero)
	collateral := s.tk.GetCollateralAmount(s.ctx, sdk.AccAddress{}, uDenom)
	s.Require().Equal(sdk.NewInt64Coin(uDenom, 0), collateral)

	// get u/umee collateral amount of non-empty account address (zero)
	collateral = s.tk.GetCollateralAmount(s.ctx, sdk.AccAddress{0x01}, uDenom)
	s.Require().Equal(sdk.NewInt64Coin(uDenom, 0), collateral)

	// creates account which has 1000 u/umee but not enabled as collateral
	addr := s.setupAccount(umeeDenom, 1000, 1000, 0, false)

	// confirm collateral amount is 0 u/uumee
	collateral = s.tk.GetCollateralAmount(s.ctx, addr, uDenom)
	s.Require().Equal(sdk.NewInt64Coin(uDenom, 0), collateral)

	// enable u/umee as collateral
	s.Require().NoError(s.tk.SetCollateralSetting(s.ctx, addr, uDenom, true))

	// confirm collateral amount is 1000 u/uumee
	collateral = s.tk.GetCollateralAmount(s.ctx, addr, uDenom)
	s.Require().Equal(sdk.NewInt64Coin(uDenom, 1000), collateral)

	// collateral amount of non-utoken denom (zero)
	collateral = s.tk.GetCollateralAmount(s.ctx, addr, umeeDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 0), collateral)

	// collateral amount of unregistered denom (zero)
	collateral = s.tk.GetCollateralAmount(s.ctx, addr, "abcd")
	s.Require().Equal(sdk.NewInt64Coin("abcd", 0), collateral)

	// disable u/umee as collateral
	s.Require().NoError(s.tk.SetCollateralSetting(s.ctx, addr, uDenom, false))

	// confirm collateral amount is 0 u/uumee
	collateral = s.tk.GetCollateralAmount(s.ctx, addr, uDenom)
	s.Require().Equal(sdk.NewInt64Coin(uDenom, 0), collateral)

	// we do not test empty denom, as that will cause a panic
}

func (s *IntegrationTestSuite) TestSetCollateralAmount() {
	uDenom := s.tk.FromTokenToUTokenDenom(s.ctx, umeeDenom)

	// set u/umee collateral amount of empty account address (error)
	err := s.tk.SetCollateralAmount(s.ctx, sdk.AccAddress{}, sdk.NewInt64Coin(uDenom, 0))
	s.Require().EqualError(err, "empty address")

	addr := sdk.AccAddress{0x01}

	// force invalid denom
	err = s.tk.SetCollateralAmount(s.ctx, addr, sdk.Coin{Denom: "", Amount: sdk.ZeroInt()})
	s.Require().EqualError(err, "0: invalid asset")

	// set u/umee collateral amount
	s.Require().NoError(s.tk.SetCollateralAmount(s.ctx, addr, sdk.NewInt64Coin(uDenom, 10)))

	// confirm effect
	collateral := s.tk.GetCollateralAmount(s.ctx, addr, uDenom)
	s.Require().Equal(sdk.NewInt64Coin(uDenom, 10), collateral)

	// set u/umee collateral amount to zero
	s.Require().NoError(s.tk.SetCollateralAmount(s.ctx, addr, sdk.NewInt64Coin(uDenom, 0)))

	// confirm effect
	collateral = s.tk.GetCollateralAmount(s.ctx, addr, uDenom)
	s.Require().Equal(sdk.NewInt64Coin(uDenom, 0), collateral)

	// set unregistered token collateral amount
	s.Require().NoError(s.tk.SetCollateralAmount(s.ctx, addr, sdk.NewInt64Coin("abcd", 10)))

	// confirm effect
	collateral = s.tk.GetCollateralAmount(s.ctx, addr, "abcd")
	s.Require().Equal(sdk.NewInt64Coin("abcd", 10), collateral)

	// set unregistered token collateral amount to zero
	s.Require().NoError(s.tk.SetCollateralAmount(s.ctx, addr, sdk.NewInt64Coin("abcd", 0)))

	// confirm effect
	collateral = s.tk.GetCollateralAmount(s.ctx, addr, "abcd")
	s.Require().Equal(sdk.NewInt64Coin("abcd", 0), collateral)

	// we do not test empty denom, as that will cause a panic
}
