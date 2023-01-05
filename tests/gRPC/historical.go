package gRPC

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

func MedianCheck(
	chainID string,
	tmrpcEndpoint string,
	grpcEndpoint string,
	val1Mnemonic string,
) error {
	val1Client, err := NewUmeeClient(chainID, tmrpcEndpoint, grpcEndpoint, "val1", val1Mnemonic)
	if err != nil {
		return err
	}

	err = val1Client.createClientContext()
	if err != nil {
		return err
	}

	val1Client.createQueryClient()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	priceStore := &PriceStore{}

	err = listenForPrices(ctx, val1Client, "umee", priceStore)
	if err != nil {
		return err
	}

	return priceStore.checkMedian()
}

func listenForPrices(
	ctx context.Context,
	umeeClient *UmeeClient,
	denom string,
	priceStore *PriceStore,
) error {
	chainHeight, err := NewChainHeight(ctx, umeeClient.clientContext.Client, zerolog.Nop())
	if err != nil {
		return err
	}

	params, _ := umeeClient.QueryParams() // TODO error handling
	fmt.Printf("%+v\n", params)

	// Wait until the end of a median period
	var beginningHeight int64
	for {
		beginningHeight = <-chainHeight.HeightChanged
		if isPeriodLastBlock(beginningHeight, params.MedianStampPeriod) {
			fmt.Printf("%d: ", beginningHeight)
			fmt.Println("median stamp period last block")
			break
		}
	}

	// Record each historic stamp when the chain should be recording them
	for i := 0; i < int(params.MedianStampPeriod); i++ {
		height := <-chainHeight.HeightChanged
		if isPeriodFirstBlock(height, params.HistoricStampPeriod) {
			fmt.Printf("%d: ", height)
			fmt.Println("historic stamp period first block")
			exchangeRates, err := umeeClient.QueryExchangeRates()
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

	// Wait one more block for the median
	height := <-chainHeight.HeightChanged
	fmt.Printf("%d: ", height)
	fmt.Println("reading final median")
	medians, err := umeeClient.QueryMedians()
	if err != nil {
		return err
	}
	for _, median := range medians {
		if median.Denom == denom {
			priceStore.median = median.Amount
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
