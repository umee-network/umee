package bpmath

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestBPToDec(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name string
		a    FixedBP
		exp  math.LegacyDec
	}{
		{"t1", 99999, math.LegacyMustNewDecFromStr("9.9999")},
		{"t2", One * 10, math.LegacyMustNewDecFromStr("10.0")},
	}
	require := require.New(t)
	for _, tc := range tcs {
		bp := BP(tc.a).ToDec()
		require.Equal(tc.exp.String(), bp.String(), fmt.Sprint("test-bp ", tc.name))
	}
}
