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

func (s *IntegrationTestSuite) TestOracle_TotalPrice() {
	totalPrice, err := s.leverageKeeper.TotalPrice(s.ctx, []string{umeeapp.BondDenom, atomIBCDenom})
	s.Require().NoError(err)
	s.Require().Equal(sdk.MustNewDecFromStr("0.00004359"), totalPrice)

	totalPrice, err = s.leverageKeeper.TotalPrice(s.ctx, []string{umeeapp.BondDenom, "foo"})
	s.Require().Error(err)
	s.Require().Equal(sdk.ZeroDec(), totalPrice)
}

func (s *IntegrationTestSuite) TestOracle_EquivalentValue() {
	c, err := s.leverageKeeper.EquivalentValue(s.ctx, sdk.NewInt64Coin(umeeapp.BondDenom, 2400000), atomIBCDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt64Coin(atomIBCDenom, 256576), c)
}

func (s *IntegrationTestSuite) TestOracle_Price() {
	p, err := s.leverageKeeper.Price(s.ctx, umeeapp.BondDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdk.MustNewDecFromStr("0.00000421"), p)

	p, err = s.leverageKeeper.Price(s.ctx, atomIBCDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdk.MustNewDecFromStr("0.00003938"), p)

	p, err = s.leverageKeeper.Price(s.ctx, "foo")
	s.Require().Error(err)
	s.Require().Equal(sdk.ZeroDec(), p)
}
