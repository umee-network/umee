package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	umeeapp "github.com/umee-network/umee/v2/app"
)

func (s *IntegrationTestSuite) TestGetEligibleLiquidationTargets_OneAddrOneAsset() {
	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee enabled as collateral.
	addr, _ := s.initBorrowScenario()

	// user borrows 100 umee (max current allowed) user amount enabled as collateral * CollateralWeight
	// = 1000 * 0.1
	// = 100
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

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

	// if it tries to borrow any other asset it should return an error
	err = s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(atomDenom, 1))
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestGetEligibleLiquidationTargets_OneAddrTwoAsset() {
	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee enabled as collateral.
	addr, _ := s.initBorrowScenario()

	// user borrows 100 umee (max current allowed) user amount enabled as collateral * CollateralWeight
	// = 1000 * 0.1
	// = 100
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

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
	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee enabled as collateral.
	supplierAddr, anotherSupplier := s.initBorrowScenario()

	// supplier borrows 100 umee (max current allowed) supplier amount enabled as collateral * CollateralWeight
	// = 1000 * 0.1
	// = 100
	err := s.app.LeverageKeeper.Borrow(s.ctx, supplierAddr, sdk.NewInt64Coin(umeeapp.BondDenom, 100000000))
	s.Require().NoError(err)

	zeroAddresses, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{}, zeroAddresses)

	// mints and send to anotherSupplier 100 atom and already
	// enable 50 u/atom as collateral.
	s.fundAccount(anotherSupplier, coin(atomDenom, 100_000000))
	s.supply(anotherSupplier, coin(atomDenom, 50_000000))
	s.collateralize(anotherSupplier, coin("u/"+atomDenom, 50_000000))

	// anotherSupplier borrows 4 atom (max current allowed - 1) anotherSupplier amount enabled as collateral * CollateralWeight
	// = (50 * 0.1) - 1
	// = 4
	err = s.app.LeverageKeeper.Borrow(s.ctx, anotherSupplier, sdk.NewInt64Coin(atomDenom, 4000000)) // 4 atom
	s.Require().NoError(err)

	// Note: Setting umee liquidation threshold to 0.05 to make the supplier eligible for liquidation
	umeeToken := newToken("uumee", "UMEE")
	umeeToken.CollateralWeight = sdk.MustNewDecFromStr("0.05")
	umeeToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.05")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, umeeToken))

	// Note: Setting atom collateral weight to 0.01 to make the supplier eligible for liquidation
	atomIBCToken := newToken(atomDenom, "ATOM")
	atomIBCToken.CollateralWeight = sdk.MustNewDecFromStr("0.01")
	atomIBCToken.LiquidationThreshold = sdk.MustNewDecFromStr("0.01")

	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, atomIBCToken))

	supplierAddress, err := s.app.LeverageKeeper.GetEligibleLiquidationTargets(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.AccAddress{supplierAddr, anotherSupplier}, supplierAddress)
}
