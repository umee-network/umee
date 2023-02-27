package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tassert "github.com/stretchr/testify/assert"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v4/tests/util"
	"github.com/umee-network/umee/v4/x/oracle/types"
)

func TestAvgKeeper(t *testing.T) {
	t.Parallel()

	s := AvgKeeperSuite{denom1: "denom1", denom2: "denom2"}
	t.Run("new counters", s.testNewCounters)
	t.Run("setAvgCounters", s.testSetAvgCounters)
	t.Run("GetCurrentAvg", s.testGetCurrentAvg)
	t.Run("UpdateAvgCounter", s.testUpdateAvgCounter)
}

type AvgKeeperSuite struct {
	denom1 string
	denom2 string
}

func (s AvgKeeperSuite) newAvgKeeper(t *testing.T, period, shift time.Duration) AvgKeeper {
	db := util.KVStore(t)
	return AvgKeeper{store: db, period: period, shift: shift}
}

func (s AvgKeeperSuite) newDefAvgKeeper(t *testing.T) AvgKeeper {
	return s.newAvgKeeper(t, AvgPeriod, AvgShift)
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
		k := s.newAvgKeeper(t, tc.period, tc.shift)
		assert.DeepEqual(t, k.newCounters(now), tc.expected)
	}
}

func (s AvgKeeperSuite) testSetAvgCounters(t *testing.T) {
	start := time.Now()
	acs := []types.AvgCounter{
		newAvgCounter(10, 20, start),
		newAvgCounter(30, 31, start),
		newAvgCounter(40, 41, start),
	}
	k := s.newDefAvgKeeper(t)

	ret := k.getAllAvgCounters(s.denom1)
	assert.Equal(t, len(ret), 0, ret)

	k.setAvgCounters(s.denom1, acs)
	ret = k.getAllAvgCounters(s.denom1)
	assert.DeepEqual(t, acs, ret)

	k.setAvgCounters(s.denom2, acs[1:])
	ret = k.getAllAvgCounters(s.denom2)
	assert.DeepEqual(t, acs[1:], ret)

	// check that s.denom1 was not ovewritten:
	ret = k.getAllAvgCounters(s.denom1)
	assert.DeepEqual(t, acs, ret)
}

func (s AvgKeeperSuite) testLatestIndx(t *testing.T) {
	k := s.newDefAvgKeeper(t)
	_, err := k.getLatestIdx(s.denom1)
	assert.Equal(t, err, types.ErrNoLatestAvgPrice)

	k.setLatestIdx(s.denom1, 3)
	k.setLatestIdx(s.denom2, 4)

	i, err := k.getLatestIdx(s.denom1)
	assert.NilError(t, err)
	assert.Equal(t, i, 3)

	// check that denom1 and denom2 don't conflict
	i, err = k.getLatestIdx(s.denom2)
	assert.NilError(t, err)
	assert.Equal(t, i, 4)

}

func (s AvgKeeperSuite) testGetCurrentAvg(t *testing.T) {
	k := s.newDefAvgKeeper(t)

	// with no latest index, zero should be returned
	v, err := k.GetCurrentAvg(s.denom1)
	assert.NilError(t, err)
	assert.DeepEqual(t, v, sdk.ZeroDec())
}

func (s AvgKeeperSuite) testUpdateAvgCounter(t *testing.T) {
	const period = time.Hour * 16
	const shift = time.Hour * 2

	now := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	k := s.newAvgKeeper(t, period, shift)

	k.updateAvgCounter(s.denom1, sdk.NewDec(1), now)
	checkAvgPrice(t, k, "1", s.denom1)

	// k.updateAvgCounter(s.denom1, sdk.NewDec(1), now.Add(time.Minute))
	// checkAvgPrice(t, k, "1", s.denom1)

	// k.updateAvgCounter(s.denom1, sdk.NewDec(4), now.Add(time.Minute*2))
	// checkAvgPrice(t, k, "2", s.denom1)

	// check that avg denoms don't conflict
	k.updateAvgCounter(s.denom2, sdk.NewDec(7), now)
	checkAvgPrice(t, k, "7", s.denom2)
	checkAvgPrice(t, k, "1", s.denom1)
}

func checkAvgPrice(t *testing.T, k AvgKeeper, expected, denom string) {
	expectedDec := sdk.MustNewDecFromStr(expected)
	v, err := k.GetCurrentAvg(denom)
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedDec, v)
	tassert.Equal(t, expectedDec, v)
}

func newAvgCounter(sum, num uint32, start time.Time) types.AvgCounter {
	return types.AvgCounter{Sum: sdk.NewDec(int64(sum)), Num: num, Start: start}
}
