package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
)

func TestNeat(t *testing.T) {
	assert := assert.New(t)

	cases := map[string]float64{
		// example
		"1000": 1000, // sdkmath.LegacyDec created from "1000" is converted to float 1000.0
		// tests
		"123456789.55":      123456000,         // truncates >1M to thousand
		"123456.55":         123456,            // truncates >100 to whole number
		"12.555":            12.55,             // truncates default to cent
		"0.00123456789":     0.001234,          // truncates <0.01 to millionth
		"0.000000987654321": 0.000000987654321, // <0.000001 gets maximum precision
		// negative
		"-123456789.55":      -123456000,         // truncates >1M to thousand
		"-123456.55":         -123456,            // truncates >100 to whole number
		"-12.555":            -12.55,             // truncates default to cent
		"-0.00123456789":     -0.001234,          // truncates <0.01 to millionth
		"-0.000000987654321": -0.000000987654321, // <0.000001 gets maximum precision
	}

	for s, f := range cases {
		assert.Equal(f, neat(sdkmath.LegacyMustNewDecFromStr(s)))
	}

	// edge case: >2^64 displays incorrectly
	// this should be fine, since this is a display-only function (not used in transactions)
	// which is used on dollar (not token) amounts
	assert.NotEqual(
		123456789123456789123456789.123456789,
		neat(sdkmath.LegacyMustNewDecFromStr("123456789123456789123456789.123456789")),
	)
}
