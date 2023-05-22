package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	k := newTestKeeper(t)

	// create a complex genesis state by running transactions
	_ = k.initScenario1()

	// get genesis state after this scenario
	gs1 := k.ExportGenesis(k.ctx)

	// require import-export idempotency on a fresh keeper
	k2 := newTestKeeper(t)
	k2.InitGenesis(k2.ctx, *gs1)
	gs2 := k2.ExportGenesis(k2.ctx)

	require.Equal(t, gs1, gs2, "genesis states equal")
}
