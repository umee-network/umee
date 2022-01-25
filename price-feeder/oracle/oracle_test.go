package oracle

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/umee-network/umee/price-feeder/config"
	"github.com/umee-network/umee/price-feeder/oracle/client"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
	"github.com/umee-network/umee/price-feeder/oracle/types"

	oracletypes "github.com/umee-network/umee/x/oracle/types"
)

type mockProvider struct {
	prices map[string]provider.TickerPrice
}

func (m mockProvider) GetTickerPrices(_ ...types.CurrencyPair) (map[string]provider.TickerPrice, error) {
	return m.prices, nil
}

type failingProvider struct {
	prices map[string]provider.TickerPrice
}

func (m failingProvider) GetTickerPrices(_ ...types.CurrencyPair) (map[string]provider.TickerPrice, error) {
	return nil, fmt.Errorf("unable to get ticker prices")
}

type OracleTestSuite struct {
	suite.Suite

	oracle *Oracle
}

// SetupSuite executes once before the suite's tests are executed.
func (ots *OracleTestSuite) SetupSuite() {
	ots.oracle = New(
		zerolog.Nop(),
		client.OracleClient{},
		[]config.CurrencyPair{
			{
				Base:      "UMEE",
				Quote:     "USDT",
				Providers: []string{config.ProviderBinance},
			},
			{
				Base:      "UMEE",
				Quote:     "USDC",
				Providers: []string{config.ProviderKraken},
			},
			{
				Base:      "XBT",
				Quote:     "USDT",
				Providers: []string{config.ProviderOsmosis},
			},
		},
	)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OracleTestSuite))
}

func (ots *OracleTestSuite) TestStop() {
	ots.Eventually(
		func() bool {
			ots.oracle.Stop()
			return true
		},
		5*time.Second,
		time.Second,
	)
}

func (ots *OracleTestSuite) TestGetLastPriceSyncTimestamp() {
	// when no tick() has been invoked, assume zero value
	ots.Require().Equal(time.Time{}, ots.oracle.GetLastPriceSyncTimestamp())
}

func (ots *OracleTestSuite) TestPrices() {
	// initial prices should be empty (not set)
	ots.Require().Empty(ots.oracle.GetPrices())

	// Use a mock provider with exchange rates that are not specified in
	// configuration.
	ots.oracle.priceProviders = map[string]provider.Provider{
		config.ProviderBinance: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDX": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		config.ProviderKraken: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDX": {
					Price:  sdk.MustNewDecFromStr("3.70"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}
	acceptList := oracletypes.DenomList{
		oracletypes.Denom{
			BaseDenom:   "UUMEE",
			SymbolDenom: "UMEE",
			Exponent:    6,
		},
	}

	ots.Require().Error(ots.oracle.SetPrices(acceptList))
	ots.Require().Empty(ots.oracle.GetPrices())

	// use a mock provider to provide prices for the configured exchange pairs
	ots.oracle.priceProviders = map[string]provider.Provider{
		config.ProviderBinance: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDT": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		config.ProviderKraken: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDC": {
					Price:  sdk.MustNewDecFromStr("3.70"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}

	ots.Require().NoError(ots.oracle.SetPrices(acceptList))

	prices := ots.oracle.GetPrices()
	ots.Require().Len(prices, 1)
	ots.Require().Equal(sdk.MustNewDecFromStr("3.710916056220858266"), prices["UMEE"])

	// use one working provider and one provider with an incorrect exchange rate
	ots.oracle.priceProviders = map[string]provider.Provider{
		config.ProviderBinance: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDX": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		config.ProviderKraken: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDC": {
					Price:  sdk.MustNewDecFromStr("3.70"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}

	ots.Require().NoError(ots.oracle.SetPrices(acceptList))
	prices = ots.oracle.GetPrices()
	ots.Require().Len(prices, 1)
	ots.Require().Equal(sdk.MustNewDecFromStr("3.70"), prices["UMEE"])

	// use one working provider and one provider that fails
	ots.oracle.priceProviders = map[string]provider.Provider{
		config.ProviderBinance: failingProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDC": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		config.ProviderKraken: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDC": {
					Price:  sdk.MustNewDecFromStr("3.71"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}

	ots.Require().NoError(ots.oracle.SetPrices(acceptList))
	prices = ots.oracle.GetPrices()
	ots.Require().Len(prices, 1)
	ots.Require().Equal(sdk.MustNewDecFromStr("3.71"), prices["UMEE"])

	// use one working provider and one for a coin not in accept list
	ots.oracle.priceProviders = map[string]provider.Provider{
		config.ProviderKraken: mockProvider{
			prices: map[string]provider.TickerPrice{
				"UMEEUSDC": {
					Price:  sdk.MustNewDecFromStr("3.71"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
		config.ProviderOsmosis: mockProvider{
			prices: map[string]provider.TickerPrice{
				"XBTUSDT": {
					Price:  sdk.MustNewDecFromStr("3.71"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}

	ots.Require().NoError(ots.oracle.SetPrices(acceptList))
	prices = ots.oracle.GetPrices()
	ots.Require().Len(prices, 1)
	ots.Require().Equal(sdk.MustNewDecFromStr("3.71"), prices["UMEE"])
	_, reportedUnacceptedDenom := prices["XBT"]
	ots.Require().False(reportedUnacceptedDenom)
}

func TestGenerateSalt(t *testing.T) {
	salt, err := GenerateSalt(0)
	require.Error(t, err)
	require.Empty(t, salt)

	salt, err = GenerateSalt(5)
	require.NoError(t, err)
	require.NotEmpty(t, salt)
}

func TestGenerateExchangeRatesString(t *testing.T) {
	testCases := map[string]struct {
		input    map[string]sdk.Dec
		expected string
	}{
		"empty input": {
			input:    make(map[string]sdk.Dec),
			expected: "",
		},
		"single denom": {
			input: map[string]sdk.Dec{
				"UMEE": sdk.MustNewDecFromStr("3.72"),
			},
			expected: "UMEE:3.720000000000000000",
		},
		"multi denom": {
			input: map[string]sdk.Dec{
				"UMEE": sdk.MustNewDecFromStr("3.72"),
				"ATOM": sdk.MustNewDecFromStr("40.13"),
				"OSMO": sdk.MustNewDecFromStr("8.69"),
			},
			expected: "ATOM:40.130000000000000000,OSMO:8.690000000000000000,UMEE:3.720000000000000000",
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			out := GenerateExchangeRatesString(tc.input)
			require.Equal(t, tc.expected, out)
		})
	}
}
