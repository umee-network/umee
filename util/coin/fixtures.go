package coin

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v6/app/params"
)

const umee = appparams.BondDenom

// common coins used in tests
//
//revive:disable:var-naming
var (
	// the uToken denom "u/uumee"
	U_umee = ToUTokenDenom(umee) //nolint:stylecheck
	// 1uumee Coin
	Umee1 = New(umee, 1)
	// 10_000uumee Coin
	Umee10k = New(umee, 10_000)
	// 1u/uumee Coin
	U_umee1 = Utoken(umee, 1) //nolint:stylecheck

	// Xuumee DecCoin
	Umee0dec = DecF(umee, 0)
	Umee1dec = DecF(umee, 1)

	Atom = "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9"
	// 1<atom ibc denom> Coin
	Atom1 = New(Atom, 1)
	// 1u/<atom ibc denom> Coin
	UAtom1 = Utoken(Atom, 1)

	Atom1dec = DecF(Atom, 1)
	// 1.25<atom ibc denom> DecCoin
	Atom1_25dec = DecF(Atom, 1.25)

	// a test denom
	Dollar = "dollar"
)

//revive:enable:var-naming

// Umee creates a BondDenom sdk.Coin with the given amount
func Umee(amount int64) sdk.Coin {
	return sdk.NewInt64Coin(umee, amount)
}

// UmeeInt creates a BondDenom sdk.Coin with the given amount
func UmeeInt(amount sdkmath.Int) sdk.Coin {
	return sdk.NewCoin(umee, amount)
}

// UmeeCoins creates an Umee (uumee) sdk.Coins with the given amount
func UmeeCoins(amount int64) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(umee, amount))
}

// UmeeDec creates a Umee (uumee) DecCoin with given amount
func UmeeDec(amount string) sdk.DecCoin {
	return Dec(appparams.BondDenom, amount)
}

// Utoken creates a uToken Coin.
func Utoken(denom string, amount int64) sdk.Coin {
	return New(ToUTokenDenom(denom), amount)
}

// UtokenDec creates a uToken DecCoin.
func UtokenDec(denom string, amount string) sdk.DecCoin {
	return Dec(ToUTokenDenom(denom), amount)
}

// UtokenDecF creates a uToken DecCoin.
func UtokenDecF(denom string, amount float64) sdk.DecCoin {
	return DecF(ToUTokenDenom(denom), amount)
}
