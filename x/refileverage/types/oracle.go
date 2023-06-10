package types

// Enumerates different ways to request the price of a token
type PriceMode uint64

const (
	// Spot mode requests the most recent prices from oracle
	PriceModeSpot PriceMode = iota
	// Historic mode requests the median of the most recent historic medians
	PriceModeHistoric
	// High mode uses the higher of either Spot or Historic prices
	PriceModeHigh
	// Low mode uses the lower of either Spot or Historic prices
	PriceModeLow
)
