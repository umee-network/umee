package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util"
	"github.com/umee-network/umee/v4/x/oracle/types"
)

type AvgKeeper struct {
	store  sdk.KVStore
	period time.Duration
	shift  time.Duration
}

func (k Keeper) AvgKeeper(ctx sdk.Context) AvgKeeper {
	return AvgKeeper{store: ctx.KVStore(k.storeKey), period: AvgPeriod, shift: AvgShift}
}

func (k AvgKeeper) numCounters() int64 {
	return int64(k.period) / int64(k.shift)
}

func (k AvgKeeper) newCounters(start time.Time) []types.AvgCounter {
	num := k.numCounters()
	acs := make([]types.AvgCounter, num)
	for i := int64(0); i < num; i++ {
		acs[i].Start = start
		acs[i].Sum = sdk.ZeroDec()
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

	for i := range acs {
		a := &acs[i]
		// initialization: in the first run, we will have all by one Starts after "now"
		if now.Before(a.Start) {
			continue
		}

		t := a.Start.Add(k.period)
		if now.Before(t) {
			a.Sum = a.Sum.Add(exchangeRate)
			a.Num++
		} else {
			a.Sum = exchangeRate
			a.Num = 1
			a.Start = t
			for t = t.Add(k.period); !now.Before(t); t = t.Add(k.period) {
				a.Start = t
			}
		}
	}
	k.setAvgCounters(denom, acs)

	// find the oldest "Start", need to do it in a separate loop, because in the loop above
	// we update "Start"
	// current counter is the one which most of the aggregated prices,
	// so the one with the oldest Start (unless it cycled over)
	oldestCounter := -1
	oldest := acs[0].Start
	for i := range acs {
		// Can't use `Before` to handle the case where there are no prices
		if !acs[i].Start.After(oldest) {
			oldest = acs[i].Start
			oldestCounter = i
		}
	}
	k.setLatestIdx(denom, byte(oldestCounter))
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

	iter := sdk.KVStorePrefixIterator(k.store, prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var av types.AvgCounter
		err := av.Unmarshal(iter.Value())
		if err != nil {
			panic(err)
		}
		avs = append(avs, av)
	}

	return avs
}

// setAvgCounters sets AllAvgCounter in the same order as in the slice.
// Contract: MUST be the same order as returned from GetAllAvgCounters
func (k AvgKeeper) setAvgCounters(denom string, acs []types.AvgCounter) {
	for i := range acs {
		key := types.KeyAvgCounter(denom, byte(i))
		bz, err := acs[i].Marshal()
		if err != nil {
			panic(err)
		}
		// bz := k.cdc.MustMarshal(&acs[i])
		k.store.Set(key, bz)
	}
}

func (k AvgKeeper) GetCurrentAvg(denom string) (sdk.Dec, error) {
	latestIdx, err := k.getLatestIdx(denom)
	if err == types.ErrNoLatestAvgPrice {
		return sdk.ZeroDec(), nil
	}
	if err != nil {
		return sdk.Dec{}, err
	}
	av, err := k.getCounter(denom, latestIdx)
	if err != nil {
		return sdk.Dec{}, nil
	}

	return av.Sum.Quo(sdk.NewDec(int64(av.Num))), nil
}

func (k AvgKeeper) getCounter(denom string, idx byte) (types.AvgCounter, error) {
	key := types.KeyAvgCounter(denom, idx)
	bz := k.store.Get(key)
	if len(bz) == 0 {
		return types.AvgCounter{}, types.ErrNoLatestAvgPrice
	}
	var av types.AvgCounter
	return av, av.Unmarshal(bz)
}
