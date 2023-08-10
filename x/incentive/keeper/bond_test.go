package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v6/util/coin"
)

func TestBonds(t *testing.T) {
	t.Parallel()
	k := newTestKeeper(t)
	ctx := k.ctx
	require := require.New(t)

	// init a supplier with bonded uTokens, and some currently unbonding
	alice := k.newBondedAccount(
		coin.New(uUmee, 100_000000),
		coin.New(uAtom, 20_000000),
	)
	k.mustBeginUnbond(alice, coin.New(uUmee, 10_000000))
	k.mustBeginUnbond(alice, coin.New(uUmee, 5_000000))

	// restricted collateral counts bonded and unbonding uTokens
	restricted := k.restrictedCollateral(ctx, alice, uUmee)
	require.Equal(coin.New(uUmee, 100_000000), restricted)
	restricted = k.restrictedCollateral(ctx, alice, uAtom)
	require.Equal(coin.New(uAtom, 20_000000), restricted)

	// bond summary
	bonded, unbonding, unbondings := k.BondSummary(ctx, alice, uUmee)
	require.Equal(coin.New(uUmee, 85_000000), bonded)
	require.Equal(coin.New(uUmee, 15_000000), unbonding)
	require.Equal(2, len(unbondings))
	bonded, unbonding, unbondings = k.BondSummary(ctx, alice, uAtom)
	require.Equal(coin.New(uAtom, 20_000000), bonded)
	require.Equal(coin.Zero(uAtom), unbonding)
	require.Equal(0, len(unbondings))

	// decreaseBond is an internal function that instantly unbonds uTokens
	err := k.decreaseBond(ctx, alice, coin.New(uAtom, 15_000000))
	require.NoError(err)
	require.Equal(coin.New(uAtom, 5_000000), k.GetBonded(ctx, alice, uAtom))
}
