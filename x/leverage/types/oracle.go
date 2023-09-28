package types

// Enumerates different ways to request the price of a token
type PriceMode uint64

const (
	// Spot mode requests the most recent prices from oracle, unless they are too old.
	PriceModeSpot PriceMode = iota
	// Query mode requests the most recent price, regardless of price age.
	PriceModeQuery
	// Historic mode requests the median of the most recent historic medians
	PriceModeHistoric
	// High mode uses the higher of either Spot or Historic prices
	PriceModeHigh
	// QueryHigh mode uses the higher of either Spot or Historic prices, allowing expired spot prices.
	PriceModeQueryHigh
	// Low mode uses the lower of either Spot or Historic prices
	PriceModeLow
	// QueryLow mode uses the lower of either Spot or Historic prices, allowing expired spot prices.
	PriceModeQueryLow
)

// IgnoreHistoric transforms a price mode in a way that uses spot prices instead of historic. This
// happens when historic prices would normally be used by a token has zero required historic medians.
func (mode PriceMode) IgnoreHistoric() PriceMode {
	switch mode {
	case PriceModeHigh, PriceModeLow, PriceModeHistoric:
		// historic modes default to spot prices
		return PriceModeSpot
	case PriceModeQueryHigh, PriceModeQueryLow:
		// historic-enabled queries allow expired prices as well
		return PriceModeQuery
	}
	// other modes are unmodified
	return mode
}

// AllowsExpired returns true if a price mode allows expired spot prices to be used
func (mode PriceMode) AllowsExpired() bool {
	switch mode {
	case PriceModeQuery, PriceModeQueryHigh, PriceModeQueryLow:
		return true
	}
	return false
}
