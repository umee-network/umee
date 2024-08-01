package keeper_test

import (
	sdkmath "cosmossdk.io/math"
)

func (s *IntegrationTestSuite) TestGetToken() {
	app, ctx, require := s.app, s.ctx, s.Require()

	uabc := newToken("uabc", "ABC", 6)
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, uabc))

	t, err := app.LeverageKeeper.GetTokenSettings(ctx, "uabc")
	require.NoError(err)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("0.2"), t.ReserveFactor)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("0.25"), t.CollateralWeight)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("0.26"), t.LiquidationThreshold)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("0.02"), t.BaseBorrowRate)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("0.22"), t.KinkBorrowRate)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("1.52"), t.MaxBorrowRate)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("0.8"), t.KinkUtilization)
	require.Equal(sdkmath.LegacyMustNewDecFromStr("0.1"), t.LiquidationIncentive)

	require.NoError(t.AssertBorrowEnabled())
	require.NoError(t.AssertSupplyEnabled())
	require.NoError(t.AssertNotBlacklisted())

	require.Equal(uint32(24), t.HistoricMedians)
}
