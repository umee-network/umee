package e2e

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/tests/grpc"
	"github.com/umee-network/umee/v6/x/leverage/fixtures"
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

func (s *E2ETest) TestLeverageScenario() {
	s.Run(
		"register leverage tokens", func() {
			tokens := []leveragetypes.Token{
				fixtures.Token("test1", "WBTC", 8),
				fixtures.Token("test2", "WETH", 18),
				fixtures.Token("test3", "USDT", 6),
			}

			err := grpc.LeverageRegistryUpdate(s.Umee, tokens, nil)
			s.Require().NoError(err)

			sets := []leveragetypes.SpecialAssetSet{
				{
					Assets:               []string{"test1", "test2"},
					CollateralWeight:     sdk.MustNewDecFromStr("0.4"),
					LiquidationThreshold: sdk.MustNewDecFromStr("0.5"),
				},
			}

			err = grpc.LeverageSpecialPairsUpdate(s.Umee, sets, []leveragetypes.SpecialAssetPair{})
			s.Require().NoError(err)
		},
	)

	acc1, err := s.Chain.Accounts[0].KeyInfo.GetAddress()
	s.Require().NoError(err)

	s.Run(
		"initial leverage supply", func() {
			// Supply 100 UMEE, 1 WBTC, 10^-10 ETH, 100 USDC using 10^8 base denom each
			s.supply(acc1, appparams.BondDenom, 100_000_000)
			s.supply(acc1, "test1", 100_000_000)
			s.supply(acc1, "test2", 100_000_000)
			s.supply(acc1, "test3", 100_000_000)
		},
	)

	// TODO: one of each transaction, optionally with some failing ones, and including
	// liquidation by lowering the threshold on some existing tokens, assuming prices are up.
}
