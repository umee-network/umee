package bpmath

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestFixedQuo(t *testing.T) {
	tcs := []struct {
		name string
		a    uint64
		b    uint64
		r    Rounding
		exp  FixedBP
	}{
		{"t1", 0, 0, UP, ONE},
		{"t2", 0, 0, DOWN, ONE},
		{"t3", 1, 0, UP, ONE},
		{"t4", 1, 0, DOWN, ONE},

		{"t5", 20, 10, UP, ONE},
		{"t6", 20, 10, DOWN, ONE},
		{"t7", 20, 20, UP, ONE},
		{"t7-1", 20, 20, DOWN, ONE},

		{"t8", 1, 2, UP, ONE / 2},
		{"t9", 1, 2, DOWN, ONE / 2},
		{"t10", 1, 3, UP, 3334},
		{"t11", 1, 3, DOWN, 3333},
		{"t12", 2, 3, UP, 6667},
		{"t13", 2, 3, DOWN, 6666},
		{"t14", 10, 99999, UP, 2},
		{"t15", 10, 99999, DOWN, 1},
		{"t16", 10, 99999999, UP, 1},
		{"t17", 10, 99999999, DOWN, 0},
	}
	require := require.New(t)
	for _, tc := range tcs {
		a, b := sdk.NewIntFromUint64(tc.a), sdk.NewIntFromUint64(tc.b)
		o := FixedQuo(a, b, tc.r)
		require.Equal(int(tc.exp), int(o), fmt.Sprint("test", tc.name))
	}
}
