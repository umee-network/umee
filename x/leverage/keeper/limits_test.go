package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) TestInt64Overflow() {
	ctx, app, require := s.ctx, s.app, s.Require()

	// set uumee max supply to unlimited
	t, err := app.LeverageKeeper.GetTokenSettings(ctx, umeeDenom)
	require.NoError(err)
	t.MaxSupply = sdk.ZeroInt()
	require.NoError(app.LeverageKeeper.SetTokenSettings(ctx, t))

	overflowCoin := sdk.NewInt64Coin(umeeDenom, 9223372036854775807) // max int64
	overflowCoin.Amount = overflowCoin.Amount.Add(sdk.OneInt())
	supplier := s.newAccount(overflowCoin)
	s.supply(supplier, overflowCoin)

	mal, err := app.LeverageKeeper.ModuleAvailableLiquidity(ctx, umeeDenom)
	require.NoError(err)
	require.Equal(overflowCoin.Amount, mal, "module available liquidity above max int64")
}
