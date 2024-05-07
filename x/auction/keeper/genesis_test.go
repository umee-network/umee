package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v6/x/auction"
)

func TestGenesis(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	k := initKeeper(t)

	gs := auction.DefaultGenesis()
	err := k.InitGenesis(gs)
	require.NoError(err)

	gsOut, err := k.ExportGenesis()
	require.NoError(err)
	require.Equal(gs, gsOut)

}
