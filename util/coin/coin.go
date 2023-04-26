package coin

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Zero returns new coin with zero amount
func Zero(denom string) sdk.Coin {
	return sdk.NewInt64Coin(denom, 0)
}

// Zero returns new coin with zero amount
func ZeroDec(denom string) sdk.DecCoin {
	return sdk.NewInt64DecCoin(denom, 0)
}

// Normalize transform nil coins to empty list
func Normalize(cs sdk.Coins) sdk.Coins {
	if cs == nil {
		return sdk.Coins{}
	}
	return cs
}

// New creates a Coin with a given base denom and amount
func New(denom string, amount int64) sdk.Coin {
	return sdk.NewInt64Coin(denom, amount)
}
