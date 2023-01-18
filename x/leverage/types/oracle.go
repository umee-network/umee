package types

type PriceMode uint64 // Enumerates different ways to request the price of a token

const (
	PriceModeSpot     PriceMode = iota // Spot mode requests the most recent prices from oracle
	PriceModeHistoric                  // Historic mode requests the median of the most recent historic medians
	PriceModeHigh                      // High mode uses the higher of either Spot or Historic prices
	PriceModeLow                       // Low mode uses the lower of either Spot or Historic prices
)
