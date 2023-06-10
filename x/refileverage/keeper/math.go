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

// ApproxExponential is the taylor series expansion of e^x centered around x=0, truncated
// to the cubic term. It can be used with great accuracy to determine e^x when x is very small.
// Note that e^x = 1 + x/1! + x^2/2! + x^3 / 3! + ...
func ApproxExponential(x sdk.Dec) sdk.Dec {
	sum := sdk.OneDec()             // 1
	sum = sum.Add(x)                // x / 1!
	next := x.Mul(x)                // x^2
	sum = sum.Add(next.QuoInt64(2)) // 2!
	next = next.Mul(x)              // x^3
	sum = sum.Add(next.QuoInt64(6)) // 3!
	return sum                      // approximated e^x
}
