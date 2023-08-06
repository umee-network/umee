package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/umee-network/umee/v5/app/params"
	"github.com/umee-network/umee/v5/tests/accs"
	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/ugov"
)

func TestGasPrice(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	k := initKeeper(t)

	gpOut := k.MinGasPrice()
	require.Equal(gpOut, coin.UmeeDec("0"), "when nothing is set, 0uumee should be returned")

	gp := coin.Atom1_25dec
	k.SetMinGasPrice(gp)
	require.Equal(k.MinGasPrice(), gp)
}

func TestEmergencyGroup(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	k := initKeeper(t)

	require.Equal(k.EmergencyGroup(), sdk.AccAddress{},
		"when nothing is set, empty address should be returned")

	k.SetEmergencyGroup(accs.Alice)
	require.Equal(k.EmergencyGroup(), accs.Alice)
}

func TestLiquidationParams(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	k := initKeeper(t)

	require.Equal(k.InflationParams(), ugov.InflationParams{},
		"when nothing is set, empty inflationp params should return")

	dip := ugov.DefaultInflationParams()
	k.SetInflationParams(dip)
	p := k.InflationParams()
	require.Equal(dip, p)
	require.Equal(dip.MaxSupply.GetDenom(), appparams.BondDenom)
}

func TestInflationCycleEnd(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	k := initKeeper(t)

	st := time.Time{}
	err := k.SetInflationCycleEnd(st)
	require.NoError(err)
	end := k.InflationCycleEnd()
	require.Equal(end.IsZero(), true, "it should be default zero time")

	cycleEnd := time.Now()
	err = k.SetInflationCycleEnd(cycleEnd)
	require.NoError(err)
	end = k.InflationCycleEnd()
	require.Equal(end, cycleEnd.Truncate(time.Millisecond), "inflation cycle end time should be same")
}
