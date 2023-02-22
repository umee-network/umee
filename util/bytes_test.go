package util

import (
	"encoding/binary"
	"math"
	"testing"

	"gotest.tools/v3/assert"
)

func TestMergeBytes(t *testing.T) {
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
	for _, tc := range tcs {
		assert.DeepEqual(t, tc.out, ConcatBytes(tc.inMargin, tc.in...))
	}
}

func TestUintWithNullPrefix(t *testing.T) {
	expected := []byte{0}
	num := make([]byte, 8)
	binary.LittleEndian.PutUint64(num, math.MaxUint64)
	expected = append(expected, num...)

	out := UintWithNullPrefix(math.MaxUint64)
	assert.DeepEqual(t, expected, out)
}
