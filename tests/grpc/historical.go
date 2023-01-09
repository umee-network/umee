package grpc

import (
	"context"
	"fmt"
	"strings"

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
	fmt.Printf("%+v\n", params)

	denomAcceptList := []string{}
	for _, acceptItem := range params.AcceptList {
		denomAcceptList = append(denomAcceptList, strings.ToUpper(acceptItem.SymbolDenom))
	}

	chainHeight, err := NewChainHeight(ctx, val1Client.clientContext.Client, zerolog.Nop())
	if err != nil {
		return err
	}

	for i := 0; i < 20; i++ {
		<-chainHeight.HeightChanged
		exchangeRates, err := val1Client.QueryExchangeRates()
		if err == nil && len(exchangeRates) > 0 {
			break
		}
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
		// err = priceStore.checkMedianDeviation()
		// if err != nil {
		// 	return err
		// }
	}
	return nil
}
