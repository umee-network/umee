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

	dlp := ugov.DefaultInflationParams()
	k.SetInflationParams(dlp)
	rlp := k.InflationParams()
	require.Equal(rlp, dlp)
	require.Equal(rlp.MaxSupply.GetDenom(), appparams.BondDenom)
}

func TestInflationCycleStartTime(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	k := initKeeper(t)

	st := time.Time{}
	err := k.SetInflationCycleStartTime(st)
	require.NoError(err)
	in_c, err := k.GetInflationCycleStartTime()
	require.NoError(err)
	require.Equal(in_c.IsZero(), true, "it should be default zero time")

	icst := time.Now()
	err = k.SetInflationCycleStartTime(icst)
	require.NoError(err)
	ricst, err := k.GetInflationCycleStartTime()
	require.NoError(err)
	require.Equal(ricst.Equal(icst), true, "inflation cycle start time should be same")
}
