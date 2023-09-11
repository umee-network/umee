package e2e

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/tests/grpc"
	"github.com/umee-network/umee/v6/x/leverage/fixtures"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
)

func (s *E2ETest) supply(addr sdk.AccAddress, asset sdk.Coin) {
	s.mustSucceedTx(leveragetypes.NewMsgSupply(addr, asset))
}

func (s *E2ETest) withdraw(addr sdk.AccAddress, asset sdk.Coin) {
	s.mustSucceedTx(leveragetypes.NewMsgWithdraw(addr, asset))
}

func (s *E2ETest) maxWithdraw(addr sdk.AccAddress, denom string) {
	s.mustSucceedTx(leveragetypes.NewMsgMaxWithdraw(addr, denom))
}

func (s *E2ETest) collateralize(addr sdk.AccAddress, asset sdk.Coin) {
	s.mustSucceedTx(leveragetypes.NewMsgCollateralize(addr, asset))
}

func (s *E2ETest) decollateralize(addr sdk.AccAddress, asset sdk.Coin) {
	s.mustSucceedTx(leveragetypes.NewMsgDecollateralize(addr, asset))
}

func (s *E2ETest) supplyCollateral(addr sdk.AccAddress, asset sdk.Coin) {
	s.mustSucceedTx(leveragetypes.NewMsgSupplyCollateral(addr, asset))
}

func (s *E2ETest) borrow(addr sdk.AccAddress, asset sdk.Coin) {
	s.mustSucceedTx(leveragetypes.NewMsgBorrow(addr, asset))
}

func (s *E2ETest) maxBorrow(addr sdk.AccAddress, denom string) {
	s.mustSucceedTx(leveragetypes.NewMsgMaxBorrow(addr, denom))
}

func (s *E2ETest) repay(addr sdk.AccAddress, asset sdk.Coin) {
	s.mustSucceedTx(leveragetypes.NewMsgRepay(addr, asset))
}

func (s *E2ETest) liquidate(addr, target sdk.AccAddress, reward string, repay sdk.Coin) {
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
		},
	)
}

/*
func (s *E2ETest) TestMetokenSwapAndRedeem() {
	var prices []metoken.IndexPrices
	var index metoken.Index
	valAddr, err := s.Chain.Validators[0].KeyInfo.GetAddress()
	s.Require().NoError(err)
	expectedBalance := mocks.EmptyUSDIndexBalances(mocks.MeUSDDenom)

	s.Run(
		"create_stable_index", func() {
			tokens := []ltypes.Token{
				mocks.ValidToken(mocks.USDTBaseDenom, mocks.USDTSymbolDenom, 6),
				mocks.ValidToken(mocks.USDCBaseDenom, mocks.USDCSymbolDenom, 6),
				mocks.ValidToken(mocks.ISTBaseDenom, mocks.ISTSymbolDenom, 6),
			}

			err = grpc.LeverageRegistryUpdate(s.Umee, tokens, nil)
			s.Require().NoError(err)

			meUSD := mocks.StableIndex(mocks.MeUSDDenom)
			err = grpc.MetokenRegistryUpdate(s.Umee, []metoken.Index{meUSD}, nil)
			s.Require().NoError(err)

			prices = s.checkMetokenBalance(meUSD.Denom, expectedBalance)
		},
	)

	s.Run(
		"swap_100USDT_success", func() {
			index = s.getIndex(mocks.MeUSDDenom)

			hundredUSDT := sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(100_000000))
			fee := index.Fee.MinFee.MulInt(hundredUSDT.Amount).TruncateInt()

			assetSettings, i := index.AcceptedAsset(mocks.USDTBaseDenom)
			s.Require().True(i >= 0)

			amountToSwap := hundredUSDT.Amount.Sub(fee)
			amountToReserves := assetSettings.ReservePortion.MulInt(amountToSwap).TruncateInt()
			amountToLeverage := amountToSwap.Sub(amountToReserves)

			usdtPrice, err := prices[0].PriceByBaseDenom(mocks.USDTBaseDenom)
			s.Require().NoError(err)
			returned := usdtPrice.SwapRate.MulInt(amountToSwap).TruncateInt()

			s.executeSwap(valAddr, hundredUSDT, mocks.MeUSDDenom)

			expectedBalance.MetokenSupply.Amount = expectedBalance.MetokenSupply.Amount.Add(returned)
			usdtBalance, i := expectedBalance.AssetBalance(mocks.USDTBaseDenom)
			s.Require().True(i >= 0)
			usdtBalance.Fees = usdtBalance.Fees.Add(fee)
			usdtBalance.Reserved = usdtBalance.Reserved.Add(amountToReserves)
			usdtBalance.Leveraged = usdtBalance.Leveraged.Add(amountToLeverage)
			expectedBalance.SetAssetBalance(usdtBalance)

			prices = s.checkMetokenBalance(mocks.MeUSDDenom, expectedBalance)
		},
	)

	s.Run(
		"redeem_200meUSD_failure", func() {
			twoHundredsMeUSD := sdk.NewCoin(mocks.MeUSDDenom, sdkmath.NewInt(200_000000))

			s.executeRedeemWithFailure(
				valAddr,
				twoHundredsMeUSD,
				mocks.USDTBaseDenom,
				"not enough",
			)

			prices = s.checkMetokenBalance(mocks.MeUSDDenom, expectedBalance)
		},
	)
}
*/
