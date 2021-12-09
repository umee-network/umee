package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	umeeapp "github.com/umee-network/umee/app"
)

func (s *IntegrationTestSuite) TestOracle() {
	// register uumee and u/uumee as an accepted asset+utoken pair
	s.leverageKeeper.SetTokenDenom(s.ctx, umeeapp.BondDenom)

	validCoin := sdk.NewInt64Coin(umeeapp.BondDenom, 1234000) // 1.234 umee

	// Get the USD value of a single coin
	value, err := s.leverageKeeper.Price(s.ctx, validCoin)
	s.Require().NoError(err)
	//   TODO #97: Change to the correct expected USD value when oracle is integrated
	s.Require().Equal(sdk.MustNewDecFromStr("1234000"), value)

	// TODO #97: Add a second valid coin, so the TotalPrice test below can
	// properly add up their prices.

	// Get the total USD value of an sdk.Coins containing multiple valid denoms
	value, err = s.leverageKeeper.TotalPrice(s.ctx, sdk.NewCoins(validCoin))
	s.Require().NoError(err)
	//   TODO #97: Change to the correct expected USD value when oracle is integrated
	s.Require().Equal(sdk.MustNewDecFromStr("1234000"), value)

	// TODO : Using two valid denoms, test keeper.EquivalentValue
}

func (s *IntegrationTestSuite) TestOracle_Invalid() {
	invalidCoin := sdk.NewInt64Coin("uabcd", 1000000)

	_, err := s.leverageKeeper.Price(s.ctx, invalidCoin)
	s.Require().Error(err)
}
