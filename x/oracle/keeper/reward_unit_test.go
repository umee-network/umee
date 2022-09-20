package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVoteTargetWithUmee(t *testing.T) {
	require := require.New(t)
	tcs := []struct {
		in  []string
		out []string
	}{
		{[]string{}, []string{"uumee"}},
		{[]string{"a"}, []string{"uumee", "a"}},
		{[]string{"x", "a", "heeeyyy"}, []string{"uumee", "x", "a", "heeeyyy"}},
		{[]string{"x", "a", "uumee"}, []string{"x", "a", "uumee"}},
	}
	for i, tc := range tcs {
		require.Equal(tc.out, voteTargetsWithUmee(tc.in), i)
	}

}
