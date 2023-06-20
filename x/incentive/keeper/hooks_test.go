package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/incentive"
)

func TestHooks(t *testing.T) {
	t.Parallel()
	k := newTestKeeper(t)
	require := require.New(t)

	// create a complex genesis state by running transactions
	alice := k.initScenario1()

	h := k.BondHooks()

	require.Equal(sdk.NewInt(100_000000), h.GetBonded(k.ctx, alice, uUmee), "initial restricted collateral")
	require.NoError(h.ForceUnbondTo(k.ctx, alice, coin.New(uUmee, 200_000000)), "liquidation unbond with no effect")
	require.Equal(sdk.NewInt(100_000000), h.GetBonded(k.ctx, alice, uUmee), "unchanged restricted collateral")

	// verify scenario 1 state is still unchanged by liquidation
	bonded, unbonding, unbondings := k.BondSummary(k.ctx, alice, uUmee)
	require.Equal(coin.New(uUmee, 90_000000), bonded)
	require.Equal(coin.New(uUmee, 10_000000), unbonding)
	require.Equal([]incentive.Unbonding{
		incentive.NewUnbonding(90, 86490, coin.New(uUmee, 5_000000)),
		incentive.NewUnbonding(90, 86490, coin.New(uUmee, 5_000000)),
	}, unbondings)

	// reduce a single in-progress unbonding by liquidation
	require.NoError(h.ForceUnbondTo(k.ctx, alice, coin.New(uUmee, 96_000000)), "liquidation unbond 1")
	require.Equal(sdk.NewInt(96_000000), h.GetBonded(k.ctx, alice, uUmee))
	bonded, unbonding, unbondings = k.BondSummary(k.ctx, alice, uUmee)
	require.Equal(coin.New(uUmee, 90_000000), bonded)
	require.Equal(coin.New(uUmee, 6_000000), unbonding)
	require.Equal([]incentive.Unbonding{
		incentive.NewUnbonding(90, 86490, coin.New(uUmee, 1_000000)),
		incentive.NewUnbonding(90, 86490, coin.New(uUmee, 5_000000)),
	}, unbondings)

	// reduce two in-progress unbondings by liquidation (one is ended altogether)
	require.NoError(h.ForceUnbondTo(k.ctx, alice, coin.New(uUmee, 92_000000)), "liquidation unbond 2")
	require.Equal(sdk.NewInt(92_000000), h.GetBonded(k.ctx, alice, uUmee))
	bonded, unbonding, unbondings = k.BondSummary(k.ctx, alice, uUmee)
	require.Equal(coin.New(uUmee, 90_000000), bonded)
	require.Equal(coin.New(uUmee, 2_000000), unbonding)
	require.Equal([]incentive.Unbonding{
		incentive.NewUnbonding(90, 86490, coin.New(uUmee, 2_000000)),
	}, unbondings)

	// end all unbondings and reduce bonded amount by liquidation
	require.NoError(h.ForceUnbondTo(k.ctx, alice, coin.New(uUmee, 46_000000)), "liquidation unbond 3")
	require.Equal(sdk.NewInt(46_000000), h.GetBonded(k.ctx, alice, uUmee))
	bonded, unbonding, unbondings = k.BondSummary(k.ctx, alice, uUmee)
	require.Equal(coin.New(uUmee, 46_000000), bonded)
	require.Equal(coin.Zero(uUmee), unbonding)
	require.Equal([]incentive.Unbonding{}, unbondings)

	// clear bonds by liquidation
	require.NoError(h.ForceUnbondTo(k.ctx, alice, coin.Zero(uUmee)), "liquidation unbond to zero")
	require.Equal(sdk.ZeroInt(), h.GetBonded(k.ctx, alice, uUmee))
	bonded, unbonding, unbondings = k.BondSummary(k.ctx, alice, uUmee)
	require.Equal(coin.Zero(uUmee), bonded)
	require.Equal(coin.Zero(uUmee), unbonding)
	require.Equal([]incentive.Unbonding{}, unbondings)
}
