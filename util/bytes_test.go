package util

import "testing"
import "github.com/stretchr/testify/require"

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
