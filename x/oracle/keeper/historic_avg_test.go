package keeper

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/store"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v4/tests/util"
	"github.com/umee-network/umee/v4/x/oracle/types"
)

func TestAvgKeeper(t *testing.T) {
	t.Parallel()

	db := util.KVStore(t)
	s := AvgKeeperSuite{store: db}

	t.Run("new counters", s.testNewCounters)
	// t.Run("another scenario, s.testScenario2")
}

type AvgKeeperSuite struct {
	store store.KVStore
}

func (s AvgKeeperSuite) newAvgKeeper(period, shift time.Duration) AvgKeeper {
	return AvgKeeper{store: s.store, period: period, shift: shift}
}

func (s AvgKeeperSuite) testNewCounters(t *testing.T) {
	allCounters := []types.AvgCounter{}
	now := time.Now()
	shift := time.Second * 10
	for i := time.Duration(0); i < 100; i++ {
		allCounters = append(allCounters, types.AvgCounter{Start: now.Add(shift * i)})
	}
	tcs := []struct {
		name     string
		period   time.Duration
		shift    time.Duration
		expected []types.AvgCounter
	}{
		{"period = shift",
			shift, shift, allCounters[:1]},
		{"period = 2*shift",
			shift * 2, shift, allCounters[:2]},
		{"period = 2.5*shift",
			shift*2 + shift/2, shift, allCounters[:2]},
	}
	for _, tc := range tcs {
		k := s.newAvgKeeper(tc.period, tc.shift)
		assert.DeepEqual(t, k.newCounters(now), tc.expected)
	}
}
