package oracle_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/price-feeder/oracle"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
)

func TestComputeVWAP(t *testing.T) {
	testCases := map[string]struct {
		prices   map[string]map[string]provider.TickerPrice
		expected map[string]sdk.Dec
	}{}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			vwap := oracle.ComputeVWAP(tc.prices)
			require.Len(t, vwap, len(tc.expected))

			for k, v := range tc.expected {
				require.Equal(t, v, vwap[k])
			}
		})
	}
}
