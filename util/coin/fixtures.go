package coin

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v4/app/params"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

const umee = appparams.BondDenom

// common coins used in tests
var (
	UumeeDenom = leveragetypes.ToUTokenDenom(umee)
	Umee1      = New(umee, 1)
	Umee10k    = New(umee, 10_000)
	UUmee1     = Utoken(umee, 1)

	Umee1dec = DecF(umee, 1)

	Atom    = "uatom"
	Atom1   = New(Atom, 1)
	Atom10k = New(Atom, 10_000)
	UAtom1  = Utoken(Atom, 1)

	Atom1dec    = DecF(Atom, 1)
	Atom1_25dec = DecF(Atom, 1.25)
)

// UmeeDec creates a Umee (uumee) DecCoin with given amount
func UmeeDec(amount float64) sdk.DecCoin {
	return DecF(appparams.BondDenom, amount)
}

// Utoken creates a uToken coin.
func Utoken(denom string, amount int64) sdk.Coin {
	return New(leveragetypes.ToUTokenDenom(denom), amount)
}

// UtokenDec creates a uToken DecCoin.
func UtokenDec(denom string, amount string) sdk.DecCoin {
	return Dec(leveragetypes.ToUTokenDenom(denom), amount)
}

// UtokenDec creates a uToken DecCoin.
func UtokenDecF(denom string, amount float64) sdk.DecCoin {
	return DecF(leveragetypes.ToUTokenDenom(denom), amount)
}
