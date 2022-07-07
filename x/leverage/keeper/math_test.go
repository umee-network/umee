package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestInterpolate(t *testing.T) {
	// Define two points (x1,y1) and (x2,y2)
	x1 := sdk.MustNewDecFromStr("3.0")
	x2 := sdk.MustNewDecFromStr("6.0")
	y1 := sdk.MustNewDecFromStr("11.1")
	y2 := sdk.MustNewDecFromStr("17.4")

	// Sloped line, endpoint checks
	x := Interpolate(x1, x1, y1, x2, y2)
	require.Equal(t, x, y1)
	x = Interpolate(x2, x1, y1, x2, y2)
	require.Equal(t, x, y2)

	// Sloped line, point on segment
	x = Interpolate(sdk.MustNewDecFromStr("4.0"), x1, y1, x2, y2)
	require.Equal(t, x, sdk.MustNewDecFromStr("13.2"))

	// Sloped line, point outside of segment
	x = Interpolate(sdk.MustNewDecFromStr("2.0"), x1, y1, x2, y2)
	require.Equal(t, x, sdk.MustNewDecFromStr("9.0"))

	// Vertical line: always return y1
	x = Interpolate(sdk.ZeroDec(), x1, y1, x1, y2)
	require.Equal(t, x, y1)
	x = Interpolate(x1, x1, y1, x1, y2)
	require.Equal(t, x, y1)

	// Undefined line (x1=x2, y1=y2): always return y1
	x = Interpolate(sdk.ZeroDec(), x1, y1, x1, y1)
	require.Equal(t, x, y1)
	x = Interpolate(x1, x1, y1, x1, y1)
	require.Equal(t, x, y1)
}

func TestReduceProportional(t *testing.T) {
	testCase := func(a, b, initial, expected int64) {
		n := sdk.NewInt(initial)
		ReduceProportionally(sdk.NewInt(a), sdk.NewInt(b), &n)
		require.Equal(t, expected, n.Int64())

		m := sdk.NewInt(initial)
		ReduceProportionallyDec(sdk.NewDecFromInt(sdk.NewInt(a)), sdk.NewDecFromInt(sdk.NewInt(b)), &m)
		require.Equal(t, expected, m.Int64())
	}

	// No-op tests
	testCase(2, 1, 40, 40) // a/b > 0
	testCase(1, 0, 50, 50) // b == 0

	// Zero result tests
	testCase(1, 2, 0, 0)  // (1/2)0 = 0
	testCase(0, 1, 60, 0) // (0/1)60 = 0

	// Round number tests
	testCase(1, 2, 70, 35)     // (1/2)70 = 35
	testCase(1, 2, 8866, 4433) // (1/2)8866 = 4433
	testCase(1, 3, 3000, 1000) // (1/3)3000 = 1000

	// Ceiling tests
	testCase(1, 3, 1, 1)  // (1/3)1 -> 1
	testCase(1, 3, 10, 4) // (1/3)10 -> 4
}
