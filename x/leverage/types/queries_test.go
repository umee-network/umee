package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryMaxWithdraw(t *testing.T) {
	require := require.New(t)
	tcs := []struct {
		name string
		q    QueryMaxWithdraw
		err  string
	}{
		{"no address", QueryMaxWithdraw{}, "empty address"},
	}
	for _, tc := range tcs {
		err := tc.q.ValidateBasic()
		if tc.err == "" {
			require.NoError(err, tc.name)
		} else {
			require.ErrorContains(err, tc.err, tc.name)
		}
	}
}
