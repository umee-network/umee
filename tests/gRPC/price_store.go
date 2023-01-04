package gRPC

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PriceStore struct {
	historicStamps map[int64]map[string]sdk.Dec
	medians        map[int64]map[string]sdk.Dec

	historicStampMtx sync.RWMutex
	medianMtx        sync.RWMutex
}

func (ps *PriceStore) GetHistoricStamps() map[int64]map[string]sdk.Dec {
	ps.historicStampMtx.RLock()
	defer ps.historicStampMtx.RUnlock()

	return ps.historicStamps
}

func (ps *PriceStore) GetMedians() map[int64]map[string]sdk.Dec {
	ps.medianMtx.RLock()
	defer ps.medianMtx.RUnlock()

	return ps.medians
}

func (ps *PriceStore) SetHistoricStamp(blockNum int64, denom string, price sdk.Dec) {
	ps.historicStampMtx.Lock()
	defer ps.historicStampMtx.Unlock()

	if _, ok := ps.historicStamps[blockNum]; !ok {
		ps.historicStamps[blockNum] = map[string]sdk.Dec{}
	}

	ps.historicStamps[blockNum][denom] = price
}

func (ps *PriceStore) SetMedian(blockNum int64, denom string, price sdk.Dec) {
	ps.medianMtx.Lock()
	defer ps.medianMtx.Unlock()

	if _, ok := ps.medians[blockNum]; !ok {
		ps.medians[blockNum] = map[string]sdk.Dec{}
	}

	ps.medians[blockNum][denom] = price
}
