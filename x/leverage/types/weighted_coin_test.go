package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/util/coin"
)

// referenceCoins aare a pre-sorted WeightedDecCoins with some equal weights and no repeated denoms
var referenceCoins = WeightedDecCoins{
	{
		Asset:  coin.Dec("VVVV", "1.0"),
		Weight: sdk.MustNewDecFromStr("1.0"),
	},
	{
		Asset:  coin.Dec("WWWW", "1.0"),
		Weight: sdk.MustNewDecFromStr("1.0"),
	},
	{
		Asset:  coin.Dec("DDDD", "1.0"),
		Weight: sdk.MustNewDecFromStr("0.4"),
	},
	{
		Asset:  coin.Dec("CCCC", "1.0"),
		Weight: sdk.MustNewDecFromStr("0.3"),
	},
	{
		Asset:  coin.Dec("BBBB", "1.0"),
		Weight: sdk.MustNewDecFromStr("0.2"),
	},
	{
		Asset:  coin.Dec("XXXX", "1.0"),
		Weight: sdk.MustNewDecFromStr("0.2"),
	},
	{
		Asset:  coin.Dec("AAAA", "1.0"),
		Weight: sdk.MustNewDecFromStr("0.1"),
	},
	{
		Asset:  coin.Dec("YYYY", "1.0"),
		Weight: sdk.MustNewDecFromStr("0.0"),
	},
	{
		Asset:  coin.Dec("ZZZZ", "1.0"),
		Weight: sdk.MustNewDecFromStr("0.0"),
	},
}

func TestWeightedDecCoinSorting(t *testing.T) {
	testCases := []struct {
		denom             string
		weight            sdk.Dec
		sortedBeforeIndex int // first index in referenceCoins which this coin should be sorted before
	}{
		{
			denom:             "ZZZZ",                       // would come before reference index zero
			weight:            sdk.MustNewDecFromStr("1.5"), // edge case (> 1)
			sortedBeforeIndex: 0,                            // sorted before all
		},
		{
			denom:             "AAAA", // would come before reference index zero
			weight:            sdk.MustNewDecFromStr("1.0"),
			sortedBeforeIndex: 0, // sorted before all
		},
		{
			denom:             "VVVV", // matches reference index zero
			weight:            sdk.MustNewDecFromStr("1.0"),
			sortedBeforeIndex: 1, // sorted before all except its match at index zero
		},
		{
			denom:             "ZZZZ", // would come before reference index zero
			weight:            sdk.MustNewDecFromStr("1.0"),
			sortedBeforeIndex: 2, // sorted before all except 0,1 due to alphabetical
		},
		{
			denom:             "AAAA",
			weight:            sdk.MustNewDecFromStr("0.35"),
			sortedBeforeIndex: 3, // sorted before reference coin C
		},
		{
			denom:             "ZZZZ",
			weight:            sdk.MustNewDecFromStr("0.0"),
			sortedBeforeIndex: len(referenceCoins), // sorted after all
		},
		{
			denom:             "AAAA",
			weight:            sdk.MustNewDecFromStr("-0.1"),
			sortedBeforeIndex: len(referenceCoins), // sorted after all
		},
	}

	for i, wdc := range referenceCoins {
		for j, c := range referenceCoins {
			assert.Equal(t, i < j, wdc.before(c), "require pre-sorted referenceCoins ", i, j)
		}
	}

	for _, tc := range testCases {
		c := WeightedDecCoin{
			Asset:  coin.Dec(tc.denom, "1.0"),
			Weight: tc.weight,
		}
		for i, wdc := range referenceCoins {
			assert.Equal(
				t,
				i >= tc.sortedBeforeIndex,
				c.before(wdc),
				"test coin sorts before reference index ", c, i,
			)
		}
	}
}
