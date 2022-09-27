package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v3/x/leverage/types"
)

func (s *IntegrationTestSuite) TestSupplyGas() {
	tk, app, ctx, require := s.tk, s.app, s.ctx, s.Require()

	// create and fund a supplier with 100 UMEE
	supplier := s.newAccount(coin(umeeDenom, 100_000000))

	// Ensure that token registry is not cached from setup
	tk.ClearCache()

	// reset gas and supply umee without having cached its token
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(10_000000))
	_, err := app.LeverageKeeper.Supply(ctx, supplier, coin(umeeDenom, 10_000000))
	require.NoError(err)

	// measure gas use (74602 observed in practice)
	firstGasUsed := int(ctx.GasMeter().GasConsumed())
	require.Equal(74602, firstGasUsed, "pre-query supply gas use")

	// Clear token registry cach from previous step
	tk.ClearCache()

	// Query market summary, which will cause registered tokens to be cached
	req := &types.QueryMarketSummary{Denom: umeeDenom}
	_, err = s.queryClient.MarketSummary(ctx, req)
	require.NoError(err)

	// reset gas and supply umee after having cached its token
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(10_000000))
	_, err = app.LeverageKeeper.Supply(ctx, supplier, coin(umeeDenom, 10_000000))
	require.NoError(err)

	// measure gas use after second supply transaction and ensure it equals the first
	secondGasUsed := int(ctx.GasMeter().GasConsumed())
	require.Equal(firstGasUsed, secondGasUsed, "post-query supply gas use")
}
