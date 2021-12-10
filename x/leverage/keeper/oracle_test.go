package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	umeeapp "github.com/umee-network/umee/app"
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
		atomIBCDenom:      sdk.MustNewDecFromStr("39.38"),
	}
}

func (s *IntegrationTestSuite) TestOracle_TokenPrice() {
	p, err := s.app.LeverageKeeper.TokenPrice(s.ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdk.MustNewDecFromStr("0.00000421"), p)

	p, err = s.app.LeverageKeeper.TokenPrice(s.ctx, atomIBCDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdk.MustNewDecFromStr("0.00003938"), p)

	p, err = s.app.LeverageKeeper.TokenPrice(s.ctx, "foo")
	s.Require().Error(err)
	s.Require().Equal(sdk.ZeroDec(), p)
}

func (s *IntegrationTestSuite) TestOracle_TokenValue() {
	// 2.4umee * $4.21
	v, err := s.app.LeverageKeeper.TokenValue(s.ctx, sdk.NewInt64Coin(umeeapp.BondDenom, 2400000))
	s.Require().NoError(err)
	s.Require().Equal(sdk.MustNewDecFromStr("10.104"), v)

	v, err = s.app.LeverageKeeper.TokenValue(s.ctx, sdk.NewInt64Coin("foo", 2400000))
	s.Require().Error(err)
	s.Require().Equal(sdk.ZeroDec(), v)
}

func (s *IntegrationTestSuite) TestOracle_TotalTokenValue() {
	// (2.4umee * $4.21) + (4.7atom * $39.38)
	v, err := s.app.LeverageKeeper.TotalTokenValue(
		s.ctx,
		sdk.NewCoins(
			sdk.NewInt64Coin(umeeapp.BondDenom, 2400000),
			sdk.NewInt64Coin(atomIBCDenom, 4700000),
		),
	)
	s.Require().NoError(err)
	s.Require().Equal(sdk.MustNewDecFromStr("195.19"), v)

	v, err = s.app.LeverageKeeper.TotalTokenValue(
		s.ctx,
		sdk.NewCoins(
			sdk.NewInt64Coin(umeeapp.BondDenom, 2400000),
			sdk.NewInt64Coin(atomIBCDenom, 4700000),
			sdk.NewInt64Coin("foo", 4700000),
		),
	)
	s.Require().Error(err)
	s.Require().Equal(sdk.ZeroDec(), v)
}

func (s *IntegrationTestSuite) TestOracle_EquivalentTokenValue() {
	c, err := s.app.LeverageKeeper.EquivalentTokenValue(s.ctx, sdk.NewInt64Coin(umeeapp.BondDenom, 2400000), atomIBCDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt64Coin(atomIBCDenom, 256576), c)

	c, err = s.app.LeverageKeeper.EquivalentTokenValue(s.ctx, sdk.NewInt64Coin("foo", 2400000), atomIBCDenom)
	s.Require().Error(err)
	s.Require().Equal(sdk.Coin{}, c)

	c, err = s.app.LeverageKeeper.EquivalentTokenValue(s.ctx, sdk.NewInt64Coin(umeeapp.BondDenom, 2400000), "foo")
	s.Require().Error(err)
	s.Require().Equal(sdk.Coin{}, c)
}
