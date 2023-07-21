package metoken

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v5/util/coin"
)

// IndexPrices holds meToken and all the underlying assets Price.
type IndexPrices struct {
	prices map[string]Price
}

// Price contains usd Price from x/oracle and Exponent from x/leverage
type Price struct {
	Price    sdk.Dec
	Exponent uint32
}

// NewIndexPrices creates an instance of IndexPrices.
func NewIndexPrices() IndexPrices {
	return IndexPrices{prices: make(map[string]Price)}
}

// Price returns a Price given a specific denom.
func (ip IndexPrices) Price(denom string) (Price, error) {
	price, found := ip.prices[denom]
	if !found || !price.Price.IsPositive() {
		return Price{}, sdkerrors.ErrNotFound.Wrapf("price not found for denom %s", denom)
	}

	return price, nil
}

// SetPrice to the IndexPrices.
func (ip IndexPrices) SetPrice(denom string, price sdk.Dec, exponent uint32) {
	ip.prices[denom] = Price{
		Price:    price,
		Exponent: exponent,
	}
}

// SwapRate converts one asset in the index to another applying exchange_rate and normalizing the exponent.
func (ip IndexPrices) SwapRate(from sdk.Coin, to string) (sdkmath.Int, error) {
	exchangeRate, err := ip.rate(from.Denom, to)
	if err != nil {
		return sdkmath.Int{}, err
	}

	exponentFactor, err := ip.ExponentFactor(from.Denom, to)
	if err != nil {
		return sdkmath.Int{}, err
	}

	return exchangeRate.MulInt(from.Amount).Mul(exponentFactor).TruncateInt(), nil
}

// rate calculates the exchange rate based on IndexPrices.
func (ip IndexPrices) rate(from, to string) (sdk.Dec, error) {
	fromPrice, err := ip.Price(from)
	if err != nil {
		return sdk.Dec{}, err
	}

	toPrice, err := ip.Price(to)
	if err != nil {
		return sdk.Dec{}, err
	}

	return fromPrice.Price.Quo(toPrice.Price), nil
}

// ExponentFactor calculates the multiplayer to be used to multiply from denom to get same decimals as to denom.
// If there is no such difference the result will be 1.
func (ip IndexPrices) ExponentFactor(from, to string) (sdk.Dec, error) {
	fromPrice, err := ip.Price(from)
	if err != nil {
		return sdk.Dec{}, err
	}

	toPrice, err := ip.Price(to)
	if err != nil {
		return sdk.Dec{}, err
	}

	return ExponentFactor(fromPrice.Exponent, toPrice.Exponent)
}

// ExponentFactor calculates the factor to multiply by which the assets with different exponents.
// If there is no such difference the result will be 1.
func ExponentFactor(initialExponent, resultExponent uint32) (sdk.Dec, error) {
	exponentDiff := int(resultExponent) - int(initialExponent)
	multiplier, ok := coin.Exponents[exponentDiff]
	if !ok {
		return sdk.Dec{}, fmt.Errorf("multiplier not found for exponentDiff %d", exponentDiff)
	}

	return multiplier, nil
}
