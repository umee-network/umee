package coin

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var Exponents = map[int]sdk.Dec{
	-18: sdk.MustNewDecFromStr("0.000000000000000001"),
	-17: sdk.MustNewDecFromStr("0.00000000000000001"),
	-16: sdk.MustNewDecFromStr("0.0000000000000001"),
	-15: sdk.MustNewDecFromStr("0.000000000000001"),
	-14: sdk.MustNewDecFromStr("0.00000000000001"),
	-13: sdk.MustNewDecFromStr("0.0000000000001"),
	-12: sdk.MustNewDecFromStr("0.000000000001"),
	-11: sdk.MustNewDecFromStr("0.00000000001"),
	-10: sdk.MustNewDecFromStr("0.0000000001"),
	-9:  sdk.MustNewDecFromStr("0.000000001"),
	-8:  sdk.MustNewDecFromStr("0.00000001"),
	-7:  sdk.MustNewDecFromStr("0.0000001"),
	-6:  sdk.MustNewDecFromStr("0.000001"),
	-5:  sdk.MustNewDecFromStr("0.00001"),
	-4:  sdk.MustNewDecFromStr("0.0001"),
	-3:  sdk.MustNewDecFromStr("0.001"),
	-2:  sdk.MustNewDecFromStr("0.01"),
	-1:  sdk.MustNewDecFromStr("0.1"),
	0:   sdk.MustNewDecFromStr("1.0"),
	1:   sdk.MustNewDecFromStr("10.0"),
	2:   sdk.MustNewDecFromStr("100.0"),
	3:   sdk.MustNewDecFromStr("1000.0"),
	4:   sdk.MustNewDecFromStr("10000.0"),
	5:   sdk.MustNewDecFromStr("100000.0"),
	6:   sdk.MustNewDecFromStr("1000000.0"),
	7:   sdk.MustNewDecFromStr("10000000.0"),
	8:   sdk.MustNewDecFromStr("100000000.0"),
	9:   sdk.MustNewDecFromStr("1000000000.0"),
	10:  sdk.MustNewDecFromStr("10000000000.0"),
	11:  sdk.MustNewDecFromStr("100000000000.0"),
	12:  sdk.MustNewDecFromStr("1000000000000.0"),
	13:  sdk.MustNewDecFromStr("10000000000000.0"),
	14:  sdk.MustNewDecFromStr("100000000000000.0"),
	15:  sdk.MustNewDecFromStr("1000000000000000.0"),
	16:  sdk.MustNewDecFromStr("10000000000000000.0"),
	17:  sdk.MustNewDecFromStr("100000000000000000.0"),
	18:  sdk.MustNewDecFromStr("1000000000000000000.0"),
}

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

// Negative1 creates a Coin with amount = -1
func Negative1(denom string) sdk.Coin {
	return sdk.Coin{
		Denom:  denom,
		Amount: sdkmath.NewInt(-1),
	}
}
