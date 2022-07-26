package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v2/x/leverage/keeper"
)

func TestInterpolate(t *testing.T) {
	// Define two points (x1,y1) and (x2,y2)
	x1 := sdk.MustNewDecFromStr("3.0")
	x2 := sdk.MustNewDecFromStr("6.0")
	y1 := sdk.MustNewDecFromStr("11.1")
	y2 := sdk.MustNewDecFromStr("17.4")

	// Sloped line, endpoint checks
	x := keeper.Interpolate(x1, x1, y1, x2, y2)
	require.Equal(t, x, y1)
	x = keeper.Interpolate(x2, x1, y1, x2, y2)
	require.Equal(t, x, y2)

	// Sloped line, point on segment
	x = keeper.Interpolate(sdk.MustNewDecFromStr("4.0"), x1, y1, x2, y2)
	require.Equal(t, x, sdk.MustNewDecFromStr("13.2"))

	// Sloped line, point outside of segment
	x = keeper.Interpolate(sdk.MustNewDecFromStr("2.0"), x1, y1, x2, y2)
	require.Equal(t, x, sdk.MustNewDecFromStr("9.0"))

	// Vertical line: always return y1
	x = keeper.Interpolate(sdk.ZeroDec(), x1, y1, x1, y2)
	require.Equal(t, x, y1)
	x = keeper.Interpolate(x1, x1, y1, x1, y2)
	require.Equal(t, x, y1)

	// Undefined line (x1=x2, y1=y2): always return y1
	x = keeper.Interpolate(sdk.ZeroDec(), x1, y1, x1, y1)
	require.Equal(t, x, y1)
	x = keeper.Interpolate(x1, x1, y1, x1, y1)
	require.Equal(t, x, y1)
}
