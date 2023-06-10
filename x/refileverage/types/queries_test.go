package types

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestQueryMaxWithdraw(t *testing.T) {
	tcs := []struct {
		name string
		q    QueryMaxWithdraw
		err  string
	}{
		{"no address", QueryMaxWithdraw{}, "empty address"},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.q.ValidateBasic()
			if tc.err == "" {
				assert.NilError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}
