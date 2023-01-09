package grpc

import (
	"fmt"

	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"
)

func listenForPrices(
	umeeClient *UmeeClient,
	params oracletypes.Params,
	denom string,
	priceStore *PriceStore,
	chainHeight *ChainHeight,
) error {
	// Wait until the end of a median period
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
			exchangeRates, err := umeeClient.QueryExchangeRates()
			fmt.Printf("%d: ", height)
			fmt.Println(exchangeRates)
			if err != nil {
				return nil
			}
			for _, rate := range exchangeRates {
				if rate.Denom == denom {
					priceStore.historicStamps = append(priceStore.historicStamps, rate.Amount)
				}
			}
		}
	}

	medians, err := umeeClient.QueryMedians()
	if err != nil {
		return err
	}
	for _, median := range medians {
		if median.Denom == denom {
			priceStore.median = median.Amount
		}
	}

	medianDeviations, err := umeeClient.QueryMedianDeviations()
	if err != nil {
		return err
	}
	for _, medianDeviation := range medianDeviations {
		if medianDeviation.Denom == denom {
			priceStore.medianDeviation = medianDeviation.Amount
		}
	}

	return nil
}

func isPeriodFirstBlock(height int64, blocksPerPeriod uint64) bool {
	return uint64(height)%blocksPerPeriod == 0
}

func isPeriodLastBlock(height int64, blocksPerPeriod uint64) bool {
	return uint64(height+1)%blocksPerPeriod == 0
}
