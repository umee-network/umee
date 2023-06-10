package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func toDec(i sdkmath.Int) sdk.Dec {
	return sdk.NewDecFromInt(i)
}
