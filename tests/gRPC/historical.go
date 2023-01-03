package gRPC

import (
	"fmt"
	"time"
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
	val1Client.createQueryClient()

	for range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {

		exchangeRates, err := val1Client.QueryMedians()
		fmt.Printf("%+v\n", exchangeRates)
		if err != nil {
			return err
		}

		medians, err := val1Client.QueryMedians()
		fmt.Printf("%+v\n", medians)
		if err != nil {
			return err
		}
		time.Sleep(2 * time.Minute)

	}

	// calcMedian, err := decmath.Median([]sdk.Dec{price1, price2, price3})
	// if err != nil {
	// 	return err
	// }
	// if median != calcMedian {
	// 	return fmt.Errorf("Expected %d for the median but got %d")
	// }

	return nil
}
