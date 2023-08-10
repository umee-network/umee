package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v6/x/metoken"
)

func TestUnitParams(t *testing.T) {
	k := initSimpleKeeper(t)
	params := metoken.DefaultParams()

	err := k.SetParams(params)
	require.NoError(t, err)

	params2 := k.GetParams()
	require.Equal(t, params, params2)
}
