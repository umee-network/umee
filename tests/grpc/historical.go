package grpc

import (
	"context"
	"strings"

	"github.com/rs/zerolog"
)

// MedianCheck waits for availability of all exchange rates from the denom accept list,
// records historical stamp data based on the oracle params, computes the
// median/median deviation and then compares that to the data in the
// median/median deviation gRPC query
func MedianCheck(val1Client *UmeeClient) error {
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

	for i := 0; i < 20; i++ {
		exchangeRates, err := val1Client.QueryExchangeRates()
		if err == nil && len(exchangeRates) == len(denomAcceptList) {
			break
		}
		<-chainHeight.HeightChanged
	}

	priceStore, err := listenForPrices(val1Client, params, chainHeight)
	if err != nil {
		return err
	}
	err = priceStore.checkMedians()
	if err != nil {
		return err
	}
	err = priceStore.checkMedianDeviations()
	if err != nil {
		return err
	}

	return nil
}
