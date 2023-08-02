package types

import (
	"testing"

	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBorrowLimit(t *testing.T) {
	position := NewAccountPosition(
		[]Token{
			{
				BaseDenom:            "AAAA",
				CollateralWeight:     sdk.MustNewDecFromStr("0.25"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.5"),
			},
		},
		[]SpecialAssetPair{},
		sdk.NewDecCoins(
			sdk.NewDecCoinFromDec("AAAA", sdk.MustNewDecFromStr("100.00")),
		),
		sdk.NewDecCoins(),
		false,
	)

	assert.Assert(t, sdk.MustNewDecFromStr("25.00").Equal(position.Limit()), "simple borrow limit")
}

func TestLiquidationThreshold(t *testing.T) {
	position := NewAccountPosition(
		[]Token{
			{
				BaseDenom:            "AAAA",
				CollateralWeight:     sdk.MustNewDecFromStr("0.25"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.5"),
			},
		},
		[]SpecialAssetPair{},
		sdk.NewDecCoins(
			sdk.NewDecCoinFromDec("AAAA", sdk.MustNewDecFromStr("100.00")),
		),
		sdk.NewDecCoins(),
		true,
	)

	assert.Assert(t, sdk.MustNewDecFromStr("50.00").Equal(position.Limit()), "simple liquidation threshold")
}
