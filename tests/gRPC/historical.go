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

	chainHeight, err := NewChainHeight(
		ctx,
		val1Client.clientContext.Client,
		zerolog.Nop(),
		2, // why pass in the initial block height?
	)
	if err != nil {
		cancel()
		return err
	}

	for {
		height := <-chainHeight.heightChanged
		fmt.Printf("height: %d\n", height)
		if err != nil {
			cancel()
			return err
		}
		if height > 60 {
			break
		}
	}

	params, err := val1Client.QueryParams()
	fmt.Printf("%+v\n", params)
	if err != nil {
		cancel()
		return err
	}

	// votePeriod := params.VotePeriod

	exchangeRates, err := val1Client.QueryMedians()
	fmt.Printf("%+v\n", exchangeRates)
	if err != nil {
		cancel()
		return err
	}

	medians, err := val1Client.QueryMedians()
	fmt.Printf("%+v\n", medians)
	if err != nil {
		cancel()
		return err
	}

	// calcMedian, err := decmath.Median([]sdk.Dec{price1, price2, price3})
	// if err != nil {
	// 	return err
	// }
	// if median != calcMedian {
	// 	return fmt.Errorf("Expected %d for the median but got %d")
	// }

	cancel()
	return nil
}
