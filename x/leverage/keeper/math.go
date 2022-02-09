package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

var ten = sdk.MustNewDecFromStr("10")

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

// oneTenthFromExponent returns an sdk.Dec corresponding to 0.1 times 10^x. This is used, for example, to
// determine that the minimum required amount of collateral for a denomination with exponent 6 is 10^5.
func oneTenthFromExponent(exponent uint32) sdk.Int {
	if exponent < 2 {
		return sdk.OneInt()
	}
	return ten.Power(uint64(exponent - 1)).TruncateInt()
}
