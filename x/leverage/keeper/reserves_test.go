package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	umeeapp "github.com/umee-network/umee/v2/app"
)

func (s *IntegrationTestSuite) TestSetReserves() {
	// get initial reserves
	amount := s.app.LeverageKeeper.GetReserveAmount(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(amount, sdk.ZeroInt())

	// artifically reserve 200 umee
	err := s.tk.SetReserveAmount(s.ctx, sdk.NewInt64Coin(umeeapp.BondDenom, 200000000))
	s.Require().NoError(err)

	// get new reserves
	amount = s.app.LeverageKeeper.GetReserveAmount(s.ctx, umeeapp.BondDenom)
	s.Require().Equal(amount, sdk.NewInt(200000000))
}

func (s *IntegrationTestSuite) TestRepayBadDebt() {
	// Creating a supplier so module account has some uumee
	_ = s.setupAccount(umeeDenom, 200000000, 200000000, 0, false) // 200 umee

	// Using an address with no assets
	addr := s.setupAccount(umeeDenom, 0, 0, 0, false)

	// Create an uncollateralized debt position
	badDebt := sdk.NewInt64Coin(umeeDenom, 100000000) // 100 umee
	err := s.tk.SetBorrow(s.ctx, addr, badDebt)
	s.Require().NoError(err)

	// Manually mark the bad debt for repayment
	s.Require().NoError(s.tk.SetBadDebtAddress(s.ctx, addr, umeeDenom, true))

	// Manually set reserves to 60 umee
	reserve := sdk.NewInt64Coin(umeeDenom, 60000000)
	err = s.tk.SetReserveAmount(s.ctx, reserve)
	s.Require().NoError(err)

	// Sweep all bad debts, which should repay 60 umee of the bad debt (partial repayment)
	err = s.app.LeverageKeeper.SweepBadDebts(s.ctx)
	s.Require().NoError(err)

	// Confirm that a debt of 40 umee remains
	remainingDebt := s.app.LeverageKeeper.GetBorrow(s.ctx, addr, umeeDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 40000000), remainingDebt)

	// Confirm that reserves are exhausted
	remainingReserve := s.app.LeverageKeeper.GetReserveAmount(s.ctx, umeeDenom)
	s.Require().Equal(sdk.ZeroInt(), remainingReserve)

	// Manually set reserves to 70 umee
	reserve = sdk.NewInt64Coin(umeeDenom, 70000000)
	err = s.tk.SetReserveAmount(s.ctx, reserve)
	s.Require().NoError(err)

	// Sweep all bad debts, which should fully repay the bad debt this time
	err = s.app.LeverageKeeper.SweepBadDebts(s.ctx)
	s.Require().NoError(err)

	// Confirm that the debt is eliminated
	remainingDebt = s.app.LeverageKeeper.GetBorrow(s.ctx, addr, umeeDenom)
	s.Require().Equal(sdk.NewInt64Coin(umeeDenom, 0), remainingDebt)

	// Confirm that reserves are now at 30 umee
	remainingReserve = s.app.LeverageKeeper.GetReserveAmount(s.ctx, umeeDenom)
	s.Require().Equal(sdk.NewInt(30000000), remainingReserve)

	// Sweep all bad debts - but there are none
	err = s.app.LeverageKeeper.SweepBadDebts(s.ctx)
	s.Require().NoError(err)
}
