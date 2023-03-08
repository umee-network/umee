package keeper

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestPrependUmeeIfUnique(t *testing.T) {
	tcs := []struct {
		in  []string
		out []string
	}{
		// Should prepend "uumee" to a slice of denoms, unless it is already present.
		{[]string{}, []string{"uumee"}},
		{[]string{"a"}, []string{"uumee", "a"}},
		{[]string{"x", "a", "heeeyyy"}, []string{"uumee", "x", "a", "heeeyyy"}},
		{[]string{"x", "a", "uumee"}, []string{"x", "a", "uumee"}},
	}
	for _, tc := range tcs {
		assert.DeepEqual(t, tc.out, prependUmeeIfUnique(tc.in))
	}
}
