package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tassert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	t.Run("UpdateAvgCounterSimple", s.testUpdateAvgCounterSimple)
	t.Run("UpdateAvgCounterShift", s.testUpdateAvgCounterShift)
	t.Run("UpdateAvgCounterCycle", s.testUpdateAvgCounterCycle)
	t.Run("UpdateAvgCounterHalt", s.testUpdateAvgCounterHalt)
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
	return s.newAvgKeeper(t, defaultAvgPeriod, defaultAvgShift)
}

func (s AvgKeeperSuite) testNewCounters(t *testing.T) {
	allCounters := []types.AvgCounter{}
	now := time.Now()
	shift := time.Second * 10
	for i := time.Duration(0); i < 100; i++ {
		allCounters = append(allCounters, types.AvgCounter{Start: now.Add(shift * i), Sum: sdk.ZeroDec()})
	}
	tcs := []struct {
		name     string
		period   time.Duration
		shift    time.Duration
		expected []types.AvgCounter
	}{
		{
			"period = shift",
			shift, shift, allCounters[:1],
		},
		{
			"period = 2*shift",
			shift * 2, shift, allCounters[:2],
		},
		{
			"period = 2.5*shift",
			shift*2 + shift/2, shift, allCounters[:2],
		},
	}
	for _, tc := range tcs {
		k := s.newAvgKeeper(t, tc.period, tc.shift)
		require.Equal(t, tc.expected, k.newCounters(now), tc.name)
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

func (s AvgKeeperSuite) setupUpdateAvgCounter(t *testing.T) (time.Time, time.Duration, AvgKeeper) {
	const period = time.Hour * 16
	const shift = time.Hour * 2

	now := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	k := s.newAvgKeeper(t, period, shift)
	return now, shift, k
}

func (s AvgKeeperSuite) testUpdateAvgCounterSimple(t *testing.T) {
	now, _, k := s.setupUpdateAvgCounter(t)

	k.updateAvgCounter(s.denom1, sdk.NewDec(1), now)
	checkAvgPrice(t, k, "1", s.denom1, 0)

	k.updateAvgCounter(s.denom1, sdk.NewDec(2), now)
	checkAvgPrice(t, k, "1.5", s.denom1, 0)

	k.updateAvgCounter(s.denom1, sdk.NewDec(6), now.Add(time.Minute*2))
	checkAvgPrice(t, k, "3", s.denom1, 0) // num: 3, sum: 9

	// check that avg denoms don't conflict
	k.updateAvgCounter(s.denom2, sdk.NewDec(7), now)
	checkAvgPrice(t, k, "7", s.denom2, 0)
	checkAvgPrice(t, k, "3", s.denom1, 0)
}

func (s AvgKeeperSuite) testUpdateAvgCounterShift(t *testing.T) {
	now, shift, k := s.setupUpdateAvgCounter(t)
	numCounters := k.numCounters()

	k.updateAvgCounter(s.denom1, sdk.NewDec(1), now)
	checkAvgPrice(t, k, "1", s.denom1, 0)
	for i := 1; i < int(numCounters); i++ {
		checkCounter(t, k, s.denom1, byte(i), "0", 0)
	}

	// new shift:
	now = now.Add(shift)
	k.updateAvgCounter(s.denom1, sdk.NewDec(5), now)
	checkAvgPrice(t, k, "3", s.denom1, 0) // {num: 2, sum: 6}
	checkCounter(t, k, s.denom1, 1, "5", 1)
	for i := 2; i < int(numCounters); i++ {
		checkCounter(t, k, s.denom1, byte(i), "0", 0)
	}

	// new shift:
	now = now.Add(shift)
	k.updateAvgCounter(s.denom1, sdk.NewDec(6), now)
	checkAvgPrice(t, k, "4", s.denom1, 0) // {num: 3, sum: 9}
	checkCounter(t, k, s.denom1, 1, "11", 2)
	checkCounter(t, k, s.denom1, 2, "6", 1)
	for i := 3; i < int(numCounters); i++ {
		checkCounter(t, k, s.denom1, byte(i), "0", 0)
	}
}

func (s AvgKeeperSuite) testUpdateAvgCounterCycle(t *testing.T) {
	now, shift, k := s.setupUpdateAvgCounter(t)
	k.updateAvgCounter(s.denom1, sdk.NewDec(1), now)
	checkAvgPrice(t, k, "1", s.denom1, 0)

	numCounters := k.numCounters()
	// go to the latest shift in the epoch
	now = now.Add(shift * time.Duration(numCounters-1))
	k.updateAvgCounter(s.denom1, sdk.NewDec(3), now)
	checkAvgPrice(t, k, "2", s.denom1, 0)
	for i := 1; i < int(numCounters); i++ {
		checkCounter(t, k, s.denom1, byte(i), "3", 1)
	}

	// cycle over by 2 shifts -> going to index=1
	// after the cycle, the current index should be the one above it -> 2
	now = now.Add(shift * 2)
	k.updateAvgCounter(s.denom1, sdk.NewDec(2), now)
	checkCounter(t, k, s.denom1, 0, "2", 1)
	checkCounter(t, k, s.denom1, 1, "2", 1)
	checkCounter(t, k, s.denom1, 2, "5", 2) // num=2 because the very first update was only registered for counter[0]
	for i := 3; i < int(numCounters); i++ {
		checkCounter(t, k, s.denom1, byte(i), "5", 2)
	}
	checkAvgPrice(t, k, "2.5", s.denom1, 2)

	now = now.Add(time.Minute)
	k.updateAvgCounter(s.denom1, sdk.NewDec(1), now)
	checkAvgPrice(t, k, "2", s.denom1, 2)

	now = now.Add(k.shift)
	k.updateAvgCounter(s.denom1, sdk.NewDec(2), now)
	checkAvgPrice(t, k, "2", s.denom1, 3)
	for i := 3; i < int(numCounters); i++ {
		checkCounter(t, k, s.denom1, byte(i), "8", 4)
	}
	checkCounter(t, k, s.denom1, 2, "2", 1)
	checkCounter(t, k, s.denom1, 1, "5", 3)
	checkCounter(t, k, s.denom1, 0, "5", 3)
}

func (s AvgKeeperSuite) testUpdateAvgCounterHalt(t *testing.T) {
	now, shift, k := s.setupUpdateAvgCounter(t)
	k.updateAvgCounter(s.denom1, sdk.NewDec(1), now)
	checkAvgPrice(t, k, "1", s.denom1, 0)

	numCounters := k.numCounters()
	// go to the latest shift in the epoch
	now = now.Add(shift * time.Duration(numCounters-1))
	k.updateAvgCounter(s.denom1, sdk.NewDec(3), now)
	checkAvgPrice(t, k, "2", s.denom1, 0)

	// go 2 periods forward
	now = now.Add(shift * time.Duration(numCounters))
	k.updateAvgCounter(s.denom1, sdk.NewDec(6), now)
	checkAvgPrice(t, k, "6", s.denom1, 0)

	// go 2 periods -1 shift forward
	now = now.Add(shift * time.Duration(2*numCounters-1))
	k.updateAvgCounter(s.denom1, sdk.NewDec(7), now)
	checkAvgPrice(t, k, "7", s.denom1, byte(numCounters)-1)
}

func checkAvgPrice(t *testing.T, k AvgKeeper, expected, denom string, idx byte) {
	i, err := k.getLatestIdx(denom)
	if err != types.ErrNoLatestAvgPrice {
		require.NoError(t, err)
	}
	require.Equal(t, idx, i, "index check")

	expectedDec := sdk.MustNewDecFromStr(expected)
	v, err := k.GetCurrentAvg(denom)
	require.NoError(t, err)
	require.Equal(t, expectedDec, v) // ussing testify to have a stack trace
}

func checkCounter(t *testing.T, k AvgKeeper, denom string, idx byte, sum string, num uint32) {
	sumDec := sdk.MustNewDecFromStr(sum)
	c, err := k.getCounter(denom, idx)
	require.NoError(t, err)
	tassert.Equal(t, sumDec, c.Sum, "sum idx=%d", idx)
	tassert.Equal(t, num, c.Num, "num idx=%d", idx)
}

func newAvgCounter(sum, num uint32, start time.Time) types.AvgCounter {
	return types.AvgCounter{Sum: sdk.NewDec(int64(sum)), Num: num, Start: start}
}
