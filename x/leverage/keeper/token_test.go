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
	require.Equal(t.ReserveFactor, sdk.MustNewDecFromStr("0.2"))
	require.Equal(t.CollateralWeight, sdk.MustNewDecFromStr("0.25"))
	require.Equal(t.LiquidationThreshold, sdk.MustNewDecFromStr("0.25"))
	require.Equal(t.BaseBorrowRate, sdk.MustNewDecFromStr("0.02"))
	require.Equal(t.KinkBorrowRate, sdk.MustNewDecFromStr("0.22"))
	require.Equal(t.MaxBorrowRate, sdk.MustNewDecFromStr("1.52"))
	require.Equal(t.KinkUtilization, sdk.MustNewDecFromStr("0.8"))
	require.Equal(t.LiquidationIncentive, sdk.MustNewDecFromStr("0.1"))

	require.NoError(t.AssertBorrowEnabled())
	require.NoError(t.AssertSupplyEnabled())
	require.NoError(t.AssertNotBlacklisted())
}
