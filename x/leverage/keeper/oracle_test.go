package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

type mockOracleKeeper struct {
	baseExchangeRates     map[string]sdk.Dec
	symbolExchangeRates   map[string]sdk.Dec
	historicExchangeRates map[string]sdk.Dec
}

func newMockOracleKeeper() *mockOracleKeeper {
	m := &mockOracleKeeper{
		baseExchangeRates:     make(map[string]sdk.Dec),
		symbolExchangeRates:   make(map[string]sdk.Dec),
		historicExchangeRates: make(map[string]sdk.Dec),
	}
	m.Reset()

	return m
}

func (m *mockOracleKeeper) MedianOfHistoricMedians(ctx sdk.Context, denom string, numStamps uint64,
) (sdk.Dec, uint32, error) {
	p, ok := m.historicExchangeRates[denom]
	if !ok {
		return sdk.ZeroDec(), 0, fmt.Errorf("invalid denom: %s", denom)
	}

	return p, 24, nil
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
		dumpDenom:           sdk.MustNewDecFromStr("0.0000005"),
		pumpDenom:           sdk.MustNewDecFromStr("0.0000020"),
	}
	m.historicExchangeRates = map[string]sdk.Dec{
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

	p, e, err := app.LeverageKeeper.TokenDefaultDenomPrice(ctx, appparams.BondDenom, false)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("4.21"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenDefaultDenomPrice(ctx, atomDenom, false)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("39.38"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenDefaultDenomPrice(ctx, "foo", false)
	require.ErrorIs(err, types.ErrNotRegisteredToken)
	require.Equal(sdk.ZeroDec(), p)
	require.Equal(uint32(0), e)

	p, e, err = app.LeverageKeeper.TokenDefaultDenomPrice(ctx, pumpDenom, false)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("2.0"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenDefaultDenomPrice(ctx, dumpDenom, false)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("0.5"), p)
	require.Equal(uint32(6), e)

	// Now with historic = true

	p, e, err = app.LeverageKeeper.TokenDefaultDenomPrice(ctx, appparams.BondDenom, true)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("4.21"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenDefaultDenomPrice(ctx, atomDenom, true)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("39.38"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenDefaultDenomPrice(ctx, "foo", true)
	require.ErrorIs(err, types.ErrNotRegisteredToken)
	require.Equal(sdk.ZeroDec(), p)
	require.Equal(uint32(0), e)

	p, e, err = app.LeverageKeeper.TokenDefaultDenomPrice(ctx, pumpDenom, true)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.00"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenDefaultDenomPrice(ctx, dumpDenom, true)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.00"), p)
	require.Equal(uint32(6), e)
}

func (s *IntegrationTestSuite) TestOracle_TokenValue() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// 2.4 UMEE * $4.21
	v, err := app.LeverageKeeper.TokenValue(ctx, coin(appparams.BondDenom, 2_400000), false)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("10.104"), v)

	v, err = app.LeverageKeeper.TokenValue(ctx, coin("foo", 2_400000), false)
	require.ErrorIs(err, types.ErrNotRegisteredToken)
	require.Equal(sdk.ZeroDec(), v)

	// 2.4 DUMP * $0.5
	v, err = app.LeverageKeeper.TokenValue(ctx, coin(dumpDenom, 2_400000), false)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.2"), v)

	// 2.4 PUMP * $2.00
	v, err = app.LeverageKeeper.TokenValue(ctx, coin(pumpDenom, 2_400000), false)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("4.8"), v)

	// Now with historic = true

	// 2.4 UMEE * $4.21
	v, err = app.LeverageKeeper.TokenValue(ctx, coin(appparams.BondDenom, 2_400000), true)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("10.104"), v)

	v, err = app.LeverageKeeper.TokenValue(ctx, coin("foo", 2_400000), true)
	require.ErrorIs(err, types.ErrNotRegisteredToken)
	require.Equal(sdk.ZeroDec(), v)

	// 2.4 DUMP * $1.00
	v, err = app.LeverageKeeper.TokenValue(ctx, coin(dumpDenom, 2_400000), true)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("2.4"), v)

	// 2.4 PUMP * $1.00
	v, err = app.LeverageKeeper.TokenValue(ctx, coin(pumpDenom, 2_400000), true)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("2.4"), v)
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
		false,
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
		false,
	)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("195.19"), v)

	// complex historic case
	v, err = app.LeverageKeeper.TotalTokenValue(
		ctx,
		sdk.NewCoins(
			coin(appparams.BondDenom, 2_400000),
			coin(atomDenom, 4_700000),
			coin("foo", 4_700000),
			coin(dumpDenom, 2_000000),
		),
		true,
	)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("197.19"), v)
}

func (s *IntegrationTestSuite) TestOracle_PriceRatio() {
	app, ctx, require := s.app, s.ctx, s.Require()

	r, err := app.LeverageKeeper.PriceRatio(ctx, appparams.BondDenom, atomDenom, false)
	require.NoError(err)
	// $4.21 / $39.38 at same exponent
	require.Equal(sdk.MustNewDecFromStr("0.106907059421025901"), r)

	r, err = app.LeverageKeeper.PriceRatio(ctx, appparams.BondDenom, daiDenom, false)
	require.NoError(err)
	// $4.21 / $1.00 at a difference of 12 exponent
	require.Equal(sdk.MustNewDecFromStr("4210000000000"), r)

	r, err = app.LeverageKeeper.PriceRatio(ctx, daiDenom, appparams.BondDenom, false)
	require.NoError(err)
	// $1.00 / $4.21 at a difference of -12 exponent
	require.Equal(sdk.MustNewDecFromStr("0.000000000000237530"), r)

	_, err = app.LeverageKeeper.PriceRatio(ctx, "foo", atomDenom, false)
	require.ErrorIs(err, types.ErrNotRegisteredToken)

	_, err = app.LeverageKeeper.PriceRatio(ctx, appparams.BondDenom, "foo", false)
	require.ErrorIs(err, types.ErrNotRegisteredToken)

	// current price of volatile assets
	r, err = app.LeverageKeeper.PriceRatio(ctx, pumpDenom, dumpDenom, false)
	require.NoError(err)
	// $2.00 / $0.50
	require.Equal(sdk.MustNewDecFromStr("4"), r)
	// historic price of volatile assets
	r, err = app.LeverageKeeper.PriceRatio(ctx, pumpDenom, dumpDenom, true)
	require.NoError(err)
	// $1.00 / $1.00
	require.Equal(sdk.MustNewDecFromStr("1"), r)
}
