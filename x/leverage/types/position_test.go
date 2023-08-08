package types_test

import (
	"testing"

	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/leverage/types"
)

var (
	noMinimumBorrowFactor   = sdk.MustNewDecFromStr("0.01")
	highMinimumBorrowFactor = sdk.MustNewDecFromStr("0.5")
)

func testToken(denom, cw, lt string) types.Token {
	return types.Token{
		BaseDenom:            denom,
		CollateralWeight:     sdk.MustNewDecFromStr(cw),
		LiquidationThreshold: sdk.MustNewDecFromStr(lt),
	}
}

func testPair(collateral, borrow, cw, lt string) types.SpecialAssetPair {
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
			// significantly over borrow limit, so calculation subtracts value of unpaired borrows
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
			// A remaining 20H is unpaired borrowed value. Borrow limit equals value minus unpaired.
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
			// A remaining 40H is unpaired borrowed value. Borrow limit equals value minus unpaired.
			// 80 H consumes all 100 F collateral (liquidation threshold 0.8 due to Special Pair).
			// A remaining 20H is unpaired borrowed value. Liquidation threshold equals value minus unpaired.
			"60.00",
			"80.00",
			"F -> H above liquidation threshold",
		},
	}

	for _, tc := range testCases {
		borrowPosition, err := types.NewAccountPosition(
			orderedTokens,
			orderedPairs,
			tc.collateral,
			tc.borrow,
			false,
			highMinimumBorrowFactor,
		)
		assert.NilError(t, err, tc.msg+" borrow limit\n\n"+borrowPosition.String())
		assert.Equal(t,
			sdk.MustNewDecFromStr(tc.borrowLimit).String(),
			borrowPosition.Limit().String(),
			tc.msg+" borrow limit\n\n"+borrowPosition.String(),
		)
		liquidationPosition, err := types.NewAccountPosition(
			orderedTokens,
			orderedPairs,
			tc.collateral,
			tc.borrow,
			true,
			highMinimumBorrowFactor,
		)
		assert.NilError(t, err, tc.msg+" liquidation threshold\n\n"+liquidationPosition.String())
		assert.Equal(t,
			sdk.MustNewDecFromStr(tc.liquidationthreshold).String(),
			liquidationPosition.Limit().String(),
			tc.msg+" liquidation threshold\n\n"+liquidationPosition.String(),
		)
	}
}

func TestMaxBorrowNoSpecialPairs(t *testing.T) {
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
		{
			// single asset, with multiple existing borrows, borrowing lowest-weighted asset
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("AAAA", "1"),
				coin.Dec("CCCC", "1"),
				coin.Dec("EEEE", "1"),
				coin.Dec("IIII", "1"),
			),
			// collateral weight 0.1, should be able to borrow 10 total
			"AAAA",
			"6.00",
			"A->ACEI max(A)",
		},
		{
			// single asset, with multiple existing borrows, borrowing mid-weighted asset
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("AAAA", "1"),
				coin.Dec("CCCC", "1"),
				coin.Dec("EEEE", "1"),
				coin.Dec("IIII", "1"),
			),
			// collateral weight 0.1, should be able to borrow 10 total
			"CCCC",
			"6.00",
			"A->ACEI max(C)",
		},
		{
			// single asset, with multiple existing borrows, borrowing highest-weighted asset
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("AAAA", "1"),
				coin.Dec("CCCC", "1"),
				coin.Dec("EEEE", "1"),
				coin.Dec("IIII", "1"),
			),
			// collateral weight 0.1, should be able to borrow 10 total
			"GGGG",
			"6.00",
			"A->ACEI max(G)",
		},
		{
			// mid-weight asset, with multiple existing borrows, borrowing lowest-weighted asset
			sdk.NewDecCoins(
				coin.Dec("CCCC", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("AAAA", "1"),
				coin.Dec("CCCC", "1"),
				coin.Dec("EEEE", "1"),
				coin.Dec("IIII", "1"),
			),
			// note that minimum borrow factor is 0.5, so all borrows here are weighted min(0.3,0.5) = 0.3
			// Total borrow will be 100 * 0.3 = 30 so max borrow will be 30 - 4 due to initial borrow
			"AAAA",
			"26.00",
			"C->ACEI max(A)",
		},
		{
			// high-weight asset, with multiple existing borrows, borrowing lowest-weighted asset
			sdk.NewDecCoins(
				coin.Dec("GGGG", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("AAAA", "1"),
				coin.Dec("CCCC", "1"),
				coin.Dec("EEEE", "1"),
				coin.Dec("IIII", "1"),
			),
			// collateral weight 0.5 for all borrows due to minimum borrow factor consumes 8 collateral.
			// remaining max borrow is 92 * 0.5
			"AAAA",
			"46.00",
			"G->ACEI max(A)",
		},
		{
			// high-weight asset, with multiple existing borrows, borrowing highest-weighted asset
			sdk.NewDecCoins(
				coin.Dec("GGGG", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("AAAA", "1"),
				coin.Dec("CCCC", "1"),
				coin.Dec("EEEE", "1"),
				coin.Dec("IIII", "1"),
			),
			// collateral weight 0.5 for all borrows due to minimum borrow factor consumes 8 collateral.
			// remaining max borrow is 92 * 0.7
			"GGGG",
			"64.40", // 46 + (92 * 0.2)
			"G->ACEI max(I)",
		},
		{
			// multiple asset
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
				coin.Dec("EEEE", "100"),
				coin.Dec("IIII", "100"),
			),
			sdk.NewDecCoins(),
			// collateral weights 0.1. 0.5. 0.0, should be able to borrow 10 A + 50 A + 0 A
			"AAAA",
			"60.00",
			"AEI max(A)",
		},
		{
			// multiple asset, with existing looped borrow
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
				coin.Dec("EEEE", "100"),
				coin.Dec("IIII", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("AAAA", "7"),
			),
			// same position - borrow should reach 60 total
			"AAAA",
			"53.00",
			"AEI->A max(A)",
		},
		{
			// single asset, with existing borrow
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
				coin.Dec("EEEE", "100"),
				coin.Dec("IIII", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("IIII", "4"),
			),
			// existing borrow has collateral weight 0.5 due to minimum borrow factor and pairs with 8E
			// the remaning 100A, 100I, 92E can borrow 10+0+46 = 56 more A
			"AAAA",
			"56.00",
			"AEI->I max(A)",
		},
	}

	for _, tc := range testCases {
		borrowPosition, err := types.NewAccountPosition(
			orderedTokens,
			orderedPairs,
			tc.collateral,
			tc.borrow,
			false,
			highMinimumBorrowFactor,
		)
		assert.NilError(t, err, tc.msg+" max borrow\n\n"+borrowPosition.String())
		maxborrow, err := borrowPosition.MaxBorrow(tc.maxBorrowDenom)
		assert.NilError(t, err, tc.msg+" max borrow\n\n"+borrowPosition.String())
		assert.Equal(t,
			sdk.MustNewDecFromStr(tc.maxBorrow).String(),
			maxborrow.String(),
			tc.msg+" max borrow\n\n"+borrowPosition.String(),
		)
	}
}

func TestMaxBorrowWithSpecialPairs(t *testing.T) {
	type testCase struct {
		collateral     sdk.DecCoins
		borrow         sdk.DecCoins
		maxBorrowDenom string
		maxBorrow      string
		msg            string
	}

	// Reminder:
	// F and H are paired at [0.6,0.8] but not looped
	// D can borrow any (B,D,F,H) at [0.5,0.5]
	// B can be borrowed by any (B,D,F,H) at [0.3,0.3]
	// H is looped at [0.1,0.1] even though its token weights are [0.0,0.85]

	testCases := []testCase{
		{
			// single asset, outside special pair
			sdk.NewDecCoins(
				coin.Dec("BBBB", "100"),
			),
			sdk.NewDecCoins(),
			// no special pair with A. collateral weight 0.2
			"AAAA",
			"20.00",
			"simple B max(A)",
		},
		{
			// single asset, loop in special pair
			sdk.NewDecCoins(
				coin.Dec("BBBB", "100"),
			),
			sdk.NewDecCoins(),
			// special pair with B at 0.3
			"BBBB",
			"30.00",
			"simple B max(B)",
		},
		{
			// single asset, borrow in special pair
			sdk.NewDecCoins(
				coin.Dec("BBBB", "100"),
			),
			sdk.NewDecCoins(),
			// special pair with B at 0.3
			"DDDD",
			"30.00",
			"simple B max(D)",
		},
	}

	for _, tc := range testCases {
		borrowPosition, err := types.NewAccountPosition(
			orderedTokens,
			orderedPairs,
			tc.collateral,
			tc.borrow,
			false,
			highMinimumBorrowFactor,
		)
		assert.NilError(t, err, tc.msg+" max borrow\n\n"+borrowPosition.String())
		maxborrow, err := borrowPosition.MaxBorrow(tc.maxBorrowDenom)
		assert.NilError(t, err, tc.msg+" max borrow\n\n"+borrowPosition.String())
		assert.Equal(t,
			sdk.MustNewDecFromStr(tc.maxBorrow).String(),
			maxborrow.String(),
			tc.msg+" max borrow\n\n"+borrowPosition.String(),
		)
	}
}

func TestMaxWithdrawNoSpecialPairs(t *testing.T) {
	type testCase struct {
		collateral       sdk.DecCoins
		borrow           sdk.DecCoins
		maxWithdrawDenom string
		maxWithdraw      string
		msg              string
	}

	testCases := []testCase{
		{
			// single asset
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(),
			// can withdraw all
			"AAAA",
			"100.00",
			"simple A maxWithdraw(A)",
		},
		{
			// single asset, with existing looped borrow
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("AAAA", "7"),
			),
			// collateral weight 0.1, should be able to withdraw 30
			"AAAA",
			"30.00",
			"simple A->A maxWithdraw(A)",
		},
		{
			// single asset, with existing borrow
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("IIII", "4"),
			),
			// collateral weight 0.1, should be able to withdraw 60 total
			"AAAA",
			"60.00",
			"simple A->I maxWithdraw(A)",
		},
		{
			// single asset, with multiple existing borrows
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("AAAA", "1"),
				coin.Dec("CCCC", "1"),
				coin.Dec("EEEE", "1"),
				coin.Dec("IIII", "1"),
			),
			// collateral weight 0.1, should be able to withdraw 60 total
			"AAAA",
			"60.00",
			"A->ACEI maxWithdraw(A)",
		},

		{
			// high-weight asset
			sdk.NewDecCoins(
				coin.Dec("GGGG", "100"),
			),
			sdk.NewDecCoins(),
			// can withdraw all
			"GGGG",
			"100.00",
			"simple G maxWithdraw(G)",
		},
		{
			// high-weight asset, with existing borrow
			sdk.NewDecCoins(
				coin.Dec("GGGG", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("AAAA", "7"),
			),
			// collateral weight 0.5 due to minimum borrow factor, should be able to withdraw 100 - 14
			"GGGG",
			"86.00",
			"simple G->A maxWithdraw(G)",
		},
		{
			// high-weight asset, with existing looped borrow
			sdk.NewDecCoins(
				coin.Dec("GGGG", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("GGGG", "7"),
			),
			// collateral weight 0.7, should be able to withdraw 100 - (7 / 0.7)
			"GGGG",
			"90.00",
			"simple G->G maxWithdraw(G)",
		},
		{
			// high-weight asset, with multiple existing borrows
			sdk.NewDecCoins(
				coin.Dec("GGGG", "100"),
			),
			sdk.NewDecCoins(
				coin.Dec("AAAA", "10"),
				coin.Dec("CCCC", "10"),
				coin.Dec("GGGG", "14"),
				coin.Dec("IIII", "10"),
			),
			// collateral weight 0.5 for A,C,I and 0.7 for G means (30 / 0.5 + 14 / 0.7) collateral
			// is reserved. Max withdraw is thus 100 - (60 + 20)
			"GGGG",
			"20.00",
			"G->ACEI maxWithdraw(G)",
		},
	}

	for _, tc := range testCases {
		borrowPosition, err := types.NewAccountPosition(
			orderedTokens,
			orderedPairs,
			tc.collateral,
			tc.borrow,
			false,
			highMinimumBorrowFactor,
		)
		assert.NilError(t, err, tc.msg+" max withdraw\n\n"+borrowPosition.String())
		maxWithdraw, err := borrowPosition.MaxWithdraw(tc.maxWithdrawDenom)
		assert.NilError(t, err, tc.msg+" max withdraw\n\n"+borrowPosition.String())
		assert.Equal(t,
			sdk.MustNewDecFromStr(tc.maxWithdraw).String(),
			maxWithdraw.String(),
			tc.msg+" max withdraw\n\n"+borrowPosition.String(),
		)
	}
}

// TODO: more cases for positions with multiple borrow and collateral types
// TODO: max borrow and max withdraw tests with special pairs involved
// TODO: clever zero cases, such as max withdraw something that does not exist
