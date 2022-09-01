package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) TestGetToken() {
	app, ctx, require := s.app, s.ctx, s.Require()

	uabc := newToken("uabc", "ABC")
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, uabc))

	t, err := app.LeverageKeeper.GetTokenSettings(ctx, "uabc")
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("0.2"), t.ReserveFactor)
	require.Equal(sdk.MustNewDecFromStr("0.25"), t.CollateralWeight)
	require.Equal(sdk.MustNewDecFromStr("0.25"), t.LiquidationThreshold)
	require.Equal(sdk.MustNewDecFromStr("0.02"), t.BaseBorrowRate)
	require.Equal(sdk.MustNewDecFromStr("0.22"), t.KinkBorrowRate)
	require.Equal(sdk.MustNewDecFromStr("1.52"), t.MaxBorrowRate)
	require.Equal(sdk.MustNewDecFromStr("0.8"), t.KinkUtilization)
	require.Equal(sdk.MustNewDecFromStr("0.1"), t.LiquidationIncentive)

	require.NoError(t.AssertBorrowEnabled())
	require.NoError(t.AssertSupplyEnabled())
	require.NoError(t.AssertNotBlacklisted())
}
