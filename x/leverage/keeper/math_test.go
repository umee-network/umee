package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"gotest.tools/v3/assert"
)

func TestInterpolate(t *testing.T) {
	// Define two points (x1,y1) and (x2,y2)
	x1 := sdkmath.LegacyMustNewDecFromStr("3.0")
	x2 := sdkmath.LegacyMustNewDecFromStr("6.0")
	y1 := sdkmath.LegacyMustNewDecFromStr("11.1")
	y2 := sdkmath.LegacyMustNewDecFromStr("17.4")

	// Sloped line, endpoint checks
	result := Interpolate(x1, x1, y1, x2, y2)
	assert.DeepEqual(t, y1, result)
	result = Interpolate(x2, x1, y1, x2, y2)
	assert.DeepEqual(t, y2, result)

	// Sloped line, point on segment
	result = Interpolate(sdkmath.LegacyMustNewDecFromStr("4.0"), x1, y1, x2, y2)
	assert.DeepEqual(t, sdkmath.LegacyMustNewDecFromStr("13.2"), result)

	// Sloped line, point outside of segment
	result = Interpolate(sdkmath.LegacyMustNewDecFromStr("2.0"), x1, y1, x2, y2)
	assert.DeepEqual(t, sdkmath.LegacyMustNewDecFromStr("9.0"), result)

	// Vertical line: always return y1
	result = Interpolate(sdkmath.LegacyZeroDec(), x1, y1, x1, y2)
	assert.Equal(t, y1, result)
	result = Interpolate(x1, x1, y1, x1, y2)
	assert.DeepEqual(t, y1, result)

	// Undefined line (x1=x2, y1=y2): always return y1
	result = Interpolate(sdkmath.LegacyZeroDec(), x1, y1, x1, y1)
	assert.DeepEqual(t, y1, result)
	result = Interpolate(x1, x1, y1, x1, y1)
	assert.DeepEqual(t, y1, result)
}
