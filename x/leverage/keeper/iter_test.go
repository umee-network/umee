package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) TestGetEligibleLiquidationTargets_OneAddrOneAsset() {
	// creates account which has supplied and collateralized 1000 uumee
	addr := s.newAccount(coin(umeeDenom, 1000))
	s.supply(addr, coin(umeeDenom, 1000))
	s.collateralize(addr, coin("u/"+umeeDenom, 1000))

	// user borrows 250 umee (max current allowed)
	s.borrow(addr, coin(umeeDenom, 250))

	zeroAddresses, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{}, zeroAddresses)

	// Note: Setting umee liquidation threshold to 0.05 to make the user eligible to liquidation
	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("0.05")
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.05")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	targetAddress, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{addr}, targetAddress)
}

func (s *IntegrationTestSuite) TestGetEligibleLiquidationTargets_OneAddrTwoAsset() {
	// creates account which has supplied and collateralized 1000 uumee
	addr := s.newAccount(coin(umeeDenom, 1000))
	s.supply(addr, coin(umeeDenom, 1000))
	s.collateralize(addr, coin("u/"+umeeDenom, 1000))

	// user borrows 250 umee (max current allowed)
	s.borrow(addr, coin(umeeDenom, 250))

	zeroAddresses, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{}, zeroAddresses)

	// mints and send to addr 100 atom and already
	// enable 50 u/atom as collateral.
	s.fundAccount(addr, coin(atomDenom, 100_000000))
	s.supply(addr, coin(atomDenom, 50_000000))
	s.collateralize(addr, coin("u/"+atomDenom, 50_000000))

	// user borrows 4 atom (max current allowed - 1) user amount enabled as collateral * CollateralWeight
	// = (50 * 0.1) - 1
	// = 4
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(atomDenom, 4000000)) // 4 atom
	s.Require().NoError(err)

	// Note: Setting umee liquidation threshold to 0.05 to make the user eligible for liquidation
	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("0.05")
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.05")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	// Note: Setting atom collateral weight to 0.01 to make the user eligible for liquidation
	atomIBCToken := newToken(atomDenom, "ATOM")
	atomIBCToken.CollateralWeight = sdk.MustNewDecFromStr("0.01")
	atomIBCToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.01")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, atomIBCToken))

	targetAddr, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{addr}, targetAddr)
}

func (s *IntegrationTestSuite) TestGetEligibleLiquidationTargets_TwoAddr() {
	// creates account which has supplied and collateralized 1000 uumee
	addr := s.newAccount(coin(umeeDenom, 1000))
	s.supply(addr, coin(umeeDenom, 1000))
	s.collateralize(addr, coin("u/"+umeeDenom, 1000))

	// user borrows 250 umee (max current allowed)
	s.borrow(addr, coin(umeeDenom, 250))

	zeroAddresses, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{}, zeroAddresses)

	// creates another account which has supplied and collateralized 100 uatom
	addr2 := s.newAccount(coin(atomDenom, 100))
	s.supply(addr2, coin(atomDenom, 100))
	s.collateralize(addr2, coin("u/"+atomDenom, 100))

	// borrows atom (max current allowed - 1)
	s.borrow(addr2, coin(atomDenom, 24))

	// Note: Setting umee liquidation threshold to 0.05 to make the first supplier eligible for liquidation
	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("0.05")
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.05")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	// Note: Setting atom collateral weight to 0.01 to make the second supplier eligible for liquidation
	atomIBCToken := newToken(atomDenom, "ATOM")
	atomIBCToken.CollateralWeight = sdk.MustNewDecFromStr("0.01")
	atomIBCToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.01")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, atomIBCToken))

	targets, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{addr, addr2}, targets)
}
