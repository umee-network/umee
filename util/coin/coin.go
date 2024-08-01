package coin

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var Exponents = map[int]sdkmath.LegacyDec{
	-18: sdkmath.LegacyMustNewDecFromStr("0.000000000000000001"),
	-17: sdkmath.LegacyMustNewDecFromStr("0.00000000000000001"),
	-16: sdkmath.LegacyMustNewDecFromStr("0.0000000000000001"),
	-15: sdkmath.LegacyMustNewDecFromStr("0.000000000000001"),
	-14: sdkmath.LegacyMustNewDecFromStr("0.00000000000001"),
	-13: sdkmath.LegacyMustNewDecFromStr("0.0000000000001"),
	-12: sdkmath.LegacyMustNewDecFromStr("0.000000000001"),
	-11: sdkmath.LegacyMustNewDecFromStr("0.00000000001"),
	-10: sdkmath.LegacyMustNewDecFromStr("0.0000000001"),
	-9:  sdkmath.LegacyMustNewDecFromStr("0.000000001"),
	-8:  sdkmath.LegacyMustNewDecFromStr("0.00000001"),
	-7:  sdkmath.LegacyMustNewDecFromStr("0.0000001"),
	-6:  sdkmath.LegacyMustNewDecFromStr("0.000001"),
	-5:  sdkmath.LegacyMustNewDecFromStr("0.00001"),
	-4:  sdkmath.LegacyMustNewDecFromStr("0.0001"),
	-3:  sdkmath.LegacyMustNewDecFromStr("0.001"),
	-2:  sdkmath.LegacyMustNewDecFromStr("0.01"),
	-1:  sdkmath.LegacyMustNewDecFromStr("0.1"),
	0:   sdkmath.LegacyMustNewDecFromStr("1.0"),
	1:   sdkmath.LegacyMustNewDecFromStr("10.0"),
	2:   sdkmath.LegacyMustNewDecFromStr("100.0"),
	3:   sdkmath.LegacyMustNewDecFromStr("1000.0"),
	4:   sdkmath.LegacyMustNewDecFromStr("10000.0"),
	5:   sdkmath.LegacyMustNewDecFromStr("100000.0"),
	6:   sdkmath.LegacyMustNewDecFromStr("1000000.0"),
	7:   sdkmath.LegacyMustNewDecFromStr("10000000.0"),
	8:   sdkmath.LegacyMustNewDecFromStr("100000000.0"),
	9:   sdkmath.LegacyMustNewDecFromStr("1000000000.0"),
	10:  sdkmath.LegacyMustNewDecFromStr("10000000000.0"),
	11:  sdkmath.LegacyMustNewDecFromStr("100000000000.0"),
	12:  sdkmath.LegacyMustNewDecFromStr("1000000000000.0"),
	13:  sdkmath.LegacyMustNewDecFromStr("10000000000000.0"),
	14:  sdkmath.LegacyMustNewDecFromStr("100000000000000.0"),
	15:  sdkmath.LegacyMustNewDecFromStr("1000000000000000.0"),
	16:  sdkmath.LegacyMustNewDecFromStr("10000000000000000.0"),
	17:  sdkmath.LegacyMustNewDecFromStr("100000000000000000.0"),
	18:  sdkmath.LegacyMustNewDecFromStr("1000000000000000000.0"),
}

// Zero returns new coin with zero amount
func Zero(denom string) sdk.Coin {
	return sdk.NewInt64Coin(denom, 0)
}

// ZeroDec returns new decCoin with zero amount
func ZeroDec(denom string) sdk.DecCoin {
	return sdk.NewInt64DecCoin(denom, 0)
}

// One returns new coin with one amount
func One(denom string) sdk.Coin {
	return sdk.NewInt64Coin(denom, 1)
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
