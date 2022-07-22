package bpmath

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestBPToDec(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name string
		a    FixedBP
		exp  sdk.Dec
	}{
		{"t1", 99999, sdk.MustNewDecFromStr("9.9999")},
		{"t2", ONE * 10, sdk.MustNewDecFromStr("10.0")},
	}
	require := require.New(t)
	for _, tc := range tcs {
		bp := BP(tc.a).ToDec()
		require.Equal(tc.exp.String(), bp.String(), fmt.Sprint("test-bp ", tc.name))
	}
}
