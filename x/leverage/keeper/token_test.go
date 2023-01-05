package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/x/leverage/types"
)

func (s *IntegrationTestSuite) TestGetToken() {
	app, ctx, require := s.app, s.ctx, s.Require()

	uabc := newToken("uabc", "ABC", 6)
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

	require.Equal(uint32(24), t.HistoricMedians)
}

func (s *IntegrationTestSuite) TestTokenNoMigration() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// Create an old token struct (had no HistoricMedians uint32 field)
	oldToken := types.OldToken{
		BaseDenom:              "uold",
		SymbolDenom:            "OLD",
		Exponent:               6,
		ReserveFactor:          sdk.MustNewDecFromStr("0.2"),
		CollateralWeight:       sdk.MustNewDecFromStr("0.25"),
		LiquidationThreshold:   sdk.MustNewDecFromStr("0.25"),
		BaseBorrowRate:         sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:         sdk.MustNewDecFromStr("0.22"),
		MaxBorrowRate:          sdk.MustNewDecFromStr("1.52"),
		KinkUtilization:        sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive:   sdk.MustNewDecFromStr("0.1"),
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdk.MustNewDecFromStr("1"),
		MaxSupplyUtilization:   sdk.MustNewDecFromStr("0.9"),
		MinCollateralLiquidity: sdk.MustNewDecFromStr("0"),
		MaxSupply:              sdk.NewInt(100_000_000000),
	}

	// Store the old token in state - simulating existing state
	require.NoError(app.LeverageKeeper.SetOldToken(ctx, oldToken))

	// Try to get it from state - using new token struct
	t, err := app.LeverageKeeper.GetTokenSettings(ctx, "uold")

	// Require no unmarshaling error
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

	// Check that historic medians starts at its default value of zero
	require.Equal(uint32(0), t.HistoricMedians)
}
