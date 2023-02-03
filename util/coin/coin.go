package coin

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ZeroCoin returns new coin with zero amount
func ZeroCoin(denom string) sdk.Coin {
	return sdk.NewInt64Coin(denom, 0)
}

// Normalize transform nil coins to empty list
func NormalizeCoins(cs sdk.Coins) sdk.Coins {
	if cs == nil {
		return sdk.Coins{}
	}
	return cs
}
