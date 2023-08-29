package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/ugov"
	"github.com/umee-network/umee/v6/x/ugov/keeper/intest"
)

// creates keeper without external dependencies (app, leverage etc...)
func initKeeper(t *testing.T) TestKeeper {
	ctx, k := intest.MkKeeper(t)
	return TestKeeper{k, t, ctx}
}

type TestKeeper struct {
	ugov.Keeper
	t   *testing.T
	ctx *sdk.Context
}
