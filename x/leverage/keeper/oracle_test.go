package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v3/app/params"
	"github.com/umee-network/umee/v3/x/leverage/types"
)

type mockOracleKeeper struct {
	baseExchangeRates   map[string]sdk.Dec
	symbolExchangeRates map[string]sdk.Dec
	medianExchangeRates map[string]sdk.Dec
}

func newMockOracleKeeper() *mockOracleKeeper {
	m := &mockOracleKeeper{
		baseExchangeRates:   make(map[string]sdk.Dec),
		symbolExchangeRates: make(map[string]sdk.Dec),
		medianExchangeRates: make(map[string]sdk.Dec),
	}
	m.Reset()

	return m
}

// TODO: Does this function take base denom or symbol denom?
func (m *mockOracleKeeper) MedianOfHistoricMedians(ctx sdk.Context, denom string, numStamps uint64,
) (sdk.Dec, error) {
	p, ok := m.medianExchangeRates[denom]
	if !ok {
		return sdk.ZeroDec(), fmt.Errorf("invalid denom: %s", denom)
	}

	return p, nil
}

func (m *mockOracleKeeper) GetExchangeRate(_ sdk.Context, denom string) (sdk.Dec, error) {
	p, ok := m.symbolExchangeRates[denom]
	if !ok {
		return sdk.ZeroDec(), fmt.Errorf("invalid denom: %s", denom)
	}

	return p, nil
}

func (m *mockOracleKeeper) GetExchangeRateBase(ctx sdk.Context, denom string) (sdk.Dec, error) {
	p, ok := m.baseExchangeRates[denom]
	if !ok {
		return sdk.ZeroDec(), fmt.Errorf("invalid denom: %s", denom)
	}

	return p, nil
}

func (m *mockOracleKeeper) Reset() {
	m.symbolExchangeRates = map[string]sdk.Dec{
		"UMEE": sdk.MustNewDecFromStr("4.21"),
		"ATOM": sdk.MustNewDecFromStr("39.38"),
		"DAI":  sdk.MustNewDecFromStr("1.00"),
		"DUMP": sdk.MustNewDecFromStr("0.50"), // A token which has recently halved in price
		"PUMP": sdk.MustNewDecFromStr("2.00"), // A token which has recently doubled in price
	}
	m.baseExchangeRates = map[string]sdk.Dec{
		appparams.BondDenom: sdk.MustNewDecFromStr("0.00000421"),
		atomDenom:           sdk.MustNewDecFromStr("0.00003938"),
		daiDenom:            sdk.MustNewDecFromStr("0.000000000000000001"),
		"udump":             sdk.MustNewDecFromStr("0.0000005"),
		"upump":             sdk.MustNewDecFromStr("0.0000020"),
	}
	m.medianExchangeRates = map[string]sdk.Dec{
		"UMEE": sdk.MustNewDecFromStr("4.21"),
		"ATOM": sdk.MustNewDecFromStr("39.38"),
		"DAI":  sdk.MustNewDecFromStr("1.00"),
		"DUMP": sdk.MustNewDecFromStr("1.00"),
		"PUMP": sdk.MustNewDecFromStr("1.00"),
	}
}

func (s *IntegrationTestSuite) TestOracle_TokenBasePrice() {
	app, ctx, require := s.app, s.ctx, s.Require()

	p, err := app.LeverageKeeper.TokenBasePrice(ctx, appparams.BondDenom)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("0.00000421"), p)

	p, err = app.LeverageKeeper.TokenBasePrice(ctx, atomDenom)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("0.00003938"), p)

	p, err = app.LeverageKeeper.TokenBasePrice(ctx, "foo")
	require.ErrorIs(err, types.ErrNotRegisteredToken)
	require.Equal(sdk.ZeroDec(), p)
}

func (s *IntegrationTestSuite) TestOracle_TokenSymbolPrice() {
	app, ctx, require := s.app, s.ctx, s.Require()

	p, e, err := app.LeverageKeeper.TokenDefaultDenomPrice(ctx, appparams.BondDenom)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("4.21"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenDefaultDenomPrice(ctx, atomDenom)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("39.38"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenDefaultDenomPrice(ctx, "foo")
	require.ErrorIs(err, types.ErrNotRegisteredToken)
	require.Equal(sdk.ZeroDec(), p)
	require.Equal(uint32(0), e)
}

func (s *IntegrationTestSuite) TestOracle_TokenValue() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// 2.4 UMEE * $4.21
	v, err := app.LeverageKeeper.TokenValue(ctx, coin(appparams.BondDenom, 2_400000))
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("10.104"), v)

	v, err = app.LeverageKeeper.TokenValue(ctx, coin("foo", 2_400000))
	require.ErrorIs(err, types.ErrNotRegisteredToken)
	require.Equal(sdk.ZeroDec(), v)
}

func (s *IntegrationTestSuite) TestOracle_TotalTokenValue() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// (2.4 UMEE * $4.21) + (4.7 ATOM * $39.38)
	v, err := app.LeverageKeeper.TotalTokenValue(
		ctx,
		sdk.NewCoins(
			coin(appparams.BondDenom, 2_400000),
			coin(atomDenom, 4_700000),
		),
	)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("195.19"), v)

	// same result, as unregistered token is ignored
	v, err = app.LeverageKeeper.TotalTokenValue(
		ctx,
		sdk.NewCoins(
			coin(appparams.BondDenom, 2_400000),
			coin(atomDenom, 4_700000),
			coin("foo", 4_700000),
		),
	)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("195.19"), v)
}

func (s *IntegrationTestSuite) TestOracle_PriceRatio() {
	app, ctx, require := s.app, s.ctx, s.Require()

	r, err := app.LeverageKeeper.PriceRatio(ctx, appparams.BondDenom, atomDenom)
	require.NoError(err)
	// $4.21 / $39.38 at same exponent
	require.Equal(sdk.MustNewDecFromStr("0.106907059421025901"), r)

	r, err = app.LeverageKeeper.PriceRatio(ctx, appparams.BondDenom, daiDenom)
	require.NoError(err)
	// $4.21 / $1.00 at a difference of 12 exponent
	require.Equal(sdk.MustNewDecFromStr("4210000000000"), r)

	r, err = app.LeverageKeeper.PriceRatio(ctx, daiDenom, appparams.BondDenom)
	require.NoError(err)
	// $1.00 / $4.21 at a difference of -12 exponent
	require.Equal(sdk.MustNewDecFromStr("0.000000000000237530"), r)

	_, err = app.LeverageKeeper.PriceRatio(ctx, "foo", atomDenom)
	require.ErrorIs(err, types.ErrNotRegisteredToken)

	_, err = app.LeverageKeeper.PriceRatio(ctx, appparams.BondDenom, "foo")
	require.ErrorIs(err, types.ErrNotRegisteredToken)
}
