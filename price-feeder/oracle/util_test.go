package oracle_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/config"
	"github.com/umee-network/umee/price-feeder/oracle"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
)

func TestComputeVWAP(t *testing.T) {
	testCases := map[string]struct {
		prices   map[string]map[string]provider.TickerPrice
		expected map[string]sdk.Dec
	}{
		"empty prices": {
			prices:   make(map[string]map[string]provider.TickerPrice),
			expected: make(map[string]sdk.Dec),
		},
		"nil prices": {
			prices:   nil,
			expected: make(map[string]sdk.Dec),
		},
		"non empty prices": {
			prices: map[string]map[string]provider.TickerPrice{
				config.ProviderBinance: {
					"ATOM": provider.TickerPrice{
						Price:  sdk.MustNewDecFromStr("28.21000000"),
						Volume: sdk.MustNewDecFromStr("2749102.78000000"),
					},
					"UMEE": provider.TickerPrice{
						Price:  sdk.MustNewDecFromStr("1.13000000"),
						Volume: sdk.MustNewDecFromStr("249102.38000000"),
					},
					"LUNA": provider.TickerPrice{
						Price:  sdk.MustNewDecFromStr("64.87000000"),
						Volume: sdk.MustNewDecFromStr("7854934.69000000"),
					},
				},
				config.ProviderKraken: {
					"ATOM": provider.TickerPrice{
						Price:  sdk.MustNewDecFromStr("28.268700"),
						Volume: sdk.MustNewDecFromStr("178277.53314385"),
					},
					"LUNA": provider.TickerPrice{
						Price:  sdk.MustNewDecFromStr("64.87853000"),
						Volume: sdk.MustNewDecFromStr("458917.46353577"),
					},
				},
				"FOO": {
					"ATOM": provider.TickerPrice{
						Price:  sdk.MustNewDecFromStr("28.168700"),
						Volume: sdk.MustNewDecFromStr("4749102.53314385"),
					},
				},
			},
			expected: map[string]sdk.Dec{
				"ATOM": sdk.MustNewDecFromStr("28.185812745610043621"),
				"UMEE": sdk.MustNewDecFromStr("1.13000000"),
				"LUNA": sdk.MustNewDecFromStr("64.870470848638112395"),
			},
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			vwap, err := oracle.ComputeVWAP(tc.prices)
			require.NoError(t, err)
			require.Len(t, vwap, len(tc.expected))

			for k, v := range tc.expected {
				require.Equalf(t, v, vwap[k], "unexpected VWAP for %s", k)
			}
		})
	}
}
