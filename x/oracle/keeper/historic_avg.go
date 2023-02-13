package keeper

import (
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util"
	"github.com/umee-network/umee/v4/x/oracle/types"
)

type AvgKeeper struct {
	cdc    codec.BinaryCodec
	store  sdk.KVStore
	period time.Duration
	shift  time.Duration
}

func (k AvgKeeper) newCounters(start time.Time) []types.AvgCounter {
	num := int64(k.period) / int64(k.shift)
	acs := make([]types.AvgCounter, num)
	for i := int64(0); i < num; i++ {
		acs[i].Start = start
		start = start.Add(k.shift)
	}
	return acs
}

// updateAvgCounter fetches avg counters from the store and adds new exchange exchange rate
// into the aggregate
func (k AvgKeeper) updateAvgCounter(
	denom string,
	exchangeRate sdk.Dec,
	now time.Time,
) {
	acs := k.getAllAvgCounters(denom)
	// if there are no counters registered, we need to initialize them
	if len(acs) == 0 {
		acs = k.newCounters(now)
	}

	// current counter is the one which most of the aggregated prices,
	// so the one with the oldest Start (unless it cycled over)
	currentCounter := -1
	oldest := acs[0].Start
	for i := range acs {
		a := &acs[i]
		// initialization: in the first run, we will have all by one Starts after "now"
		if a.Start.After(now) {
			continue
		}

		// TODO: update the algorithm to handle a chain halt scenario
		// https://linear.app/umee/issue/UMEE-308/

		t := a.Start.Add(k.period)
		if t.After(now) {
			a.Sum = exchangeRate
			a.Num = 1
			a.Start = now
		} else {
			a.Sum = a.Sum.Add(exchangeRate)
			a.Num++
		}
		// Can't use `Before` to handle the case where there are no prices
		if !a.Start.After(oldest) {
			oldest = t
			currentCounter = i
		}
	}
	k.setAvgCounters(denom, acs)

	if currentCounter >= 0 {
		k.setLatestIdx(denom, byte(currentCounter))
	}
}

func (k AvgKeeper) getLatestIdx(denom string) (byte, error) {
	bz := k.store.Get(k.latestIdxKey(denom))
	if len(bz) == 0 {
		return 0, types.ErrNoLatestAvgPrice
	}
	if len(bz) != 1 {
		return 0, types.ErrMalformedLatestAvgPrice
	}
	return bz[0], nil
}

func (k AvgKeeper) setLatestIdx(denom string, idx byte) {
	k.store.Set(k.latestIdxKey(denom), []byte{idx})
}

func (k AvgKeeper) latestIdxKey(denom string) []byte {
	return append(types.KeyLatestAvgCounter, []byte(denom)...)
}

func (k AvgKeeper) getAllAvgCounters(denom string) []types.AvgCounter {
	avs := make([]types.AvgCounter, 0)
	prefix := util.ConcatBytes(0, types.KeyPrefixAvgCounter, []byte(denom))
	if !k.store.Has(prefix) {
		return avs
	}

	iter := sdk.KVStorePrefixIterator(k.store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var av types.AvgCounter
		k.cdc.MustUnmarshal(iter.Value(), &av) // SYGET
		avs = append(avs, av)
	}

	return avs
}

// setAvgCounters sets AllAvgCounter in the same order as in the slice.
// Contract: MUST be the same order as returned from GetAllAvgCounters
func (k AvgKeeper) setAvgCounters(denom string, acs []types.AvgCounter) {
	denom = strings.ToUpper(denom) // setters enforce uppercase symbol denom
	key := types.KeyAvgCounter(denom, 0)
	lastIdx := len(key) - 1
	for i := range acs {
		key[lastIdx] = byte(i)
		bz := k.cdc.MustMarshal(&acs[i])
		k.store.Set(key, bz) // SYSET
	}
}

func (k AvgKeeper) GetCurrentAvg(denom string) (sdk.Dec, error) {
	latestIdx, err := k.getLatestIdx(denom)
	if err != nil {
		return sdk.Dec{}, err
	}

	key := types.KeyAvgCounter(denom, latestIdx)
	var av types.AvgCounter
	bz := k.store.Get(key) // SYGET
	if len(bz) == 0 {
		return sdk.Dec{}, types.ErrNoLatestAvgPrice
	}
	k.cdc.MustUnmarshal(bz, &av)

	return av.Sum.Quo(sdk.NewDec(int64(av.Num))), nil
}
