package keeper_test

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"
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
		// This error matches oracle behavior on zero historic medians
		return sdk.ZeroDec(), 0, types.ErrNoHistoricMedians.Wrapf(
			"requested %d, got %d",
			numStamps,
			0,
		)
	}

	return p, uint32(numStamps), nil
}

func (m *mockOracleKeeper) GetExchangeRate(_ sdk.Context, denom string) (sdk.Dec, error) {
	p, ok := m.symbolExchangeRates[denom]
	if !ok {
		// This error matches oracle behavior on missing asset price
		return sdk.ZeroDec(), oracletypes.ErrUnknownDenom.Wrap(denom)
	}

	return p, nil
}

// Clear clears a denom from the mock oracle, simulating an outage.
func (m *mockOracleKeeper) Clear(denom string) {
	delete(m.symbolExchangeRates, denom)
	delete(m.historicExchangeRates, denom)
}

// Reset restores the mock oracle's prices to its default values.
func (m *mockOracleKeeper) Reset() {
	m.symbolExchangeRates = map[string]sdk.Dec{
		"UMEE": sdk.MustNewDecFromStr("4.21"),
		"ATOM": sdk.MustNewDecFromStr("39.38"),
		"DAI":  sdk.MustNewDecFromStr("1.00"),
		"DUMP": sdk.MustNewDecFromStr("0.50"), // A token which has recently halved in price
		"PUMP": sdk.MustNewDecFromStr("2.00"), // A token which has recently doubled in price
	}
	m.historicExchangeRates = map[string]sdk.Dec{
		"UMEE": sdk.MustNewDecFromStr("4.21"),
		"ATOM": sdk.MustNewDecFromStr("39.38"),
		"DAI":  sdk.MustNewDecFromStr("1.00"),
		"DUMP": sdk.MustNewDecFromStr("1.00"),
		"PUMP": sdk.MustNewDecFromStr("1.00"),
	}
}

func (s *IntegrationTestSuite) TestOracle_TokenPrice() {
	app, ctx, require := s.app, s.ctx, s.Require()

	p, e, err := app.LeverageKeeper.TokenPrice(ctx, appparams.BondDenom, types.PriceModeSpot)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("4.21"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenPrice(ctx, atomDenom, types.PriceModeSpot)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("39.38"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenPrice(ctx, "foo", types.PriceModeSpot)
	require.ErrorIs(err, types.ErrNotRegisteredToken)
	require.Equal(sdk.ZeroDec(), p)
	require.Equal(uint32(0), e)

	p, e, err = app.LeverageKeeper.TokenPrice(ctx, pumpDenom, types.PriceModeSpot)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("2.0"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenPrice(ctx, dumpDenom, types.PriceModeSpot)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("0.5"), p)
	require.Equal(uint32(6), e)

	// Now with historic = true

	p, e, err = app.LeverageKeeper.TokenPrice(ctx, appparams.BondDenom, types.PriceModeHistoric)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("4.21"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenPrice(ctx, atomDenom, types.PriceModeHistoric)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("39.38"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenPrice(ctx, "foo", types.PriceModeHistoric)
	require.ErrorIs(err, types.ErrNotRegisteredToken)
	require.Equal(sdk.ZeroDec(), p)
	require.Equal(uint32(0), e)

	p, e, err = app.LeverageKeeper.TokenPrice(ctx, pumpDenom, types.PriceModeHistoric)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.00"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenPrice(ctx, dumpDenom, types.PriceModeHistoric)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.00"), p)
	require.Equal(uint32(6), e)

	// Additional high/low cases

	p, e, err = app.LeverageKeeper.TokenPrice(ctx, pumpDenom, types.PriceModeHigh)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("2.00"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenPrice(ctx, dumpDenom, types.PriceModeHigh)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.00"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenPrice(ctx, pumpDenom, types.PriceModeLow)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.00"), p)
	require.Equal(uint32(6), e)

	p, e, err = app.LeverageKeeper.TokenPrice(ctx, dumpDenom, types.PriceModeLow)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("0.50"), p)
	require.Equal(uint32(6), e)

	// Lowercase must be converted to uppercase symbol denom ("DUMP" from "dump")
	p, e, err = app.LeverageKeeper.TokenPrice(ctx, strings.ToLower(dumpDenom), types.PriceModeLow)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("0.50"), p)
	require.Equal(uint32(6), e)
}

func (s *IntegrationTestSuite) TestOracle_TokenValue() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// 2.4 UMEE * $4.21
	v, err := app.LeverageKeeper.TokenValue(ctx, coin.New(appparams.BondDenom, 2_400000), types.PriceModeSpot)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("10.104"), v)

	v, err = app.LeverageKeeper.TokenValue(ctx, coin.New("foo", 2_400000), types.PriceModeSpot)
	require.ErrorIs(err, types.ErrNotRegisteredToken)
	require.Equal(sdk.ZeroDec(), v)

	// 2.4 DUMP * $0.5
	v, err = app.LeverageKeeper.TokenValue(ctx, coin.New(dumpDenom, 2_400000), types.PriceModeSpot)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.2"), v)

	// 2.4 PUMP * $2.00
	v, err = app.LeverageKeeper.TokenValue(ctx, coin.New(pumpDenom, 2_400000), types.PriceModeSpot)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("4.8"), v)

	// Now with historic = true

	// 2.4 UMEE * $4.21
	v, err = app.LeverageKeeper.TokenValue(ctx, coin.New(appparams.BondDenom, 2_400000), types.PriceModeHistoric)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("10.104"), v)

	v, err = app.LeverageKeeper.TokenValue(ctx, coin.New("foo", 2_400000), types.PriceModeHistoric)
	require.ErrorIs(err, types.ErrNotRegisteredToken)
	require.Equal(sdk.ZeroDec(), v)

	// 2.4 DUMP * $1.00
	v, err = app.LeverageKeeper.TokenValue(ctx, coin.New(dumpDenom, 2_400000), types.PriceModeHistoric)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("2.4"), v)

	// 2.4 PUMP * $1.00
	v, err = app.LeverageKeeper.TokenValue(ctx, coin.New(pumpDenom, 2_400000), types.PriceModeHistoric)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("2.4"), v)

	// Additional high/low cases

	// 2.4 DUMP * $1.00
	v, err = app.LeverageKeeper.TokenValue(ctx, coin.New(dumpDenom, 2_400000), types.PriceModeHigh)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("2.4"), v)

	// 2.4 PUMP * $2.00
	v, err = app.LeverageKeeper.TokenValue(ctx, coin.New(pumpDenom, 2_400000), types.PriceModeHigh)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("4.8"), v)

	// 2.4 DUMP * $0.50
	v, err = app.LeverageKeeper.TokenValue(ctx, coin.New(dumpDenom, 2_400000), types.PriceModeLow)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.2"), v)

	// 2.4 PUMP * $1.00
	v, err = app.LeverageKeeper.TokenValue(ctx, coin.New(pumpDenom, 2_400000), types.PriceModeLow)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("2.4"), v)

	// lowercase 2.4 PUMP * $1.00
	v, err = app.LeverageKeeper.TokenValue(ctx, coin.New(strings.ToLower(pumpDenom), 2_400000), types.PriceModeLow)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("2.4"), v)
}

func (s *IntegrationTestSuite) TestOracle_TotalTokenValue() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// (2.4 UMEE * $4.21) + (4.7 ATOM * $39.38)
	v, err := app.LeverageKeeper.TotalTokenValue(
		ctx,
		sdk.NewCoins(
			coin.New(appparams.BondDenom, 2_400000),
			coin.New(atomDenom, 4_700000),
		),
		types.PriceModeSpot,
	)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("195.19"), v)

	// same result, as unregistered token is ignored
	v, err = app.LeverageKeeper.TotalTokenValue(
		ctx,
		sdk.NewCoins(
			coin.New(appparams.BondDenom, 2_400000),
			coin.New(atomDenom, 4_700000),
			coin.New("foo", 4_700000),
		),
		types.PriceModeSpot,
	)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("195.19"), v)

	// complex historic case
	v, err = app.LeverageKeeper.TotalTokenValue(
		ctx,
		sdk.NewCoins(
			coin.New(appparams.BondDenom, 2_400000),
			coin.New(atomDenom, 4_700000),
			coin.New("foo", 4_700000),
			coin.New(dumpDenom, 2_000000),
		),
		types.PriceModeHistoric,
	)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("197.19"), v)

	// uses the higher price for each token
	v, err = app.LeverageKeeper.TotalTokenValue(
		ctx,
		sdk.NewCoins(
			coin.New(pumpDenom, 1_000000),
			coin.New(dumpDenom, 1_000000),
		),
		types.PriceModeHigh,
	)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("3.00"), v)

	// uses the lower price for each token
	v, err = app.LeverageKeeper.TotalTokenValue(
		ctx,
		sdk.NewCoins(
			coin.New(pumpDenom, 1_000000),
			coin.New(dumpDenom, 1_000000),
		),
		types.PriceModeLow,
	)
	require.NoError(err)
	require.Equal(sdk.MustNewDecFromStr("1.50"), v)
}

func (s *IntegrationTestSuite) TestOracle_PriceRatio() {
	app, ctx, require := s.app, s.ctx, s.Require()

	r, err := app.LeverageKeeper.PriceRatio(ctx, appparams.BondDenom, atomDenom, types.PriceModeSpot)
	require.NoError(err)
	// $4.21 / $39.38 at same exponent
	require.Equal(sdk.MustNewDecFromStr("0.106907059421025901"), r)

	r, err = app.LeverageKeeper.PriceRatio(ctx, appparams.BondDenom, daiDenom, types.PriceModeSpot)
	require.NoError(err)
	// $4.21 / $1.00 at a difference of 12 exponent
	require.Equal(sdk.MustNewDecFromStr("4210000000000"), r)

	r, err = app.LeverageKeeper.PriceRatio(ctx, daiDenom, appparams.BondDenom, types.PriceModeSpot)
	require.NoError(err)
	// $1.00 / $4.21 at a difference of -12 exponent
	require.Equal(sdk.MustNewDecFromStr("0.000000000000237530"), r)

	_, err = app.LeverageKeeper.PriceRatio(ctx, "foo", atomDenom, types.PriceModeSpot)
	require.ErrorIs(err, types.ErrNotRegisteredToken)

	_, err = app.LeverageKeeper.PriceRatio(ctx, appparams.BondDenom, "foo", types.PriceModeSpot)
	require.ErrorIs(err, types.ErrNotRegisteredToken)

	// current price of volatile assets
	r, err = app.LeverageKeeper.PriceRatio(ctx, pumpDenom, dumpDenom, types.PriceModeSpot)
	require.NoError(err)
	// $2.00 / $0.50
	require.Equal(sdk.MustNewDecFromStr("4"), r)
	// historic price of volatile assets
	r, err = app.LeverageKeeper.PriceRatio(ctx, pumpDenom, dumpDenom, types.PriceModeHistoric)
	require.NoError(err)
	// $1.00 / $1.00
	require.Equal(sdk.MustNewDecFromStr("1"), r)
}
