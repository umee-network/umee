package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/x/leverage/keeper"
	"github.com/umee-network/umee/v2/x/leverage/types"
)

// checkInvariants is used during other tests to quickly test all invariants
func (s *IntegrationTestSuite) checkInvariants(msg string) {
	desc, broken := keeper.AllInvariants(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken, msg, desc)
}

func (s *IntegrationTestSuite) TestReserveAmountInvariant() {
	// artificially set reserves
	err := s.tk.SetReserveAmount(s.ctx, sdk.NewInt64Coin(umeeapp.BondDenom, 300000000)) // 300 umee
	s.Require().NoError(err)

	// check invariants
	_, broken := keeper.ReserveAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)
}

func (s *IntegrationTestSuite) TestCollateralAmountInvariant() {
	addr, _ := s.initBorrowScenario()

	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// check invariant
	_, broken := keeper.CollateralAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)

	uTokenDenom := types.ToUTokenDenom(umeeapp.BondDenom)

	// withdraw the supplied umee in the initBorrowScenario
	_, err := s.app.LeverageKeeper.Withdraw(s.ctx, addr, sdk.NewInt64Coin(uTokenDenom, 1000000000))
	s.Require().NoError(err)

	// check invariant
	_, broken = keeper.CollateralAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)
}

func (s *IntegrationTestSuite) TestBorrowAmountInvariant() {
	addr, _ := s.initBorrowScenario()

	// The "supplier" user from the init scenario is being used because it
	// already has 1k u/umee for collateral

	// user borrows 20 umee
	err := s.app.LeverageKeeper.Borrow(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 20000000))
	s.Require().NoError(err)

	// check invariant
	_, broken := keeper.BorrowAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)

	// user repays 30 umee, actually only 20 because is the min between
	// the amount borrowed and the amount repaid
	_, err = s.app.LeverageKeeper.Repay(s.ctx, addr, sdk.NewInt64Coin(umeeapp.BondDenom, 30000000))
	s.Require().NoError(err)

	// check invariant
	_, broken = keeper.BorrowAmountInvariant(s.app.LeverageKeeper)(s.ctx)
	s.Require().False(broken)
}
