package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v4/util/coin"
)

func TestBonds(t *testing.T) {
	k := newTestKeeper(t)
	ctx := k.ctx
	require := require.New(t)

	// init a supplier with bonded uTokens, and some currently unbonding
	alice := k.newBondedAccount(
		coin.New(u_umee, 100_000000),
		coin.New(u_atom, 20_000000),
	)
	k.mustBeginUnbond(alice, coin.New(u_umee, 10_000000))
	k.mustBeginUnbond(alice, coin.New(u_umee, 5_000000))

	// restricted collateral counts bonded and unbonding uTokens
	restricted := k.restrictedCollateral(ctx, alice, u_umee)
	require.Equal(coin.New(u_umee, 100_000000), restricted)
	restricted = k.restrictedCollateral(ctx, alice, u_atom)
	require.Equal(coin.New(u_atom, 20_000000), restricted)

	// bond summary
	bonded, unbonding, unbondings := k.BondSummary(ctx, alice, u_umee)
	require.Equal(coin.New(u_umee, 85_000000), bonded)
	require.Equal(coin.New(u_umee, 15_000000), unbonding)
	require.Equal(2, len(unbondings))
	bonded, unbonding, unbondings = k.BondSummary(ctx, alice, u_atom)
	require.Equal(coin.New(u_atom, 20_000000), bonded)
	require.Equal(coin.Zero(u_atom), unbonding)
	require.Equal(0, len(unbondings))

	// decreaseBond is an internal function that instantly unbonds uTokens
	err := k.decreaseBond(ctx, alice, coin.New(u_atom, 15_000000))
	require.NoError(err)
	require.Equal(coin.New(u_atom, 5_000000), k.GetBonded(ctx, alice, u_atom))
}
