package coin

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Dec creates a DecCoin with a given base denom and amount
func Dec(denom string, amount string) sdk.DecCoin {
	return sdk.NewDecCoinFromDec(denom, sdk.MustNewDecFromStr(amount))
}

// NewInt creates a DecCoin with a given base denom and integer amount
func DecInt(denom string, amount int64) sdk.DecCoin {
	return sdk.NewInt64DecCoin(denom, amount)
}

func DecFloat(denom string, amount float64) sdk.DecCoin {
	d := sdk.MustNewDecFromStr(strconv.FormatFloat(amount, 'f', -1, 64))
	return sdk.NewDecCoinFromDec(denom, d)
}
