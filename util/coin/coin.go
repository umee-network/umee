package coin

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Zero returns new coin with zero amount
func Zero(denom string) sdk.Coin {
	return sdk.NewInt64Coin(denom, 0)
}

// Normalize transform nil coins to empty list
func Normalize(cs sdk.Coins) sdk.Coins {
	if cs == nil {
		return sdk.Coins{}
	}
	return cs
}
