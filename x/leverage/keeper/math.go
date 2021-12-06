package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// Interpolate takes a line defined by two points (xMin, yMin) and (xMax, yMax), then finds the y-value of the
// point on that line for an input x-value. It will return yMin if xMin = xMax (i.e. a vertical line).
// While this function is intended for interpolation (xMin < x < xMax), it works correctly even when x is outside
// that range or when xMin > xMax.
func Interpolate(x, xMin, yMin, xMax, yMax sdk.Dec) sdk.Dec {
	if xMin.Equal(xMax) {
		return yMin
	}
	slope := yMax.Sub(yMin).Quo(xMax.Sub(xMin))
	// y = y1 + m(x-x1)
	return yMin.Add(x.Sub(xMin).Mul(slope))
}

// MinInt returns the minimum of two sdk.Int
func MinInt(x, y sdk.Int) sdk.Int {
	if x.GTE(y) {
		return y
	}
	return x
}
