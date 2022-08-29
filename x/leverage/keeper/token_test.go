package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) TestGetToken() {
	uabc := newToken("uabc", "ABC")
	s.Require().NoError(s.app.LeverageKeeper.SetTokenSettings(s.ctx, uabc))

	t, err := s.app.LeverageKeeper.GetTokenSettings(s.ctx, "uabc")
	s.Require().NoError(err)
	s.Require().Equal(t.ReserveFactor, sdk.MustNewDecFromStr("0.2"))
	s.Require().Equal(t.CollateralWeight, sdk.MustNewDecFromStr("0.25"))
	s.Require().Equal(t.LiquidationThreshold, sdk.MustNewDecFromStr("0.25"))
	s.Require().Equal(t.BaseBorrowRate, sdk.MustNewDecFromStr("0.02"))
	s.Require().Equal(t.KinkBorrowRate, sdk.MustNewDecFromStr("0.22"))
	s.Require().Equal(t.MaxBorrowRate, sdk.MustNewDecFromStr("1.52"))
	s.Require().Equal(t.KinkUtilization, sdk.MustNewDecFromStr("0.8"))
	s.Require().Equal(t.LiquidationIncentive, sdk.MustNewDecFromStr("0.1"))

	s.Require().NoError(t.AssertBorrowEnabled())
	s.Require().NoError(t.AssertSupplyEnabled())
	s.Require().NoError(t.AssertNotBlacklisted())
}
