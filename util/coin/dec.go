package coin

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Dec creates a DecCoin with a given base denom and amount.
func Dec(denom, amount string) sdk.DecCoin {
	return sdk.NewDecCoinFromDec(denom, sdk.MustNewDecFromStr(amount))
}

// DecF creates a DecCoin based on float amount.
// MUST not be used in production code. Can only be used in tests.
func DecF(denom string, amount float64) sdk.DecCoin {
	d := sdk.MustNewDecFromStr(strconv.FormatFloat(amount, 'f', -1, 64))
	return sdk.NewDecCoinFromDec(denom, d)
}
