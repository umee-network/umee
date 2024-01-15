package grpc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/util/decmath"
)

type PriceStore struct {
	historicStamps   map[string][]sdkmath.LegacyDec
	medians          map[string]sdkmath.LegacyDec
	medianDeviations map[string]sdkmath.LegacyDec
}

func NewPriceStore() *PriceStore {
	return &PriceStore{
		historicStamps:   map[string][]sdkmath.LegacyDec{},
		medians:          map[string]sdkmath.LegacyDec{},
		medianDeviations: map[string]sdkmath.LegacyDec{},
	}
}

func (ps *PriceStore) addStamp(denom string, stamp sdkmath.LegacyDec) {
	if _, ok := ps.historicStamps[denom]; !ok {
		ps.historicStamps[denom] = []sdkmath.LegacyDec{}
	}
	ps.historicStamps[denom] = append(ps.historicStamps[denom], stamp)
}

func (ps *PriceStore) checkMedians() error {
	for denom, stamps := range ps.historicStamps {
		calcMedian, err := decmath.Median(stamps)
		if err != nil {
			return err
		}
		if !ps.medians[denom].Equal(calcMedian) {
			return fmt.Errorf(
				"expected %d for the %s median but got %d",
				ps.medians[denom],
				denom,
				calcMedian,
			)
		}
	}
	return nil
}

func (ps *PriceStore) checkMedianDeviations() error {
	for denom, median := range ps.medians {
		calcMedianDeviation, err := decmath.MedianDeviation(median, ps.historicStamps[denom])
		if err != nil {
			return err
		}
		if !ps.medianDeviations[denom].Equal(calcMedianDeviation) {
			return fmt.Errorf(
				"expected %d for the %s median deviation but got %d",
				ps.medianDeviations[denom],
				denom,
				calcMedianDeviation,
			)
		}
	}
	return nil
}
