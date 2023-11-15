package types

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/util/coin"
)

// speicalPair creates a WeightedSpecialPair with given components, e.g, "10 uumee, 3.5 uatom, 0.35"
func specialPair(collateral, borrow, weight string) WeightedSpecialPair {
	c := strings.Split(collateral, " ")
	b := strings.Split(borrow, " ")
	return WeightedSpecialPair{
		Collateral:    coin.Dec(c[1], c[0]),
		Borrow:        coin.Dec(b[1], b[0]),
		SpecialWeight: sdk.MustNewDecFromStr(weight),
	}
}

func TestWeightedSpecialPairBefore(t *testing.T) {
	// referencePairs are a pre-sorted WeightedSpecialPairs with some equal weights and repeated denoms
	referencePairs := WeightedSpecialPairs{
		// this section of V & W assets confirms alphabetical sorting of equal-weight pairs
		// completely disregarding amount
		specialPair("0 VVVV", "0 VVVV", "1.0"),
		specialPair("0 VVVV", "0 WWWW", "1.0"),
		specialPair("30 WWWW", "30 VVVV", "1.0"),
		specialPair("60 WWWW", "60 WWWW", "1.0"),
		// this Y -> W pair confirms that weight of the pair
		// take precedence over alphabetical
		specialPair("0 YYYY", "0 WWWW", "0.6"),
		// this section ensures regular sorting by collateral weight
		specialPair("0 DDDD", "0 CCCC", "0.4"),
		specialPair("0 CCCC", "0 BBBB", "0.3"),
		specialPair("100 BBBB", "20 AAAA", "0.2"),
		specialPair("0 AAAA", "0 DDDD", "0.1"),
		// these zero-weight pairs should always be last,
		// regardless of coin amounts
		specialPair("0 YYYY", "0 YYYY", "0.0"),
		specialPair("100 ZZZZ", "0 ZZZZ", "0.0"),
	}

	// check before() using referencePairs
	for i, wsp := range referencePairs {
		for j, c := range referencePairs {
			assert.Equal(t, i < j, wsp.before(c), "require pre-sorted referencePairs ", i, j)
		}
	}
}

func TestWeightedSpecialPairsAdd(t *testing.T) {
	var (
		// special pairs names for "Collateral, Borrow, Collateral Amount"
		// weights are the following: A->B 0.7   A->C 0.6   A->A 0.5
		ab0   = specialPair("0 AAAA", "0 BBBB", "0.7")
		ab50  = specialPair("50 AAAA", "35 BBBB", "0.7")
		ac0   = specialPair("0 AAAA", "0 CCCC", "0.6")
		ac50  = specialPair("50 AAAA", "30 CCCC", "0.6")
		ac100 = specialPair("100 AAAA", "60 CCCC", "0.6")
		ac150 = specialPair("150 AAAA", "90 CCCC", "0.6")
		aa20  = specialPair("20 AAAA", "10 AAAA", "0.5")
	)

	testCases := []struct {
		initial WeightedSpecialPairs
		add     WeightedSpecialPair
		sum     WeightedSpecialPairs
		message string
	}{
		{
			[]WeightedSpecialPair{ab0, ac100, aa20},
			ac50,
			[]WeightedSpecialPair{ab0, ac150, aa20},
			"existing pair addition with zero values present",
		},
		{
			[]WeightedSpecialPair{ab0, ac100, aa20},
			ac0,
			[]WeightedSpecialPair{ab0, ac100, aa20},
			"existing pair zero valued input",
		},
		{
			[]WeightedSpecialPair{ab0, aa20},
			ac50,
			[]WeightedSpecialPair{ab0, ac50, aa20},
			"new pair addition with zero values present",
		},
		{
			[]WeightedSpecialPair{ab0, aa20},
			ac0,
			[]WeightedSpecialPair{ab0, ac0, aa20},
			"new zero-valued pair addition with zero values present",
		},
		{
			[]WeightedSpecialPair{aa20, ab0},
			ac50,
			[]WeightedSpecialPair{ab0, ac50, aa20},
			"fixes unsorted input",
		},
		{
			[]WeightedSpecialPair{ab50},
			specialPair("50 AAAA", "30 BBBB", "0.6"), // AB weight is usually 0.7
			[]WeightedSpecialPair{
				specialPair("100 AAAA", "65 BBBB", "0.7"),
				// adds numbers but keeps original weight
				// which is now inaccurate (65 != 0.7 * 100)
			},
			"differing weights (should never happen) keeps original weight",
		},
	}

	for _, tc := range testCases {
		sum := tc.initial.Add(tc.add)
		assert.Equal(t, len(tc.sum), len(sum), tc.message)
		for i, wsp := range tc.sum {
			assert.Equal(t, wsp.String(), sum[i].String(), tc.message)
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
