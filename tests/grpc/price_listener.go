package grpc

import (
	"fmt"

	"github.com/umee-network/umee/v4/tests/grpc/client"
	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"
)

func listenForPrices(
	umeeClient *client.UmeeClient,
	params oracletypes.Params,
	chainHeight *ChainHeight,
) (*PriceStore, error) {
	priceStore := NewPriceStore()
	// Wait until the beginning of a median period
	var beginningHeight int64
	for {
		beginningHeight = <-chainHeight.HeightChanged
		if isPeriodFirstBlock(beginningHeight, params.MedianStampPeriod) {
			break
		}
	}

	// Record each historic stamp when the chain should be recording them
	for i := 0; i < int(params.MedianStampPeriod); i++ {
		height := <-chainHeight.HeightChanged
		if isPeriodFirstBlock(height, params.HistoricStampPeriod) {
			exchangeRates, err := umeeClient.QueryClient.QueryExchangeRates()
			fmt.Printf("block %d stamp: %+v\n", height, exchangeRates)
			if err != nil {
				return nil, err
			}
			for _, rate := range exchangeRates {
				priceStore.addStamp(rate.Denom, rate.Amount)
			}
		}
	}

	medians, err := umeeClient.QueryClient.QueryMedians()
	if err != nil {
		return nil, err
	}
	// Saves the last median for each denom
	for _, median := range medians {
		priceStore.medians[median.Denom] = median.Amount
	}

	medianDeviations, err := umeeClient.QueryClient.QueryMedianDeviations()
	if err != nil {
		return nil, err
	}
	// Saves the last median deviation for each denom
	for _, medianDeviation := range medianDeviations {
		priceStore.medianDeviations[medianDeviation.Denom] = medianDeviation.Amount
	}

	return priceStore, nil
}

func isPeriodFirstBlock(height int64, blocksPerPeriod uint64) bool {
	return uint64(height)%blocksPerPeriod == 0
}
