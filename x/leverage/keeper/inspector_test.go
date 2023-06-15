package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestNeat(t *testing.T) {
	assert := assert.New(t)

	cases := map[string]float64{
		// example
		"1000": 1000, // sdk.Dec created from "1000" is converted to float 1000.0
		// tests
		"123456789.55":      123456000,         // truncates >1M to thousand
		"123456.55":         123456,            // truncates >100 to whole number
		"12.555":            12.55,             // truncates default to cent
		"0.00123456789":     0.001234,          // truncates <0.01 to millionth
		"0.000000987654321": 0.000000987654321, // <0.000001 gets maximum precision
	}

	for s, f := range cases {
		assert.Equal(f, neat(sdk.MustNewDecFromStr(s)))
	}
}
