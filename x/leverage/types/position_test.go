package types

import (
	"testing"

	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func testToken(denom, cw, lt string) Token {
	return Token{
		BaseDenom:            denom,
		CollateralWeight:     sdk.MustNewDecFromStr(cw),
		LiquidationThreshold: sdk.MustNewDecFromStr(lt),
	}
}

func testPair(borrow, collateral, cw, lt string) SpecialAssetPair {
	return SpecialAssetPair{
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
var orderedTokens = []Token{
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
var orderedPairs = []SpecialAssetPair{
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

func TestBorrowLimit(t *testing.T) {
	position := NewAccountPosition(
		orderedTokens,
		orderedPairs,
		sdk.NewDecCoins(
			sdk.NewDecCoinFromDec("AAAA", sdk.MustNewDecFromStr("100.00")),
		),
		sdk.NewDecCoins(),
		false,
	)

	assert.Assert(t, sdk.MustNewDecFromStr("10.00").Equal(position.Limit()), "simple borrow limit")
}

func TestLiquidationThreshold(t *testing.T) {
	position := NewAccountPosition(
		orderedTokens,
		orderedPairs,
		sdk.NewDecCoins(
			sdk.NewDecCoinFromDec("AAAA", sdk.MustNewDecFromStr("100.00")),
		),
		sdk.NewDecCoins(),
		true,
	)

	assert.Assert(t, sdk.MustNewDecFromStr("15.00").Equal(position.Limit()), "simple liquidation threshold")
}
