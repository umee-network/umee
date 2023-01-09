package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"
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

	err = val1Client.createQueryClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	params, err := val1Client.QueryParams()
	if err != nil {
		return err
	}

	denomAcceptList := []string{}
	for _, acceptItem := range params.AcceptList {
		denomAcceptList = append(denomAcceptList, strings.ToUpper(acceptItem.SymbolDenom))
	}

	chainHeight, err := NewChainHeight(ctx, val1Client.clientContext.Client, zerolog.Nop())
	if err != nil {
		return err
	}

	fmt.Println("waiting for exchange rates")
	for i := 0; i < 20; i++ {
		<-chainHeight.HeightChanged
		exchangeRates, err := val1Client.QueryExchangeRates()
		if err == nil && len(exchangeRates) == len(denomAcceptList) {
			break
		}
	}

	priceStore, err := listenForPrices(val1Client, params, chainHeight)
	if err != nil {
		return err
	}
	err = priceStore.checkMedians()
	if err != nil {
		return err
	}

	return nil
}

func listenForPrices(
	umeeClient *UmeeClient,
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
			exchangeRates, err := umeeClient.QueryExchangeRates()
			fmt.Printf("rates at block %d: %+v\n", height, exchangeRates)
			if err != nil {
				return nil, err
			}
			for _, rate := range exchangeRates {
				priceStore.addStamp(rate.Denom, rate.Amount)
			}
		}
	}

	medians, err := umeeClient.QueryMedians()
	if err != nil {
		return nil, err
	}
	for _, median := range medians {
		priceStore.medians[median.Denom] = median.Amount
	}
	return priceStore, nil
}

func isPeriodFirstBlock(height int64, blocksPerPeriod uint64) bool {
	return uint64(height)%blocksPerPeriod == 0
}
