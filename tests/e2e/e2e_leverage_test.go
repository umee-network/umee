package e2e

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
)

func (s *E2ETest) executeSupply(addr sdk.AccAddress, asset sdk.Coin) {
	s.executeTx(leveragetypes.NewMsgSupply(addr, asset))
}

func (s *E2ETest) executeWithdraw(addr sdk.AccAddress, asset sdk.Coin) {
	s.executeTx(leveragetypes.NewMsgWithdraw(addr, asset))
}

func (s *E2ETest) executeMaxWithdraw(addr sdk.AccAddress, denom string) {
	s.executeTx(leveragetypes.NewMsgMaxWithdraw(addr, denom))
}

func (s *E2ETest) executeCollateralize(addr sdk.AccAddress, asset sdk.Coin) {
	s.executeTx(leveragetypes.NewMsgCollateralize(addr, asset))
}

func (s *E2ETest) executeDecollateralize(addr sdk.AccAddress, asset sdk.Coin) {
	s.executeTx(leveragetypes.NewMsgDecollateralize(addr, asset))
}

func (s *E2ETest) executeSupplyCollateral(addr sdk.AccAddress, asset sdk.Coin) {
	s.executeTx(leveragetypes.NewMsgSupplyCollateral(addr, asset))
}

func (s *E2ETest) executeBorrow(addr sdk.AccAddress, asset sdk.Coin) {
	s.executeTx(leveragetypes.NewMsgBorrow(addr, asset))
}

func (s *E2ETest) executeMaxBorrow(addr sdk.AccAddress, denom string) {
	s.executeTx(leveragetypes.NewMsgMaxBorrow(addr, denom))
}

func (s *E2ETest) executeRepay(addr sdk.AccAddress, asset sdk.Coin) {
	s.executeTx(leveragetypes.NewMsgRepay(addr, asset))
}

func (s *E2ETest) executeLiquidate(addr, target sdk.AccAddress, reward string, repay sdk.Coin) {
	s.executeTx(leveragetypes.NewMsgLiquidate(addr, target, repay, reward))
}

func (s *E2ETest) executeLeveragedLiquidate(addr, target sdk.AccAddress, repay, reward string) {
	s.executeTx(leveragetypes.NewMsgLeveragedLiquidate(addr, target, repay, reward))
}
