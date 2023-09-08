package e2e

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
)

func (s *E2ETest) executeSupply(addr string, asset sdk.Coin) {
	msg := &leveragetypes.MsgSupply{
		Supplier: addr,
		Asset:    asset,
	}
	s.executeTx(msg)
}

func (s *E2ETest) executeWithdraw(addr string, asset sdk.Coin) {
	msg := &leveragetypes.MsgWithdraw{
		Supplier: addr,
		Asset:    asset,
	}
	s.executeTx(msg)
}

func (s *E2ETest) executeMaxWithdraw(addr, denom string) {
	msg := &leveragetypes.MsgMaxWithdraw{
		Supplier: addr,
		Denom:    denom,
	}
	s.executeTx(msg)
}

func (s *E2ETest) executeCollateralize(addr string, asset sdk.Coin) {
	msg := &leveragetypes.MsgCollateralize{
		Borrower: addr,
		Asset:    asset,
	}
	s.executeTx(msg)
}

func (s *E2ETest) executeDecollateralize(addr string, asset sdk.Coin) {
	msg := &leveragetypes.MsgDecollateralize{
		Borrower: addr,
		Asset:    asset,
	}
	s.executeTx(msg)
}

func (s *E2ETest) executeSupplyCollateral(addr string, asset sdk.Coin) {
	msg := &leveragetypes.MsgSupplyCollateral{
		Supplier: addr,
		Asset:    asset,
	}
	s.executeTx(msg)
}

func (s *E2ETest) executeBorrow(addr string, asset sdk.Coin) {
	msg := &leveragetypes.MsgBorrow{
		Borrower: addr,
		Asset:    asset,
	}
	s.executeTx(msg)
}

func (s *E2ETest) executeRepay(addr string, asset sdk.Coin) {
	msg := &leveragetypes.MsgRepay{
		Borrower: addr,
		Asset:    asset,
	}
	s.executeTx(msg)
}

func (s *E2ETest) executeLiquidate(addr, target, reward string, repay sdk.Coin) {
	msg := &leveragetypes.MsgLiquidate{
		Liquidator:  addr,
		Borrower:    target,
		Repayment:   repay,
		RewardDenom: reward,
	}
	s.executeTx(msg)
}

func (s *E2ETest) executeLeveragedLiquidate(addr, target, repay, reward string) {
	msg := &leveragetypes.MsgLeveragedLiquidate{
		Liquidator:  addr,
		Borrower:    target,
		RepayDenom:  repay,
		RewardDenom: reward,
	}
	s.executeTx(msg)
}
