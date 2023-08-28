package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/util/coin"
)

func weightedDecCoin(denom, amount, weight string) WeightedDecCoin {
	return WeightedDecCoin{
		Asset:  coin.Dec(denom, amount),
		Weight: sdk.MustNewDecFromStr(weight),
	}
}

func TestWeightedDecCoinSorting(t *testing.T) {
	// referenceCoins are a pre-sorted WeightedDecCoins with some equal weights and no repeated denoms
	referenceCoins := WeightedDecCoins{
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

func TestWeightedDecCoinsSub(t *testing.T) {
	testCases := []struct {
		initial WeightedDecCoins
		sub     WeightedDecCoin
		diff    WeightedDecCoins
		message string
	}{
		{
			WeightedDecCoins{
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("BBBB", "2.0", "0.1"),
				weightedDecCoin("CCCC", "3.0", "0.1"),
			},
			weightedDecCoin("CCCC", "3.0", "0.1"),
			WeightedDecCoins{
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("BBBB", "2.0", "0.1"),
				weightedDecCoin("CCCC", "0.0", "0.1"),
			},
			"sub equal weight assets",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("BBBB", "4.0", "0.2"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
			},
			weightedDecCoin("BBBB", "2.0", "0.2"),
			WeightedDecCoins{
				weightedDecCoin("BBBB", "2.0", "0.2"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
			},
			"partial sub asset",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("BBBB", "2.0", "0.2"),
				weightedDecCoin("CCCC", "3.0", "0.3"),
			},
			weightedDecCoin("CCCC", "3.0", "0.3"),
			WeightedDecCoins{
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("BBBB", "2.0", "0.2"),
				weightedDecCoin("CCCC", "0.0", "0.3"),
			},
			// note that this Sub function is used during sorting
			// operations which rely on coin index - no denom's
			// index should change as a result
			"does not fix unsorted input",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("BBBB", "0.0", "0.2"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("CCCC", "3.0", "0.3"),
			},
			weightedDecCoin("CCCC", "3.0", "0.3"),
			WeightedDecCoins{
				weightedDecCoin("BBBB", "0.0", "0.2"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("CCCC", "0.0", "0.3"),
			},
			// input denom indexes cannot change (even by being removed)
			"does not omit existing zero input",
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
			"survives zero input",
		},
		{
			WeightedDecCoins{
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("CCCC", "3.0", "0.3"),
			},
			weightedDecCoin("CCCC", "3.0", "0.3"),
			WeightedDecCoins{
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("AAAA", "1.0", "0.1"),
				weightedDecCoin("CCCC", "0.0", "0.3"),
			},
			// input denom indexes cannot change (even by fix duplicate)
			"does not fix duplicate input",
		},
	}

	for _, tc := range testCases {
		diff := tc.initial.Sub(tc.sub.Asset)
		assert.Equal(t, len(tc.diff), len(diff), tc.message)
		for i, wc := range tc.diff {
			assert.Equal(t, wc.String(), diff[i].String(), tc.message)
		}
	}
}

func TestWeightedNormalPairBefore(t *testing.T) {
	// referencePairs are a pre-sorted WeightedNormalPairs with some equal weights and repeated denoms
	referencePairs := WeightedNormalPairs{
		// this section of V & W assets confirms alphabetical sorting of equal-weight pairs
		{
			Collateral: weightedDecCoin("VVVV", "1.0", "1.0"),
			Borrow:     weightedDecCoin("VVVV", "1.0", "1.0"),
		},
		{
			Collateral: weightedDecCoin("VVVV", "1.0", "1.0"),
			Borrow:     weightedDecCoin("WWWW", "1.0", "1.0"),
		},
		{
			Collateral: weightedDecCoin("WWWW", "1.0", "1.0"),
			Borrow:     weightedDecCoin("VVVV", "1.0", "1.0"),
		},
		{
			Collateral: weightedDecCoin("WWWW", "1.0", "1.0"),
			Borrow:     weightedDecCoin("WWWW", "1.0", "1.0"),
		},
		// this V -> A pair confirms that weight of the borrow (0.1)
		// take precedence over alphabetical of the collateral (V-W)
		// when weight of the collateral is equal (1.0)
		{
			Collateral: weightedDecCoin("VVVV", "1.0", "1.0"),
			Borrow:     weightedDecCoin("AAAA", "1.0", "0.1"),
		},
		// this section of ABCD assets confirms weight sorting of pairs,
		// which prioritizes collateral weight and breaks ties
		// with borrow weight. this must work even if borrow weight
		// is out of order (which should not happen in practice)
		{
			Collateral: weightedDecCoin("DDDD", "1.0", "0.4"),
			Borrow:     weightedDecCoin("BBBB", "1.0", "0.2"),
		},
		{
			Collateral: weightedDecCoin("DDDD", "1.0", "0.4"),
			Borrow:     weightedDecCoin("AAAA", "1.0", "0.1"),
		},
		{
			Collateral: weightedDecCoin("CCCC", "1.0", "0.3"),
			Borrow:     weightedDecCoin("CCCC", "1.0", "0.3"),
		},
		{
			Collateral: weightedDecCoin("BBBB", "1.0", "0.2"),
			Borrow:     weightedDecCoin("CCCC", "1.0", "0.3"),
		},
		// this zero weight collateral should always be sorted last
		// regardless of what borrow it is paired with
		{
			Collateral: weightedDecCoin("ZZZZ", "1.0", "0.0"),
			Borrow:     weightedDecCoin("VVVV", "1.0", "1.0"),
		},
	}

	// check before() using referencePairs
	for i, wnp := range referencePairs {
		for j, c := range referencePairs {
			assert.Equal(t, i < j, wnp.before(c), "require pre-sorted referencePairs ", i, j)
		}
	}
}

func TestWeightedSpecialPairBefore(t *testing.T) {
	// referencePairs are a pre-sorted WeightedSpecialPairs with some equal weights and repeated denoms
	referencePairs := WeightedSpecialPairs{
		// this section of V & W assets confirms alphabetical sorting of equal-weight pairs
		// completely disregarding amount
		{
			Collateral:    coin.ZeroDec("VVVV"),
			Borrow:        coin.ZeroDec("VVVV"),
			SpecialWeight: sdk.MustNewDecFromStr("1.0"),
		},
		{
			Collateral:    coin.ZeroDec("VVVV"),
			Borrow:        coin.ZeroDec("WWWW"),
			SpecialWeight: sdk.MustNewDecFromStr("1.0"),
		},
		{
			Collateral:    coin.Dec("WWWW", "30.00"),
			Borrow:        coin.Dec("VVVV", "30.00"),
			SpecialWeight: sdk.MustNewDecFromStr("1.0"),
		},
		{
			Collateral:    coin.Dec("WWWW", "60.00"),
			Borrow:        coin.Dec("WWWW", "60.00"),
			SpecialWeight: sdk.MustNewDecFromStr("1.0"),
		},
		// this Y -> W pair confirms that weight of the pair
		// take precedence over alphabetical
		{
			Collateral:    coin.ZeroDec("YYYY"),
			Borrow:        coin.ZeroDec("WWWW"),
			SpecialWeight: sdk.MustNewDecFromStr("0.6"),
		},
		// this section ensures regular sorting by collateral weight
		{
			Collateral:    coin.ZeroDec("DDDD"),
			Borrow:        coin.ZeroDec("CCCC"),
			SpecialWeight: sdk.MustNewDecFromStr("0.4"),
		},
		{
			Collateral:    coin.ZeroDec("CCCC"),
			Borrow:        coin.ZeroDec("BBBB"),
			SpecialWeight: sdk.MustNewDecFromStr("0.3"),
		},
		{
			Collateral:    coin.Dec("BBBB", "100.00"),
			Borrow:        coin.Dec("AAAA", "20.00"),
			SpecialWeight: sdk.MustNewDecFromStr("0.2"),
		},
		{
			Collateral:    coin.ZeroDec("AAAA"),
			Borrow:        coin.ZeroDec("DDDD"),
			SpecialWeight: sdk.MustNewDecFromStr("0.1"),
		},
		// these zero-weight pairs should always be last,
		// regardless of coin amounts
		{
			Collateral:    coin.ZeroDec("YYYY"),
			Borrow:        coin.ZeroDec("YYYY"),
			SpecialWeight: sdk.MustNewDecFromStr("0.0"),
		},
		{
			Collateral:    coin.Dec("ZZZZ", "100.00"),
			Borrow:        coin.Dec("ZZZZ", "0.00"),
			SpecialWeight: sdk.MustNewDecFromStr("0.0"),
		},
	}

	// check before() using referencePairs
	for i, wsp := range referencePairs {
		for j, c := range referencePairs {
			assert.Equal(t, i < j, wsp.before(c), "require pre-sorted referencePairs ", i, j)
		}
	}
}

func TestWeightedSpecialPairsAdd(t *testing.T) {
	testCases := []struct {
		initial WeightedSpecialPairs
		add     WeightedSpecialPair
		sum     WeightedSpecialPairs
		message string
	}{
		// TODO: cases
	}

	for _, tc := range testCases {
		sum := tc.initial.Add(tc.add)
		assert.Equal(t, len(tc.sum), len(sum), tc.message)
		for i, wsp := range tc.sum {
			assert.Equal(t, wsp.String(), sum[i].String(), tc.message)
		}
	}
}

func TestWeightedNormalPairsAdd(t *testing.T) {
	testCases := []struct {
		initial WeightedNormalPairs
		add     WeightedNormalPair
		sum     WeightedNormalPairs
		message string
	}{
		// TODO: cases
	}

	for _, tc := range testCases {
		sum := tc.initial.Add(tc.add)
		assert.Equal(t, len(tc.sum), len(sum), tc.message)
		for i, wnp := range tc.sum {
			assert.Equal(t, wnp.String(), sum[i].String(), tc.message)
		}
	}
}

func TestWeightedSpecialPairsCanCombine(t *testing.T) {
	testCases := []struct {
		pairs      WeightedSpecialPairs
		canCombine bool
		message    string
	}{
		{
			[]WeightedSpecialPair{
				{
					Collateral:    coin.ZeroDec("AAAA"),
					Borrow:        coin.ZeroDec("BBBB"),
					SpecialWeight: sdk.MustNewDecFromStr("0.7"),
				},
				{
					Collateral:    coin.Dec("AAAA", "0.1"),
					Borrow:        coin.Dec("BBBB", "100.0"),
					SpecialWeight: sdk.MustNewDecFromStr("0.6"),
				},
				{
					Collateral:    coin.Dec("AAAA", "20.0"),
					Borrow:        coin.ZeroDec("BBBB"),
					SpecialWeight: sdk.MustNewDecFromStr("0.5"),
				},
			},
			true,
			"AB pairs, disregarding differing weights",
		},
		{
			[]WeightedSpecialPair{
				{
					Collateral:    coin.ZeroDec("AAAA"),
					Borrow:        coin.ZeroDec("AAAA"),
					SpecialWeight: sdk.MustNewDecFromStr("0.7"),
				},
				{
					Collateral:    coin.Dec("AAAA", "0.1"),
					Borrow:        coin.Dec("AAAA", "100.0"),
					SpecialWeight: sdk.MustNewDecFromStr("0.6"),
				},
				{
					Collateral:    coin.Dec("AAAA", "20.0"),
					Borrow:        coin.ZeroDec("AAAA"),
					SpecialWeight: sdk.MustNewDecFromStr("0.5"),
				},
			},
			true,
			"AA pairs, disregarding differing weights",
		},
		{
			[]WeightedSpecialPair{
				{
					Collateral:    coin.ZeroDec("AAAA"),
					Borrow:        coin.ZeroDec("AAAA"),
					SpecialWeight: sdk.MustNewDecFromStr("0.7"),
				},
				{
					Collateral:    coin.Dec("AAAA", "0.1"),
					Borrow:        coin.Dec("BBBB", "100.0"),
					SpecialWeight: sdk.MustNewDecFromStr("0.6"),
				},
				{
					Collateral:    coin.Dec("BBBB", "20.0"),
					Borrow:        coin.ZeroDec("AAAA"),
					SpecialWeight: sdk.MustNewDecFromStr("0.5"),
				},
				{
					Collateral:    coin.ZeroDec("BBBB"),
					Borrow:        coin.ZeroDec("BBBB"),
					SpecialWeight: sdk.MustNewDecFromStr("0.7"),
				},
			},
			false,
			"unique pairs",
		},
	}

	// each test case is constructed so that its pairs must all be combinable,
	// or all be unique.
	for _, tc := range testCases {
		for i, a := range tc.pairs {
			for j, b := range tc.pairs {
				// Test every possible relation within the set, except
				// elements with themselves in the unique (cannot combine) case.
				if i != j || tc.canCombine {
					assert.Equal(t, tc.canCombine, a.canCombine(b), tc.message, a, b)
				}
			}
		}
	}
}

func TestWeightedNormalPairsCanCombine(t *testing.T) {
	testCases := []struct {
		pairs      WeightedNormalPairs
		canCombine bool
		message    string
	}{
		{
			[]WeightedNormalPair{
				{
					Collateral: weightedDecCoin("AAAA", "10.0", "0.1"),
					Borrow:     weightedDecCoin("BBBB", "1.0", "0.2"),
				},
				{
					Collateral: weightedDecCoin("AAAA", "10.0", "0.1"),
					Borrow:     weightedDecCoin("BBBB", "1.0", "0.3"),
				},
				{
					Collateral: weightedDecCoin("AAAA", "0", "0.6"),
					Borrow:     weightedDecCoin("BBBB", "0", "0.7"),
				},
			},
			true,
			"AB pairs, disregarding differing weights",
		},
		{
			[]WeightedNormalPair{
				{
					Collateral: weightedDecCoin("AAAA", "10.0", "0.1"),
					Borrow:     weightedDecCoin("AAAA", "1.0", "0.2"),
				},
				{
					Collateral: weightedDecCoin("AAAA", "10.0", "0.1"),
					Borrow:     weightedDecCoin("AAAA", "1.0", "0.3"),
				},
				{
					Collateral: weightedDecCoin("AAAA", "0", "0.6"),
					Borrow:     weightedDecCoin("AAAA", "0", "0.7"),
				},
			},
			true,
			"AA pairs, disregarding differing weights",
		},

		{
			[]WeightedNormalPair{
				{
					Collateral: weightedDecCoin("AAAA", "10.0", "0.1"),
					Borrow:     weightedDecCoin("AAAA", "1.0", "0.2"),
				},
				{
					Collateral: weightedDecCoin("AAAA", "10.0", "0.1"),
					Borrow:     weightedDecCoin("BBBB", "1.0", "0.3"),
				},
				{
					Collateral: weightedDecCoin("BBBB", "0", "0.6"),
					Borrow:     weightedDecCoin("AAAA", "0", "0.7"),
				},
				{
					Collateral: weightedDecCoin("BBBB", "0", "0.6"),
					Borrow:     weightedDecCoin("BBBB", "0", "0.7"),
				},
			},
			false,
			"unique pairs",
		},
	}

	// each test case is constructed so that its pairs must all be combinable,
	// or all be unique.
	for _, tc := range testCases {
		for i, a := range tc.pairs {
			for j, b := range tc.pairs {
				// Test every possible relation within the set, except
				// elements with themselves in the unique (cannot combine) case.
				if i != j || tc.canCombine {
					assert.Equal(t, tc.canCombine, a.canCombine(b), tc.message, a, b)
				}
			}
		}
	}
}
