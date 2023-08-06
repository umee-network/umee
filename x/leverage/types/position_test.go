package types_test

import (
	"testing"

	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/leverage/types"
)

func testToken(denom, cw, lt string) types.Token {
	return types.Token{
		BaseDenom:            denom,
		CollateralWeight:     sdk.MustNewDecFromStr(cw),
		LiquidationThreshold: sdk.MustNewDecFromStr(lt),
	}
}

func testPair(borrow, collateral, cw, lt string) types.SpecialAssetPair {
	return types.SpecialAssetPair{
		Borrow:               borrow,
		Collateral:           collateral,
		CollateralWeight:     sdk.MustNewDecFromStr(cw),
		LiquidationThreshold: sdk.MustNewDecFromStr(lt),
	}
}

// These tokens are used for testing asset positions. They are arranged so that
// A < B < C < D < E < F < G for both collateral weight and liquidation threshold,
// but H = I = 0 for collateral weight and G < H < I for liquidation threshold.
// This should produce a wide range of behaviors.
var orderedTokens = []types.Token{
	testToken("AAAA", "0.1", "0.15"),
	testToken("BBBB", "0.2", "0.25"),
	testToken("CCCC", "0.3", "0.35"),
	testToken("DDDD", "0.4", "0.45"),
	testToken("EEEE", "0.5", "0.55"),
	testToken("FFFF", "0.6", "0.65"),
	testToken("GGGG", "0.7", "0.75"),
	testToken("HHHH", "0.0", "0.85"),
	testToken("IIII", "0.0", "0.95"),
}

// These special asset pairs are used for testing asset positions.
// The even-numbered assets (B,D,F,H) are involved in special pairs, and the others aren't.
// When combined with the order of the test assets, many complex positions can be formed.
var orderedPairs = []types.SpecialAssetPair{
	// F and H are paired at [0.6,0.8] but not looped
	// D can borrow any (B,D,F,H) at [0.5,0.5]
	// B can be borrowed by any (B,D,F,H) at [0.3,0.3]
	// H is looped at [0.1,0.1] even though its token weights are [0.0,0.85]
	testPair("FFFF", "HHHH", "0.6", "0.8"),
	testPair("HHHH", "FFFF", "0.6", "0.8"),

	testPair("DDDD", "BBBB", "0.5", "0.5"),
	testPair("DDDD", "DDDD", "0.5", "0.5"),
	testPair("DDDD", "FFFF", "0.5", "0.5"),
	testPair("DDDD", "HHHH", "0.5", "0.5"),

	testPair("BBBB", "BBBB", "0.3", "0.3"),
	testPair("BBBB", "DDDD", "0.3", "0.3"),
	testPair("BBBB", "FFFF", "0.3", "0.3"),
	testPair("BBBB", "HHHH", "0.3", "0.3"),

	testPair("HHHH", "HHHH", "0.1", "0.1"),
}

// TestBorrowLimit verifies the borrow limit and liquidation threshold of various positions created
// from given borrowed and collateral values after token weights and special pairs are applied.
func TestBorrowLimit(t *testing.T) {
	type testCase struct {
		collateral           sdk.DecCoins
		borrow               sdk.DecCoins
		borrowLimit          string
		liquidationthreshold string
		msg                  string
	}

	testCases := []testCase{
		{
			// single asset
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(),
			// collateral weight 0.1, liquidation threshold 0.15
			"10.00",
			"15.00",
			"simple A",
		},
		{
			// single asset, with borrow
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("CCCC", "6"),
			),
			// collateral weight 0.1, liquidation threshold 0.15
			"10.00",
			"15.00",
			"A -> C",
		},
		{
			// single asset, at borrow limit
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("CCCC", "10"),
			),
			// collateral weight 0.1, liquidation threshold 0.15
			"10.00",
			"15.00",
			"A -> C at borrow limit",
		},
		{
			// single asset, at liquidation threshold
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("CCCC", "15"),
			),
			// collateral weight 0.1, liquidation threshold 0.15
			"10.00",
			"15.00",
			"A -> C at liquidation threshold",
		},
		{
			// single asset, above liquidation threshold
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("CCCC", "25"),
			),
			// collateral weight 0.1, liquidation threshold 0.15
			"10.00",
			"15.00",
			"A -> C above liquidation threshold",
		},
		{
			// multiple assets, one with zero weight
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
				coin.Dec("GGGG", "100"),
				coin.Dec("IIII", "100"),
			),
			sdk.NewDecCoins(),
			// sum of multiple assets, each using its collateral weight and liquidation threshold
			"80.00",
			"185.00",
			"simple AGI",
		},
		{
			// multiple assets, one with zero weight, at borrow limit
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
				coin.Dec("GGGG", "100"),
				coin.Dec("IIII", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("GGGG", "80"),
			),
			// effectiveness of I collateral is reduced to due to G liquidation threshold, thus leading
			// to a lower liquidation threshold than "simple AGI" test case above
			"80.00",
			"165.00",
			"AGI -> G at borrow limit",
		},
		{
			// multiple assets, one with zero weight, at liquidation threshold
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
				coin.Dec("GGGG", "100"),
				coin.Dec("IIII", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("GGGG", "165"),
			),
			// significantly over borrow limit, so calculation subtracts value of surplus borrows
			// from total borrowed value to determine borrow limit
			"80.00",
			"165.00",
			"AGI -> G at liquidation threshold",
		},
		{
			// multiple assets, one with zero weight, above liquidation threshold
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
				coin.Dec("GGGG", "100"),
				coin.Dec("IIII", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("GGGG", "500"),
			),
			// significantly over borrow limit and liquidation threshold, but calculation still reaches
			// the same values for them
			"80.00",
			"165.00",
			"AGI -> G above liquidation threshold",
		},
		{
			// single asset unused with special pair (no borrows)
			sdk.NewDecCoins(
				coin.Dec("FFFF", "100"),
			),
			sdk.NewDecCoins(),
			// collateral weight 0.6, liquidation threshold 0.65
			// the F <-> H special pair has no effect
			"60.00",
			"65.00",
			"simple F",
		},
		{
			// single asset with unused special pair (normal borrow)
			sdk.NewDecCoins(
				coin.Dec("FFFF", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("FFFF", "30"),
			),
			// borrow limit is unaffected since F -> F does not benefit from special pairs or suffer from borrow factor
			// the F <-> H special pair has no effect
			"60.00",
			"65.00",
			"F loop",
		},
		{
			// single asset with unused special pair (borrowFactor reducing weight, minimumBorrowFactor active)
			sdk.NewDecCoins(
				coin.Dec("FFFF", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("AAAA", "40"),
			),
			// 40 A consumes 80 F collateral (weight 0.5 due to MinimumBorrowFactor), leaving 20 F collateral unused.
			// Total borrow limit and liquidation thresholds are 40 + [0.6 and 0.65] * 20
			// the F <-> H special pair has no effect
			"52.00",
			"53.00",
			"F -> A",
		},
		{
			// single asset with special pair in effect
			sdk.NewDecCoins(
				coin.Dec("FFFF", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("HHHH", "30"),
			),
			// 30 H consumes 50 F collateral (weight 0.6 due to Special Pair), leaving 50 F collateral unused.
			// Remaining borrow limit is 30 + 0.6 * 50 = 60.
			// Meanwhile, 30A consumes 37.5 F collateral (liquidation threshold 0.8 due to special pair),
			// leaving 62.5. Total liquidation threshold is 30 + 62.50 * 0.65
			"60.00",
			"70.625",
			"F -> H below borrow limit",
		},
		{
			// single asset with special pair in effect - exactly at borrow limit
			sdk.NewDecCoins(
				coin.Dec("FFFF", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("HHHH", "60"),
			),
			// 60 H consumes all 100 F collateral (weight 0.6 due to Special Pair). Borrow limit equals value.
			// Meanwhile, 60A consumes 75 F collateral (liquidation threshold 0.8 due to special pair),
			// leaving 25. Total liquidation threshold is 60 + 25 * 0.65
			"60.00",
			"76.25",
			"F -> H at borrow limit",
		},
		{
			// single asset with special pair in effect - exactly at liquidation threshold
			sdk.NewDecCoins(
				coin.Dec("FFFF", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("HHHH", "80"),
			),
			// 60 H consumes all 100 F collateral (weight 0.6 due to Special Pair).
			// A remaining 20H is surplus borrowed value. Borrow limit equals value minus surplus.
			// Meanwhile, 80A consumes 100 F collateral (liquidation threshold 0.8 due to special pair).
			// Liquidation threshold is exactly borrowed value.
			"60.00",
			"80.00",
			"F -> H at liquidation threshold",
		},
		{
			// single asset with special pair in effect - above liquidation threshold
			sdk.NewDecCoins(
				coin.Dec("FFFF", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("HHHH", "100"),
			),
			// 60 H consumes all 100 F collateral (weight 0.6 due to Special Pair).
			// A remaining 40H is surplus borrowed value. Borrow limit equals value minus surplus.
			// 80 H consumes all 100 F collateral (liquidation threshold 0.8 due to Special Pair).
			// A remaining 20H is surplus borrowed value. Liquidation threshold equals value minus surplus.
			"60.00",
			"80.00",
			"F -> H above liquidation threshold",
		},
	}

	for _, tc := range testCases {
		borrowPosition := types.NewAccountPosition(
			orderedTokens,
			orderedPairs,
			tc.collateral,
			tc.borrow,
			false,
		)
		if !sdk.MustNewDecFromStr(tc.borrowLimit).Equal(borrowPosition.Limit()) {
			assert.Equal(t, borrowPosition.String(), "borrow limit position "+tc.msg)
		}
		assert.Equal(t,
			sdk.MustNewDecFromStr(tc.borrowLimit).String(),
			borrowPosition.Limit().String(),
			tc.msg+" borrow limit",
		)
		liquidationPosition := types.NewAccountPosition(
			orderedTokens,
			orderedPairs,
			tc.collateral,
			tc.borrow,
			true,
		)
		assert.Equal(t,
			sdk.MustNewDecFromStr(tc.liquidationthreshold).String(),
			liquidationPosition.Limit().String(),
			tc.msg+" liquidation threshold\n\n"+liquidationPosition.String(),
		)
	}
}

func TestMaxBorrow(t *testing.T) {
	type testCase struct {
		collateral     sdk.DecCoins
		borrow         sdk.DecCoins
		maxBorrowDenom string
		maxBorrow      string
		msg            string
	}

	testCases := []testCase{
		{
			// single asset
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(),
			// collateral weight 0.1, should be able to borrow 10 A
			"AAAA",
			"10.00",
			"simple A max(A)",
		},
		{
			// single asset, with existing looped borrow
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("AAAA", "7"),
			),
			// collateral weight 0.1, should be able to borrow 10 total
			"AAAA",
			"3.00",
			"simple A->A max(A)",
		},
		{
			// single asset, with existing borrow
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("IIII", "4"),
			),
			// collateral weight 0.1, should be able to borrow 10 total
			"AAAA",
			"6.00",
			"simple A->I max(A)",
		},
	}

	for _, tc := range testCases {
		borrowPosition := types.NewAccountPosition(
			orderedTokens,
			orderedPairs,
			tc.collateral,
			tc.borrow,
			false,
		)
		assert.Equal(t,
			sdk.MustNewDecFromStr(tc.maxBorrow).String(),
			borrowPosition.MaxBorrow(tc.maxBorrowDenom).String(),
			tc.msg+" max borrow\n\n"+borrowPosition.String(),
		)
	}
}
