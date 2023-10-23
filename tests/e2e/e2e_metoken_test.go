package e2e

import (
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/umee-network/umee/v6/tests/grpc"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/metoken"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

func (s *E2ETest) TestMetokenSwapAndRedeem() {
	var index metoken.Index
	testAddr := s.AccountAddr(0)
	expectedBalance := mocks.EmptyUSDIndexBalances(mocks.MeUSDDenom)

	s.Run(
		"create_stable_index", func() {
			tokens := []ltypes.Token{
				mocks.ValidToken(mocks.USDTBaseDenom, mocks.USDTSymbolDenom, 6),
				mocks.ValidToken(mocks.USDCBaseDenom, mocks.USDCSymbolDenom, 6),
				mocks.ValidToken(mocks.ISTBaseDenom, mocks.ISTSymbolDenom, 6),
			}

			err := grpc.LeverageRegistryUpdate(s.AccountClient(0), tokens, nil)
			s.Require().NoError(err)

			meUSD := mocks.StableIndex(mocks.MeUSDDenom)
			err = grpc.MetokenRegistryUpdate(s.AccountClient(0), []metoken.Index{meUSD}, nil)
			s.Require().NoError(err)

			s.checkMetokenBalance(testAddr.String(), mocks.MeUSDDenom)
		},
	)

	s.Run(
		"swap_100USDT_success", func() {
			index = s.getMetokenIndex(mocks.MeUSDDenom)
			hundredUSDT := sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(100_000000))
			fee := index.Fee.MinFee.MulInt(hundredUSDT.Amount).TruncateInt()

			assetSettings, i := index.AcceptedAsset(mocks.USDTBaseDenom)
			s.Require().True(i >= 0)

			amountToSwap := hundredUSDT.Amount.Sub(fee)
			amountToReserves := assetSettings.ReservePortion.MulInt(amountToSwap).TruncateInt()
			amountToLeverage := amountToSwap.Sub(amountToReserves)

			prices := s.getPrices(mocks.MeUSDDenom)
			usdtPrice, err := prices[0].PriceByBaseDenom(mocks.USDTBaseDenom)
			s.Require().NoError(err)
			returned := usdtPrice.SwapRate.MulInt(amountToSwap).TruncateInt()

			s.executeSwap(testAddr.String(), hundredUSDT, mocks.MeUSDDenom)

			expectedBalance.MetokenSupply.Amount = expectedBalance.MetokenSupply.Amount.Add(returned)
			usdtBalance, i := expectedBalance.AssetBalance(mocks.USDTBaseDenom)
			s.Require().True(i >= 0)
			usdtBalance.Fees = usdtBalance.Fees.Add(fee)
			usdtBalance.Reserved = usdtBalance.Reserved.Add(amountToReserves)
			usdtBalance.Leveraged = usdtBalance.Leveraged.Add(amountToLeverage)
			expectedBalance.SetAssetBalance(usdtBalance)

			s.checkMetokenBalance(testAddr.String(), mocks.MeUSDDenom)
		},
	)

	s.Run(
		"redeem_200meUSD_failure", func() {
			twoHundredsMeUSD := sdk.NewCoin(mocks.MeUSDDenom, sdkmath.NewInt(200_000000))

			s.executeRedeemWithFailure(
				testAddr.String(),
				twoHundredsMeUSD,
				mocks.USDTBaseDenom,
				"not enough",
			)

			s.checkMetokenBalance(testAddr.String(), mocks.MeUSDDenom)
		},
	)

	s.Run(
		"redeem_50meUSD_success", func() {
			prices := s.getPrices(mocks.MeUSDDenom)
			fiftyMeUSD := sdk.NewCoin(mocks.MeUSDDenom, sdkmath.NewInt(50_000000))

			s.executeRedeemSuccess(testAddr.String(), fiftyMeUSD, mocks.USDTBaseDenom)

			usdtToRedeem, err := prices[0].RedeemRate(fiftyMeUSD, mocks.USDTBaseDenom)
			s.Require().NoError(err)
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

			s.checkMetokenBalance(testAddr.String(), mocks.MeUSDDenom)
		},
	)
}

func (s *E2ETest) checkMetokenBalance(valAddr, denom string) {
	s.Require().Eventually(
		func() bool {
			resp, err := s.AccountClient(0).QueryMetokenIndexBalances(denom)
			if err != nil {
				return false
			}

			coins, err := s.QueryUmeeAllBalances(s.UmeeREST(), authtypes.NewModuleAddress(metoken.ModuleName).String())
			if err != nil {
				return false
			}

			for _, coin := range coins {
				var exist bool
				for _, balance := range resp.IndexBalances[0].AssetBalances {
					if balance.Denom == coin.Denom {
						exist = true
						expectedBalance := balance.Interest.Add(balance.Fees).Add(balance.Reserved)
						s.Require().Equal(coin.Amount, expectedBalance)
						continue
					}

					if "u/"+balance.Denom == coin.Denom {
						exist = true
						s.Require().Equal(coin.Amount, balance.Leveraged)
						continue
					}
				}
				s.Require().True(exist)
			}

			coins, err = s.QueryUmeeAllBalances(s.UmeeREST(), valAddr)
			if err != nil {
				return false
			}

			for _, coin := range coins {
				if coin.Denom == mocks.MeUSDDenom {
					s.Require().Equal(coin.Amount, resp.IndexBalances[0].MetokenSupply.Amount)
				}
			}

			return true
		},
		30*time.Second,
		500*time.Millisecond,
	)
}

func (s *E2ETest) getPrices(denom string) []metoken.IndexPrices {
	var prices []metoken.IndexPrices
	s.Require().Eventually(
		func() bool {
			resp, err := s.AccountClient(0).QueryMetokenIndexPrices(denom)
			if err != nil {
				return false
			}

			prices = resp.Prices
			return true
		},
		30*time.Second,
		500*time.Millisecond,
	)
	return prices
}

func (s *E2ETest) getMetokenIndex(denom string) metoken.Index {
	index := metoken.Index{}
	s.Require().Eventually(
		func() bool {
			resp, err := s.AccountClient(0).QueryMetokenIndexes(denom)
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

func (s *E2ETest) executeSwap(umeeAddr string, asset sdk.Coin, meTokenDenom string) {
	s.Require().Eventually(
		func() bool {
			return s.TxMetokenSwap(umeeAddr, asset, meTokenDenom) == nil
		},
		30*time.Second,
		500*time.Millisecond,
	)
}

func (s *E2ETest) executeRedeemSuccess(umeeAddr string, meToken sdk.Coin, assetDenom string) {
	s.Require().Eventually(
		func() bool {
			return s.TxMetokenRedeem(umeeAddr, meToken, assetDenom) == nil
		},
		30*time.Second,
		500*time.Millisecond,
	)
}

func (s *E2ETest) executeRedeemWithFailure(umeeAddr string, meToken sdk.Coin, assetDenom, errMsg string) {
	s.Require().Eventually(
		func() bool {
			err := s.TxMetokenRedeem(umeeAddr, meToken, assetDenom)
			if err != nil && strings.Contains(err.Error(), errMsg) {
				return true
			}

			return false
		},
		30*time.Second,
		500*time.Millisecond,
	)
}

func (s *E2ETest) TxMetokenSwap(umeeAddr string, asset sdk.Coin, meTokenDenom string) error {
	req := &metoken.MsgSwap{
		User:         umeeAddr,
		Asset:        asset,
		MetokenDenom: meTokenDenom,
	}

	return s.BroadcastTxWithRetry(req, s.AccountClient(0))
}

func (s *E2ETest) TxMetokenRedeem(umeeAddr string, meToken sdk.Coin, assetDenom string) error {
	req := &metoken.MsgRedeem{
		User:       umeeAddr,
		Metoken:    meToken,
		AssetDenom: assetDenom,
	}

	return s.BroadcastTxWithRetry(req, s.AccountClient(0))
}
