package bpmath

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestFixedQuo(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name   string
		a      uint64
		b      uint64
		r      Rounding
		exp    FixedBP
		panics bool
	}{
		{"t1", 0, 0, UP, ONE, true},
		{"t2", 0, 0, DOWN, ONE, true},
		{"t3", 1, 0, UP, ONE, true},
		{"t4", 1, 0, DOWN, ONE, true},

		{"t5", 20, 10, UP, 0, true},
		{"t6", 20, 10, DOWN, 0, true},
		{"t7", 20, 20, UP, ONE, false},
		{"t7-1", 20, 20, DOWN, ONE, false},

		{"t8", 1, 2, UP, ONE / 2, false},
		{"t9", 1, 2, DOWN, ONE / 2, false},
		{"t10", 1, 3, UP, 3334, false},
		{"t11", 1, 3, DOWN, 3333, false},
		{"t12", 2, 3, UP, 6667, false},
		{"t13", 2, 3, DOWN, 6666, false},
		{"t14", 10, 99999, UP, 2, false},
		{"t15", 10, 99999, DOWN, 1, false},
		{"t16", 10, 99999999, UP, 1, false},
		{"t17", 10, 99999999, DOWN, 0, false},
	}
	require := require.New(t)
	for _, tc := range tcs {
		a, b := math.NewIntFromUint64(tc.a), math.NewIntFromUint64(tc.b)
		if tc.panics {
			require.Panics(func() {
				FixedFromQuo(a, b, tc.r)
			}, tc.name)
			continue
		}
		o := FixedFromQuo(a, b, tc.r)
		require.Equal(int(tc.exp), int(o), fmt.Sprint("test ", tc.name))
	}
}

func TestFixedMul(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name string
		a    uint64
		b    BP
		exp  uint64
	}{
		{"t1", 20, 0, 0},
		{"t2", 20, 1, 0},
		{"t3", 20, ONE, 20},
		{"t4", 20000, 0, 0},
		{"t5", 20000, 1, 2},
		{"t6", 20000, 2, 4},
		{"t7", 20000, half, 10000},
		{"t8", 2000, 4, 0},
		{"t9", 2000, 5, 1},
		{"t10", 2000, half, 1000},
	}
	require := require.New(t)
	for _, tc := range tcs {
		a := math.NewIntFromUint64(tc.a)
		o := Mul(a, tc.b)
		require.Equal(int64(tc.exp), o.Int64(), fmt.Sprint("test ", tc.name))

		// must work with both FixedBP and BP
		o = Mul(a, FixedBP(tc.b))
		require.Equal(int64(tc.exp), o.Int64(), fmt.Sprint("test ", tc.name))

	}
}

func TestFixedToDec(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name string
		a    FixedBP
		exp  math.LegacyDec
	}{
		{"t1", 0, math.LegacyZeroDec()},
		{"t2", 1, math.LegacyMustNewDecFromStr("0.0001")},
		{"t3", 20, math.LegacyMustNewDecFromStr("0.002")},
		{"t4", 9999, math.LegacyMustNewDecFromStr("0.9999")},
		{"t5", ONE, math.LegacyNewDec(1)},
	}
	require := require.New(t)
	for _, tc := range tcs {
		o := tc.a.ToDec()
		require.Equal(tc.exp.String(), o.String(), fmt.Sprint("test-fixedbp ", tc.name))

		bp := BP(tc.a).ToDec()
		require.Equal(tc.exp.String(), bp.String(), fmt.Sprint("test-bp ", tc.name))
	}
}
