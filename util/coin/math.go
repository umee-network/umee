package coin

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: add unit tests for `util/coin/math.go`
// DecBld is a Builder pattern for dec coin
type DecBld struct {
	D sdk.DecCoin
}

// NewDecBld is a constructor for DecBld type
func NewDecBld(d sdk.DecCoin) *DecBld {
	return &DecBld{D: d}
}

// Scale scales dec coin by given factor
func (d *DecBld) Scale(f int64) *DecBld {
	d.D = sdk.DecCoin{Denom: d.D.Denom, Amount: d.D.Amount.MulInt64(f)}
	return d
}

// Scale scales dec coin by given factor provided as string.
// Panics if f is not a correct decimal number.
func (d *DecBld) ScaleStr(f string) *DecBld {
	d.D = sdk.DecCoin{Denom: d.D.Denom, Amount: d.D.Amount.Mul(sdk.MustNewDecFromStr(f))}
	return d
}

// ToCoin converts DecCoin to sdk.Coin rounding up
func (d *DecBld) ToCoin() sdk.Coin {
	return sdk.NewCoin(d.D.Denom, d.D.Amount.Ceil().RoundInt())
}

// ToCoins converts DecCoin to sdk.Coins rounding up
func (d *DecBld) ToCoins() sdk.Coins {
	return sdk.Coins{d.ToCoin()}
}

// ToCoin converts DecCoin to sdk.DecCoins
func (d *DecBld) ToDecCoins() sdk.DecCoins {
	return sdk.DecCoins{d.D}
}
