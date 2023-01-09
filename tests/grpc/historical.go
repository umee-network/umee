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

	// Wait 10 blocks for price feeder prices
	for i := 0; i < 20; i++ {
		<-chainHeight.HeightChanged
	}

	for _, denom := range denomAcceptList {
		priceStore := &PriceStore{}
		err = listenForPrices(val1Client, params, denom, priceStore, chainHeight)
		if err != nil {
			return err
		}
		err = priceStore.checkMedian()
		if err != nil {
			return err
		}
	}
	return nil
}

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

	// Wait one more block for the median
	<-chainHeight.HeightChanged
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
