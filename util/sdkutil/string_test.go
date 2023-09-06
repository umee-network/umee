package sdkutil

import (
	"testing"

	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestFormatDec(t *testing.T) {
	type testCase struct {
		input  string
		output string
	}

	testCases := []testCase{
		{
			"0",
			"0",
		},
		{
			"1.00",
			"1",
		},
		{
			"1.23",
			"1.23",
		},
		{
			"1.500000",
			"1.5",
		},
		{
			"1234.567800000",
			"1234.5678",
		},
		{
			"-250.02130",
			"-250.0213",
		},
		{
			"-0.73190",
			"-0.7319",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t,
			tc.output,
			FormatDec(sdk.MustNewDecFromStr(tc.input)),
		)
	}
}

func TestFormatDecCoin(t *testing.T) {
	type testCase struct {
		amount string
		denom  string
		output string
	}

	testCases := []testCase{
		{
			"1.00",
			"AAAA",
			"1 AAAA",
		},
		{
			"1.23",
			"BBBB",
			"1.23 BBBB",
		},
		{
			"1.500000",
			"ibc/CCCC",
			"1.5 ibc/CCCC",
		},
		{
			"1234.567800000",
			"u/DDDD",
			"1234.5678 u/DDDD",
		},
		{
			"0",
			"EEEE",
			"0 EEEE",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t,
			tc.output,
			FormatDecCoin(sdk.NewDecCoinFromDec(tc.denom, sdk.MustNewDecFromStr(tc.amount))),
		)
	}
}
