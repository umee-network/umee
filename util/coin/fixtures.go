package coin

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v4/app/params"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// common coins used in tests
var (
	UumeeDenom = leveragetypes.ToUTokenDenom(appparams.BondDenom)
	Umee1      = New(appparams.BondDenom, 1)
	Umee10k    = New(appparams.BondDenom, 10_000)
	UUmee1     = UUmee(1)

	Umee1dec = DecFloat(appparams.BondDenom, 1)

	Atom    = "atom"
	Atom1   = New(Atom, 1)
	Atom10k = New(Atom, 10_000)
	UAtom1  = Utoken(Atom, 1)

	Atom1dec    = DecFloat(Atom, 1)
	Atom1_25dec = DecFloat(Atom, 1.25)
)

// UUmee creates a uToken UMEE with given amount
func UmeeDec(amount float64) sdk.DecCoin {
	return DecFloat(appparams.BondDenom, amount)
}

// UUmee creates a uToken UMEE with given amount
func UUmee(amount int64) sdk.Coin {
	return New(UumeeDenom, amount)
}

// UUmee creates a uToken UMEE with given amount
func UUmeeDec(amount float64) sdk.DecCoin {
	return DecFloat(UumeeDenom, amount)
}

// Utoken creates a uToken with given base denom and amount
func Utoken(denom string, amount int64) sdk.Coin {
	return New(leveragetypes.ToUTokenDenom(denom), amount)
}
