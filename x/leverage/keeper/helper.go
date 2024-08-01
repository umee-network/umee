package keeper

import (
	sdkmath "cosmossdk.io/math"
)

func toDec(i sdkmath.Int) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecFromInt(i)
}
