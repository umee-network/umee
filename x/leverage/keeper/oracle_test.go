package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	umeeapp "github.com/umee-network/umee/v3/app"
)

type mockOracleKeeper struct {
	exchangeRates map[string]sdk.Dec
}

func newMockOracleKeeper() *mockOracleKeeper {
	m := &mockOracleKeeper{
		exchangeRates: make(map[string]sdk.Dec),
	}
	m.Reset()

	return m
}

func (m *mockOracleKeeper) GetExchangeRate(_ sdk.Context, denom string) (sdk.Dec, error) {
	p, ok := m.exchangeRates[denom]
	if !ok {
		return sdk.ZeroDec(), fmt.Errorf("invalid denom: %s", denom)
	}

	return p, nil
}

func (m *mockOracleKeeper) GetExchangeRateBase(ctx sdk.Context, denom string) (sdk.Dec, error) {
	p, err := m.GetExchangeRate(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	// assume 10^6 for the base denom
	return p.Quo(sdk.MustNewDecFromStr("1000000.00")), nil
}

func (m *mockOracleKeeper) Reset() {
	m.exchangeRates = map[string]sdk.Dec{
		umeeapp.BondDenom: sdk.MustNewDecFromStr("4.21"),
		atomDenom:         sdk.MustNewDecFromStr("39.38"),
	}
}

func (s *IntegrationTestSuite) TestOracle_TokenPrice() {
	app, ctx, require := s.app, s.ctx, s.Require()

	p, err := app.LeverageKeeper.TokenPrice(ctx, umeeapp.BondDenom)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("0.00000421"), p)

	p, err = app.LeverageKeeper.TokenPrice(ctx, atomDenom)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("0.00003938"), p)

	p, err = app.LeverageKeeper.TokenPrice(ctx, "foo")
	require.Error(err)
	require.Equal(sdk.ZeroDec(), p)
}

func (s *IntegrationTestSuite) TestOracle_TokenValue() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// 2.4 UMEE * $4.21
	v, err := app.LeverageKeeper.TokenValue(ctx, sdk.NewInt64Coin(umeeapp.BondDenom, 2400000))
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("10.104"), v)

	v, err = app.LeverageKeeper.TokenValue(ctx, sdk.NewInt64Coin("foo", 2400000))
	require.Error(err)
	require.Equal(sdk.ZeroDec(), v)
}

func (s *IntegrationTestSuite) TestOracle_TotalTokenValue() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// (2.4 UMEE * $4.21) + (4.7 ATOM * $39.38)
	v, err := app.LeverageKeeper.TotalTokenValue(
		ctx,
		sdk.NewCoins(
			sdk.NewInt64Coin(umeeapp.BondDenom, 2400000),
			sdk.NewInt64Coin(atomDenom, 4700000),
		),
	)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("195.19"), v)

	// same result, as unregistered token is ignored
	v, err = app.LeverageKeeper.TotalTokenValue(
		ctx,
		sdk.NewCoins(
			sdk.NewInt64Coin(umeeapp.BondDenom, 2400000),
			sdk.NewInt64Coin(atomDenom, 4700000),
			sdk.NewInt64Coin("foo", 4700000),
		),
	)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("195.19"), v)
}

func (s *IntegrationTestSuite) TestOracle_PriceRatio() {
	app, ctx, require := s.app, s.ctx, s.Require()

	r, err := app.LeverageKeeper.PriceRatio(ctx, umeeapp.BondDenom, atomDenom)
	require.NoError(err)
	// $4.21 / $39.38
	require.Equal(sdk.MustNewDecFromStr("0.106907059421025901"), r)

	_, err = app.LeverageKeeper.PriceRatio(ctx, "foo", atomDenom)
	require.Error(err)

	_, err = app.LeverageKeeper.PriceRatio(ctx, umeeapp.BondDenom, "foo")
	require.Error(err)
}
