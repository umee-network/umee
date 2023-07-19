package bpmath

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// Tests if it works both with sdk.Int and math.Int
func TestInt(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	bp := BP(100)
	si := sdk.NewInt(1234)

	var sresult sdk.Int = Mul(si, bp)
	require.Equal(sresult, sdk.NewInt(12))
	var mresult math.Int = Mul(si, bp)
	require.Equal(mresult, math.NewInt(12))

	// now let's check math.Int
	mi := math.NewInt(1234)
	sresult = Mul(mi, bp)
	require.Equal(sresult, sdk.NewInt(12))
	mresult = Mul(mi, bp)
	require.Equal(mresult, math.NewInt(12))

	// test rounding
	si = sdk.NewInt(1299)
	require.Equal(bp.Mul(si), sdk.NewInt(12))

	si = sdk.NewInt(-1299)
	require.Equal(bp.Mul(si), sdk.NewInt(-12))

	si = sdk.NewInt(-1201)
	require.Equal(bp.Mul(si), sdk.NewInt(-12))
}
