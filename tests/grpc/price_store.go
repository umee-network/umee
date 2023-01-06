package grpc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/util/decmath"
)

type PriceStore struct {
	historicStamps []sdk.Dec
	median         sdk.Dec
}

func (ps *PriceStore) checkMedian() error {
	calcMedian, err := decmath.Median(ps.historicStamps)
	if err != nil {
		return err
	}
	if ps.median.Equal(calcMedian) {
		return fmt.Errorf("expected %d for the median but got %d", ps.median, calcMedian)
	}
	return nil
}
