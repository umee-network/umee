package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func toDec(i sdk.Int) sdk.Dec {
	return sdk.NewDecFromInt(i)
}
