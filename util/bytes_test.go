package util

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeBytes(t *testing.T) {
	require := require.New(t)
	tcs := []struct {
		in       [][]byte
		inMargin int
		out      []byte
	}{
		{[][]byte{}, 0, []byte{}},
		{[][]byte{}, 2, []byte{0, 0}},
		{[][]byte{{1}}, 0, []byte{1}},
		{[][]byte{{1}}, 1, []byte{1, 0}},
		{[][]byte{{1, 2}, {2}}, 0, []byte{1, 2, 2}},
		{[][]byte{{1, 2}, {2}}, 3, []byte{1, 2, 2, 0, 0, 0}},
		{[][]byte{{1, 2}, {2}, {3, 3}, {4}}, 1, []byte{1, 2, 2, 3, 3, 4, 0}},
	}
	for i, tc := range tcs {
		require.Equal(tc.out, ConcatBytes(tc.inMargin, tc.in...), i)
	}
}

func TestUintWithNullPrefix(t *testing.T) {
	expected := []byte{0}
	num := make([]byte, 8)
	binary.LittleEndian.PutUint64(num, math.MaxUint64)
	expected = append(expected, num...)

	out := UintWithNullPrefix(math.MaxUint64)
	require.Equal(t, expected, out)
}
