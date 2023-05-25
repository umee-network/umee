package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v4/x/ugov"
)

func TestGenesis(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	k := initKeeper(t)

	gs := ugov.DefaultGenesis()
	err := k.InitGenesis(gs)
	require.NoError(err)

	gsOut := k.ExportGenesis()
	require.Equal(gs, gsOut)
}
