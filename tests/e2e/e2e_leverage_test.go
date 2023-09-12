package e2e

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v6/app/params"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
)

func (s *E2ETest) supply(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgSupply(addr, asset))
}

func (s *E2ETest) withdraw(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgWithdraw(addr, asset))
}

func (s *E2ETest) maxWithdraw(addr sdk.AccAddress, denom string) {
	s.mustSucceedTx(leveragetypes.NewMsgMaxWithdraw(addr, denom))
}

func (s *E2ETest) collateralize(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgCollateralize(addr, asset))
}

func (s *E2ETest) decollateralize(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgDecollateralize(addr, asset))
}

func (s *E2ETest) supplyCollateral(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgSupplyCollateral(addr, asset))
}

func (s *E2ETest) borrow(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgBorrow(addr, asset))
}

func (s *E2ETest) maxBorrow(addr sdk.AccAddress, denom string) {
	s.mustSucceedTx(leveragetypes.NewMsgMaxBorrow(addr, denom))
}

func (s *E2ETest) repay(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgRepay(addr, asset))
}

func (s *E2ETest) liquidate(addr, target sdk.AccAddress, repayDenom string, repayAmount uint64, reward string) {
	repay := sdk.NewCoin(repayDenom, sdk.NewIntFromUint64(repayAmount))
	s.mustSucceedTx(leveragetypes.NewMsgLiquidate(addr, target, repay, reward))
}

func (s *E2ETest) leveragedLiquidate(addr, target sdk.AccAddress, repay, reward string) {
	s.mustSucceedTx(leveragetypes.NewMsgLeveragedLiquidate(addr, target, repay, reward))
}

func (s *E2ETest) TestLeverageBasics() {
	valAddr, err := s.Chain.Validators[0].KeyInfo.GetAddress()
	s.Require().NoError(err)

	s.Run(
		"initial leverage supply", func() {
			s.supply(valAddr, appparams.BondDenom, 100_000_000)
		},
	)
	s.Run(
		"initial leverage withdraw", func() {
			s.supply(valAddr, "u/"+appparams.BondDenom, 10_000_000)
		},
	)
	s.Run(
		"initial leverage collateralize", func() {
			s.collateralize(valAddr, "u/"+appparams.BondDenom, 70_000_000)
		},
	)
	s.Run(
		"initial leverage borrow", func() {
			s.borrow(valAddr, appparams.BondDenom, 10_000_000)
		},
	)
	s.Run(
		"initial leverage repay", func() {
			s.repay(valAddr, appparams.BondDenom, 5_000_000)
		},
	)
}
