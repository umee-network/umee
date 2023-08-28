package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/util/coin"
)

// referenceCoins are a pre-sorted WeightedDecCoins with some equal weights and no repeated denoms
var referenceCoins = WeightedDecCoins{
	weightedDecCoin("VVVV", "1.0", "1.0"),
	weightedDecCoin("WWWW", "2.0", "1.0"),
	weightedDecCoin("DDDD", "1.0", "0.4"),
	weightedDecCoin("CCCC", "2.0", "0.3"),
	weightedDecCoin("BBBB", "1.0", "0.2"),
	weightedDecCoin("XXXX", "2.0", "0.2"),
	weightedDecCoin("AAAA", "1.0", "0.1"),
	weightedDecCoin("YYYY", "2.0", "0.0"),
	weightedDecCoin("ZZZZ", "1.0", "0.0"),
}

func weightedDecCoin(denom, amount, weight string) WeightedDecCoin {
	return WeightedDecCoin{
		Asset:  coin.Dec(denom, amount),
		Weight: sdk.MustNewDecFromStr(weight),
	}
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
			denom:             "AAAA",
			weight:            sdk.MustNewDecFromStr("1.0"),
			sortedBeforeIndex: 0, // sorted before all
		},
		{
			denom:             "VVVV",
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
			weight:            sdk.MustNewDecFromStr("-0.1"), // edge case (< 0)
			sortedBeforeIndex: len(referenceCoins),           // sorted after all
		},
	}

	// check before() using referenceCoins
	for i, wdc := range referenceCoins {
		for j, c := range referenceCoins {
			assert.Equal(t, i < j, wdc.before(c), "require pre-sorted referenceCoins ", i, j)
		}
	}

	// check new coins, including matching coins and edge cases
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

func TestWeightedDecCoinTotal(t *testing.T) {
	testCases := []struct {
		weightedCoins WeightedDecCoins
		denom         string
		denomTotal    string
		total         string
		message       string
	}{
		{
			WeightedDecCoins{
				weightedDecCoin("AAAA", "0.1", "0.1"),
			},
			"AAAA",
			"0.1",
			"0.1",
			"single asset",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("AAAA", "0.1", "0.1"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
			},
			"AAAA",
			"1.1",
			"1.1",
			"duplicate asset",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("AAAA", "2.0", "0.1"),
				weightedDecCoin("BBBB", "1.0", "0.1"),
			},
			"BBBB",
			"1.0",
			"3.0",
			"different assets",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("AAAA", "0.0", "0.1"),
				weightedDecCoin("BBBB", "1.0", "0.1"),
				weightedDecCoin("CCCC", "2.0", "0.1"),
				weightedDecCoin("DDDD", "3.0", "0.1"),
				weightedDecCoin("EEEE", "4.0", "0.1"),
			},
			"AAAA",
			"0.0",
			"10.0",
			"multiple same-weight assets, including a zero",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("AAAA", "0.0", "0.4"),
				weightedDecCoin("BBBB", "1.0", "0.3"),
				weightedDecCoin("CCCC", "2.0", "0.2"),
				weightedDecCoin("DDDD", "3.0", "0.3"),
				weightedDecCoin("EEEE", "4.0", "0.0"),
			},
			"",
			"10.0",
			"10.0",
			"multiple weighted assets, including a zero",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("BBBB", "1.0", "0.3"),
				weightedDecCoin("AAAA", "0.0", "0.4"),
				weightedDecCoin("CCCC", "2.0", "0.2"),
				weightedDecCoin("EEEE", "4.0", "0.0"),
				weightedDecCoin("DDDD", "3.0", "0.3"),
			},
			"AAAA",
			"0.0",
			"10.0",
			"unsorted weighted assets, including a zero",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t,
			sdk.MustNewDecFromStr(tc.total).String(),
			tc.weightedCoins.Total("").String(),
			"full total"+tc.message,
		)
		assert.Equal(t,
			sdk.MustNewDecFromStr(tc.denomTotal).String(),
			tc.weightedCoins.Total(tc.denom).String(),
			tc.denom+" denom total "+tc.message,
		)
	}
}

func TestWeightedDecCoinsAdd(t *testing.T) {
	testCases := []struct {
		initial WeightedDecCoins
		add     WeightedDecCoin
		sum     WeightedDecCoins
		message string
	}{
		{
			WeightedDecCoins{
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("BBBB", "2.0", "0.1"),
			},
			weightedDecCoin("CCCC", "3.0", "0.1"),
			WeightedDecCoins{
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("BBBB", "2.0", "0.1"),
				weightedDecCoin("CCCC", "3.0", "0.1"),
			},
			"add equal weight assets",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("BBBB", "2.0", "0.2"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
			},
			weightedDecCoin("CCCC", "3.0", "0.3"),
			WeightedDecCoins{
				weightedDecCoin("CCCC", "3.0", "0.3"),
				weightedDecCoin("BBBB", "2.0", "0.2"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
			},
			"sorts by weight",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("BBBB", "2.0", "0.2"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
			},
			weightedDecCoin("BBBB", "2.0", "0.2"),
			WeightedDecCoins{
				weightedDecCoin("BBBB", "4.0", "0.2"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
			},
			"existing asset",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("BBBB", "2.0", "0.2"),
			},
			weightedDecCoin("CCCC", "3.0", "0.3"),
			WeightedDecCoins{
				weightedDecCoin("CCCC", "3.0", "0.3"),
				weightedDecCoin("BBBB", "2.0", "0.2"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
			},
			"fixes unsorted input",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("BBBB", "0.0", "0.2"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
			},
			weightedDecCoin("CCCC", "3.0", "0.3"),
			WeightedDecCoins{
				weightedDecCoin("CCCC", "3.0", "0.3"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
			},
			"omits existing zero input",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("BBBB", "2.0", "0.2"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
			},
			weightedDecCoin("CCCC", "0.0", "0.3"),
			WeightedDecCoins{
				weightedDecCoin("BBBB", "2.0", "0.2"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
			},
			"omits new zero input",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
			},
			weightedDecCoin("CCCC", "3.0", "0.3"),
			WeightedDecCoins{
				weightedDecCoin("CCCC", "3.0", "0.3"),
				weightedDecCoin("AAAA", "2.0", "0.1"),
			},
			"fixes duplicate input",
		},
	}

	for _, tc := range testCases {
		sum := tc.initial.Add(tc.add)
		assert.Equal(t, len(tc.sum), len(sum), tc.message)
		for i, wc := range tc.sum {
			assert.Equal(t, wc.String(), sum[i].String(), tc.message)
		}
	}
}

func TestWeightedDecCoinsSub(_ *testing.T) {
	// TODO
}

func TestWeightedNormalPairBefore(_ *testing.T) {
	// TODO
}

func TestWeightedSpecialPairBefore(_ *testing.T) {
	// TODO
}

func TestWeightedSpecialPairsAdd(_ *testing.T) {
	// TODO
}

func TestWeightedNormalPairsAdd(_ *testing.T) {
	// TODO
}

func TestWeightedSpecialPairsCanCombine(_ *testing.T) {
	// TODO
}

func TestWeightedNormalPairsCanCombine(_ *testing.T) {
	// TODO
}
