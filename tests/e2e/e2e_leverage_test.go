package e2e

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/tests/grpc"
	"github.com/umee-network/umee/v6/x/leverage/fixtures"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
)

// sends a msgSupply from one of the suite's test accounts
func (s *E2ETest) leverageSupply(accountIndex int, denom string, amount uint64) {
	addr := s.AccountAddr(accountIndex)
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgSupply(addr, asset), s.AccountClient(accountIndex))
}

// sends a msgWithdraw from one of the suite's test accounts
func (s *E2ETest) leverageWithdraw(accountIndex int, denom string, amount uint64) {
	addr := s.AccountAddr(accountIndex)
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgWithdraw(addr, asset), s.AccountClient(accountIndex))
}

// sends a msgMaxWithdraw from one of the suite's test accounts
func (s *E2ETest) leverageMaxWithdraw(accountIndex int, denom string) {
	addr := s.AccountAddr(accountIndex)
	s.mustSucceedTx(leveragetypes.NewMsgMaxWithdraw(addr, denom), s.AccountClient(accountIndex))
}

// sends a msgCollateralize from one of the suite's test accounts
func (s *E2ETest) leverageCollateralize(accountIndex int, denom string, amount uint64) {
	addr := s.AccountAddr(accountIndex)
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgCollateralize(addr, asset), s.AccountClient(accountIndex))
}

// sends a msgDecollateralize from one of the suite's test accounts
func (s *E2ETest) leverageDecollateralize(accountIndex int, denom string, amount uint64) {
	addr := s.AccountAddr(accountIndex)
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgDecollateralize(addr, asset), s.AccountClient(accountIndex))
}

// sends a msgSupplyCollateral from one of the suite's test accounts
func (s *E2ETest) leverageSupplyCollateral(accountIndex int, denom string, amount uint64) {
	addr := s.AccountAddr(accountIndex)
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgSupplyCollateral(addr, asset), s.AccountClient(accountIndex))
}

// sends a msgBorrow from one of the suite's test accounts
func (s *E2ETest) leverageBorrow(accountIndex int, denom string, amount uint64) {
	addr := s.AccountAddr(accountIndex)
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgBorrow(addr, asset), s.AccountClient(accountIndex))
}

// sends a msgMaxBorrow from one of the suite's test accounts
func (s *E2ETest) leverageMaxBorrow(accountIndex int, denom string) {
	addr := s.AccountAddr(accountIndex)
	s.mustSucceedTx(leveragetypes.NewMsgMaxBorrow(addr, denom), s.AccountClient(accountIndex))
}

// sends a msgRepay from one of the suite's test accounts
func (s *E2ETest) leverageRepay(accountIndex int, denom string, amount uint64) {
	addr := s.AccountAddr(accountIndex)
	asset := sdk.NewCoin(denom, sdk.NewIntFromUint64(amount))
	s.mustSucceedTx(leveragetypes.NewMsgRepay(addr, asset), s.AccountClient(accountIndex))
}

// sends a msgLiquidate from one of the suite's test accounts targeting another
func (s *E2ETest) leverageLiquidate(accountIndex, targetIndex int, repayDenom string, repayAmount uint64, reward string) {
	addr := s.AccountAddr(accountIndex)
	target := s.AccountAddr(targetIndex)
	repay := sdk.NewCoin(repayDenom, sdk.NewIntFromUint64(repayAmount))
	s.mustSucceedTx(leveragetypes.NewMsgLiquidate(addr, target, repay, reward), s.AccountClient(accountIndex))
}

// sends a msgLeveragedLiquidate from one of the suite's test accounts targeting another
func (s *E2ETest) leverageLeveragedLiquidate(accountIndex, targetIndex int, repay, reward string) {
	addr := s.AccountAddr(accountIndex)
	target := s.AccountAddr(targetIndex)
	s.mustSucceedTx(leveragetypes.NewMsgLeveragedLiquidate(
		addr, target, repay, reward, sdk.ZeroDec()), s.AccountClient(accountIndex),
	)
}

func (s *E2ETest) TestLeverageBasics() {
	umeeNoMedians := fixtures.Token(appparams.BondDenom, "UMEE", 6)
	umeeNoMedians.HistoricMedians = 0
	updateTokens := []leveragetypes.Token{
		umeeNoMedians,
	}

	s.Run(
		"leverage update registry", func() {
			propID, err := grpc.LeverageRegistryUpdate(s.AccountClient(0), []leveragetypes.Token{}, updateTokens)
			s.Require().NoError(err)
			s.GovVoteAndWait(propID)
		},
	)

	// valAddr, err := s.Chain.Validators[0].KeyInfo.GetAddress()
	// s.Require().NoError(err)

	// TODO: check the blocks, rather than waiting arbitrary number of seconds
	// next tests depnds on the previous one, and we need to wait for the block.
	sleepTime := time.Millisecond * 1000 // 1.1s
	//
	s.Run(
		"initial leverage supply", func() {
			s.leverageSupply(0, appparams.BondDenom, 100_000_000)
		},
	)
	time.Sleep(sleepTime)
	s.Run(
		"initial leverage withdraw", func() {
			s.leverageWithdraw(0, "u/"+appparams.BondDenom, 10_000_000)
		},
	)
	time.Sleep(sleepTime)
	s.Run(
		"initial leverage collateralize", func() {
			s.leverageCollateralize(0, "u/"+appparams.BondDenom, 80_000_000)
		},
	)
	time.Sleep(sleepTime)
	s.Run(
		"initial leverage borrow", func() {
			s.leverageBorrow(0, appparams.BondDenom, 12_000_000)
		},
	)
	s.Run(
		"initial leverage repay", func() {
			s.leverageRepay(0, appparams.BondDenom, 2_000_000)
		},
	)
	s.Run(
		"too high leverage borrow", func() {
			asset := sdk.NewCoin(
				appparams.BondDenom,
				sdk.NewIntFromUint64(30_000_000),
			)
			s.mustFailTx(
				leveragetypes.NewMsgBorrow(s.AccountAddr(0), asset),
				s.AccountClient(0),
				"undercollateralized",
			)
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
				grpc.LeverageSpecialPairsUpdate(s.AccountClient(0), []leveragetypes.SpecialAssetSet{}, pairs),
			)
		},
	)
	// s.Run(
	// 	"special pair leverage borrow", func() {
	// 		asset := sdk.NewCoin(
	// 			appparams.BondDenom,
	// 			sdk.NewIntFromUint64(30_000_000),
	// 		)
	// 		s.mustSucceedTx(leveragetypes.NewMsgBorrow(valAddr, asset), s.AccountClient(0))
	// 	},
	// )
}
