package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v5/tests/accs"
	"github.com/umee-network/umee/v5/util/coin"
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
