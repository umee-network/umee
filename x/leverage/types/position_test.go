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

func TestSimpleBorrowLimit(t *testing.T) {
	type testCase struct {
		collateral    sdk.DecCoins
		borrow        sdk.DecCoins
		limit         string
		isLiquidation bool
		msg           string
	}

	testCases := []testCase{
		{
			// single asset
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(),
			"10.00",
			false,
			"collateral weight A",
		},
		{
			// single asset
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
			),
			sdk.NewDecCoins(),
			"15.00",
			true,
			"liquidation threshold A",
		},
		{
			// multiple assets, one with zero weight
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
				coin.Dec("GGGG", "100"),
				coin.Dec("IIII", "100"),
			),
			sdk.NewDecCoins(),
			"80.00",
			false,
			"collateral weight AGI",
		},
		{
			// multiple assets
			sdk.NewDecCoins(
				coin.Dec("AAAA", "100"),
				coin.Dec("GGGG", "100"),
				coin.Dec("IIII", "100"),
			),
			sdk.NewDecCoins(),
			"185.00",
			true,
			"liquidation threshold AGI",
		},
	}

	for _, tc := range testCases {
		position := types.NewAccountPosition(
			orderedTokens,
			orderedPairs,
			tc.collateral,
			tc.borrow,
			tc.isLiquidation,
		)
		// assert.Equal(t, position.String(), "")
		assert.Equal(t, sdk.MustNewDecFromStr(tc.limit).String(), (position.Limit().String()), tc.msg)
	}
}
