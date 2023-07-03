package grpc

import (
	"errors"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/umee-network/umee/v5/client"
)

// MedianCheck waits for availability of all exchange rates from the denom accept list,
// records historical stamp data based on the oracle params, computes the
// median/median deviation and then compares that to the data in the
// median/median deviation gRPC query
func MedianCheck(umee client.Client) error {
	ctx, cancel := umee.NewQCtxWithCancel()
	defer cancel()

	params, err := umee.QueryOracleParams()
	if err != nil {
		return err
	}

	denomAcceptList := []string{}
	for _, acceptItem := range params.AcceptList {
		denomAcceptList = append(denomAcceptList, strings.ToUpper(acceptItem.SymbolDenom))
	}

	chainHeight, err := umee.NewChainHeightListener(ctx, zerolog.Nop())
	if err != nil {
		return err
	}

	var exchangeRates sdk.DecCoins
	for i := 0; i < 50; i++ {
		exchangeRates, err = umee.QueryExchangeRates()
		if err == nil && len(exchangeRates) == len(denomAcceptList) {
			break
		}
		<-chainHeight.HeightChanged
	}
	// error if the loop above didn't succeed
	if err != nil {
		return err
	}
	if len(exchangeRates) != len(denomAcceptList) {
		return errors.New("couldn't fetch exchange rates matching denom accept list")
	}

	priceStore, err := listenForPrices(umee, params, chainHeight)
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
