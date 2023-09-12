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

func (s *E2ETest) TestLeverageBasics() {
	umeeNoMedians := fixtures.Token(appparams.BondDenom, "UMEE", 6)
	umeeNoMedians.HistoricMedians = 0
	updateTokens := []leveragetypes.Token{
		umeeNoMedians,
	}

	s.Run(
		"leverage update registry", func() {
			s.Require().NoError(
				grpc.LeverageRegistryUpdate(s.Umee, []leveragetypes.Token{}, updateTokens),
			)
		},
	)

	valAddr, err := s.Chain.Validators[0].KeyInfo.GetAddress()
	s.Require().NoError(err)

	s.Run(
		"initial leverage supply", func() {
			s.supply(valAddr, appparams.BondDenom, 100_000_000)
		},
	)
	s.Run(
		"initial leverage withdraw", func() {
			s.withdraw(valAddr, "u/"+appparams.BondDenom, 10_000_000)
		},
	)
	s.Run(
		"initial leverage collateralize", func() {
			s.collateralize(valAddr, "u/"+appparams.BondDenom, 80_000_000)
		},
	)
	s.Run(
		"initial leverage borrow", func() {
			s.borrow(valAddr, appparams.BondDenom, 12_000_000)
		},
	)
	s.Run(
		"initial leverage repay", func() {
			s.repay(valAddr, appparams.BondDenom, 2_000_000)
		},
	)
	s.Run(
		"too high leverage borrow", func() {
			asset := sdk.NewCoin(
				appparams.BondDenom,
				sdk.NewIntFromUint64(30_000_000),
			)
			s.mustFailTx(leveragetypes.NewMsgBorrow(valAddr, asset), "undercollateralized")
		},
	)
	s.Run(
		"leverage add special pairs", func() {
			sets := []leveragetypes.SpecialAssetSet{
				{
					// a set allowing UMEE to borrow more of itself
					Assets:               []string{appparams.BondDenom},
					CollateralWeight:     sdk.MustNewDecFromStr("0.75"),
					LiquidationThreshold: sdk.MustNewDecFromStr("0.8"),
				},
			}
			s.Require().NoError(
				grpc.LeverageSpecialPairsUpdate(s.Umee, sets, []leveragetypes.SpecialAssetPair{}),
			)
		},
	)
	s.Run(
		"special pair leverage borrow", func() {
			asset := sdk.NewCoin(
				appparams.BondDenom,
				sdk.NewIntFromUint64(30_000_000),
			)
			s.mustEventuallySucceedTx(leveragetypes.NewMsgBorrow(valAddr, asset))
		},
	)
}
