package e2e

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/tests/grpc"
	"github.com/umee-network/umee/v6/x/leverage/fixtures"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
)

func (s *E2ETest) leverageSupply(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgSupply(addr, asset))
}

func (s *E2ETest) leverageWithdraw(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgWithdraw(addr, asset))
}

func (s *E2ETest) leverageMaxWithdraw(addr sdk.AccAddress, denom string) {
	s.mustSucceedTx(leveragetypes.NewMsgMaxWithdraw(addr, denom))
}

func (s *E2ETest) leverageCollateralize(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgCollateralize(addr, asset))
}

func (s *E2ETest) leverageDecollateralize(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgDecollateralize(addr, asset))
}

func (s *E2ETest) leverageSupplyCollateral(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgSupplyCollateral(addr, asset))
}

func (s *E2ETest) leverageBorrow(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgBorrow(addr, asset))
}

func (s *E2ETest) leverageMaxBorrow(addr sdk.AccAddress, denom string) {
	s.mustSucceedTx(leveragetypes.NewMsgMaxBorrow(addr, denom))
}

func (s *E2ETest) leverageRepay(addr sdk.AccAddress, denom string, amount uint64) {
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgRepay(addr, asset))
}

func (s *E2ETest) leverageLiquidate(addr, target sdk.AccAddress, repayDenom string, repayAmount uint64, reward string) {
	repay := sdk.NewCoin(repayDenom, sdk.NewIntFromUint64(repayAmount))
	s.mustSucceedTx(leveragetypes.NewMsgLiquidate(addr, target, repay, reward))
}

func (s *E2ETest) leverageLeveragedLiquidate(addr, target sdk.AccAddress, repay, reward string) {
	s.mustSucceedTx(leveragetypes.NewMsgLeveragedLiquidate(addr, target, repay, reward, sdk.ZeroDec()))
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
			s.leverageSupply(valAddr, appparams.BondDenom, 100_000_000)
		},
	)
	s.Run(
		"initial leverage withdraw", func() {
			s.leverageWithdraw(valAddr, "u/"+appparams.BondDenom, 10_000_000)
		},
	)
	s.Run(
		"initial leverage collateralize", func() {
			s.leverageCollateralize(valAddr, "u/"+appparams.BondDenom, 80_000_000)
		},
	)
	s.Run(
		"initial leverage borrow", func() {
			s.leverageBorrow(valAddr, appparams.BondDenom, 12_000_000)
		},
	)
	s.Run(
		"initial leverage repay", func() {
			s.leverageRepay(valAddr, appparams.BondDenom, 2_000_000)
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
			pairs := []leveragetypes.SpecialAssetPair{
				{
					// a set allowing UMEE to borrow more of itself
					Borrow:               appparams.BondDenom,
					Collateral:           appparams.BondDenom,
					CollateralWeight:     sdk.MustNewDecFromStr("0.75"),
					LiquidationThreshold: sdk.MustNewDecFromStr("0.8"),
				},
			}
			s.Require().NoError(
				grpc.LeverageSpecialPairsUpdate(s.Umee, []leveragetypes.SpecialAssetSet{}, pairs),
			)
		},
	)
	s.Run(
		"special pair leverage borrow", func() {
			asset := sdk.NewCoin(
				appparams.BondDenom,
				sdk.NewIntFromUint64(30_000_000),
			)
			s.mustSucceedTx(leveragetypes.NewMsgBorrow(valAddr, asset))
		},
	)
}
