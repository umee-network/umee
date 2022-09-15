package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v3/app/params"
)

func (s *IntegrationTestSuite) TestDeriveExchangeRate() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 uumee
	addr := s.newAccount(coin(umeeDenom, 1000))
	s.supply(addr, coin(umeeDenom, 1000))
	s.collateralize(addr, coin("u/"+umeeDenom, 1000))

	// artificially increase total borrows (by affecting a single address)
	err := s.tk.SetBorrow(ctx, addr, coin(umeeDenom, 2000))
	require.NoError(err)

	// artificially set reserves
	s.setReserves(coin(umeeDenom, 300))

	// expected token:uToken exchange rate
	//    = (total borrows + module balance - reserves) / utoken supply
	//    = 2000 + 1000 - 300 / 1000
	//    = 2.7

	// get derived exchange rate
	rate := app.LeverageKeeper.DeriveExchangeRate(ctx, appparams.BondDenom)
	require.Equal(sdk.MustNewDecFromStr("2.7"), rate)
}
