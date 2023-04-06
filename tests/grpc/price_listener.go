package grpc

import (
	"fmt"

	"github.com/umee-network/umee/v4/client"
	sdkclient "github.com/umee-network/umee/v4/sdkclient"
	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"
)

func listenForPrices(
	umee client.Client,
	params oracletypes.Params,
	chainHeight *sdkclient.ChainHeightListener,
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
			exchangeRates, err := umee.QueryExchangeRates()
			fmt.Printf("block %d stamp: %+v\n", height, exchangeRates)
			if err != nil {
				return nil, err
			}
			for _, rate := range exchangeRates {
				priceStore.addStamp(rate.Denom, rate.Amount)
			}
		}
	}

	medians, err := umee.QueryMedians()
	if err != nil {
		return nil, err
	}

	expectedNumMedians := int(params.MaximumMedianStamps) * len(params.AcceptList)
	if len(medians) != expectedNumMedians {
		return nil, fmt.Errorf("amount of medians %d does not match the expected amount %d", len(medians), expectedNumMedians)
	}

	// Saves the last median for each denom
	for _, median := range medians {
		priceStore.medians[median.ExchangeRateTuple.Denom] = median.ExchangeRateTuple.ExchangeRate
	}

	medianDeviations, err := umee.QueryMedianDeviations()
	if err != nil {
		return nil, err
	}

	// Saves the last median deviation for each denom
	for _, medianDeviation := range medianDeviations {
		priceStore.medianDeviations[medianDeviation.ExchangeRateTuple.Denom] = medianDeviation.ExchangeRateTuple.ExchangeRate
	}

	return priceStore, nil
}

func isPeriodFirstBlock(height int64, blocksPerPeriod uint64) bool {
	return uint64(height)%blocksPerPeriod == 0
}
