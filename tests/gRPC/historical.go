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

	priceStore := &PriceStore{}

	listenForPrices(ctx, val1Client, priceStore)

	fmt.Printf("%+v\n", priceStore)

	cancel()
	return nil
}

func listenForPrices(
	ctx context.Context,
	umeeClient *UmeeClient,
	priceStore *PriceStore,
) {
	chainHeight, _ := NewChainHeight(
		ctx,
		umeeClient.clientContext.Client,
		zerolog.Nop(),
	)

	params, _ := umeeClient.QueryParams()

	for i := 1; i <= int(params.MedianStampPeriod*2); i++ {
		select {
		case <-ctx.Done():
			return
		case height := <-chainHeight.HeightChanged:
			if isPeriodFirstBlock(height, params.HistoricStampPeriod) {
				exchangeRates, _ := umeeClient.QueryExchangeRates()
				for _, rate := range exchangeRates {
					priceStore.SetHistoricStamp(height, rate.Denom, rate.Amount)
				}
			}
			if isPeriodFirstBlock(height, params.MedianStampPeriod) {
				medians, _ := umeeClient.QueryMedians()
				for _, median := range medians {
					priceStore.SetMedian(height, median.Denom, median.Amount)
				}
			}
		}
	}
}

func isPeriodFirstBlock(height int64, blocksPerPeriod uint64) bool {
	return uint64(height)%blocksPerPeriod == 0
}
