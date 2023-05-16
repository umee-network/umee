package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/v4/util/coin"
)

func TestGasPrice(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	k := initKeeper(t)

	gpOut := k.MinGasPrice()
	require.Equal(*gpOut, coin.UmeeDec("0"), "when nothing is set, 0uumee should be returned")

	gp := coin.Atom1_25dec
	k.SetMinGasPrice(&gp)
	require.Equal(k.MinGasPrice(), &gp)
}
