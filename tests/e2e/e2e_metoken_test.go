package e2e

import (
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/tests/grpc"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/metoken"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

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

	s.Run(
		"redeem_50meUSD_success", func() {
			fiftyMeUSD := sdk.NewCoin(mocks.MeUSDDenom, sdkmath.NewInt(50_000000))

			s.executeRedeemSuccess(valAddr, fiftyMeUSD, mocks.USDTBaseDenom)

			usdtPrice, err := prices[0].PriceByBaseDenom(mocks.USDTBaseDenom)
			s.Require().NoError(err)
			usdtToRedeem := usdtPrice.RedeemRate.MulInt(fiftyMeUSD.Amount).TruncateInt()
			fee := index.Fee.MinFee.MulInt(usdtToRedeem).TruncateInt()

			assetSettings, i := index.AcceptedAsset(mocks.USDTBaseDenom)
			s.Require().True(i >= 0)
			amountFromReserves := assetSettings.ReservePortion.MulInt(usdtToRedeem).TruncateInt()
			amountFromLeverage := usdtToRedeem.Sub(amountFromReserves)

			expectedBalance.MetokenSupply.Amount = expectedBalance.MetokenSupply.Amount.Sub(fiftyMeUSD.Amount)
			usdtBalance, i := expectedBalance.AssetBalance(mocks.USDTBaseDenom)
			s.Require().True(i >= 0)
			usdtBalance.Fees = usdtBalance.Fees.Add(fee)
			usdtBalance.Reserved = usdtBalance.Reserved.Sub(amountFromReserves)
			usdtBalance.Leveraged = usdtBalance.Leveraged.Sub(amountFromLeverage)
			expectedBalance.SetAssetBalance(usdtBalance)

			_ = s.checkMetokenBalance(mocks.MeUSDDenom, expectedBalance)
		},
	)
}

func (s *E2ETest) checkMetokenBalance(denom string, expectedBalance metoken.IndexBalances) []metoken.IndexPrices {
	var prices []metoken.IndexPrices
	s.Require().Eventually(
		func() bool {
			resp, err := s.Umee.QueryMetokenBalances(denom)
			if err != nil {
				return false
			}

			var exist bool
			for _, balance := range resp.IndexBalances {
				if balance.MetokenSupply.Denom == expectedBalance.MetokenSupply.Denom {
					exist = true
					s.Require().Equal(expectedBalance, balance)
					break
				}
			}

			s.Require().True(exist)
			prices = resp.Prices
			return true
		},
		30*time.Second,
		500*time.Millisecond,
	)

	return prices
}

func (s *E2ETest) getIndex(denom string) metoken.Index {
	index := metoken.Index{}
	s.Require().Eventually(
		func() bool {
			resp, err := s.Umee.QueryMetokenIndexes(denom)
			if err != nil {
				return false
			}

			var exist bool
			for _, indx := range resp.Registry {
				if indx.Denom == denom {
					exist = true
					index = indx
					break
				}
			}

			s.Require().True(exist)
			return true
		},
		30*time.Second,
		500*time.Millisecond,
	)

	return index
}

func (s *E2ETest) executeSwap(umeeAddr sdk.AccAddress, asset sdk.Coin, meTokenDenom string) {
	s.Require().Eventually(
		func() bool {
			err := s.Umee.TxMetokenSwap(umeeAddr, asset, meTokenDenom)
			if err != nil {
				fmt.Printf("SWAP ERR: %s\n", err.Error())
				return false
			}

			return true
		},
		30*time.Second,
		500*time.Millisecond,
	)
}

func (s *E2ETest) executeRedeemSuccess(umeeAddr sdk.AccAddress, meToken sdk.Coin, assetDenom string) {
	s.Require().Eventually(
		func() bool {
			err := s.Umee.TxMetokenRedeem(umeeAddr, meToken, assetDenom)
			if err != nil {
				fmt.Printf("REDEEM SCS ERR: %s\n", err.Error())
				return false
			}

			return true
		},
		30*time.Second,
		500*time.Millisecond,
	)
}

func (s *E2ETest) executeRedeemWithFailure(umeeAddr sdk.AccAddress, meToken sdk.Coin, assetDenom, errMsg string) {
	s.Require().Eventually(
		func() bool {
			err := s.Umee.TxMetokenRedeem(umeeAddr, meToken, assetDenom)
			if err != nil && strings.Contains(err.Error(), errMsg) {
				return true
			}

			fmt.Printf("REDEEM FAIL ERR: %s\n", err.Error())
			return false
		},
		30*time.Second,
		500*time.Millisecond,
	)
}
