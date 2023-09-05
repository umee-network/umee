package e2e

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/tests/grpc"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/metoken"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

func (s *E2ETest) TestMetokenSwapAndRedeem() {
	tokens := []ltypes.Token{
		mocks.ValidToken(mocks.USDTBaseDenom, mocks.USDTSymbolDenom, 6),
		mocks.ValidToken(mocks.USDCBaseDenom, mocks.USDCSymbolDenom, 6),
		mocks.ValidToken(mocks.ISTBaseDenom, mocks.ISTSymbolDenom, 6),
	}

	err := grpc.LeverageRegistryUpdate(s.Umee, tokens, nil)
	s.Require().NoError(err)

	meUSD := mocks.StableIndex(mocks.MeUSDDenom)
	err = grpc.MetokenRegistryUpdate(s.Umee, []metoken.Index{meUSD}, nil)
	s.Require().NoError(err)

	umeeAPIEndpoint := s.UmeeREST()
	prices := s.checkMetokenBalance(umeeAPIEndpoint, meUSD.Denom, mocks.EmptyUSDIndexBalances(mocks.MeUSDDenom))
	index := s.getIndex(umeeAPIEndpoint, meUSD.Denom)

	valAddr, err := s.Chain.Validators[0].KeyInfo.GetAddress()
	s.Require().NoError(err)

	hundredUSDT := sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(100_000000))
	fee := sdk.NewCoin(hundredUSDT.Denom, index.Fee.MinFee.MulInt(hundredUSDT.Amount).TruncateInt())

	assetSettings, i := index.AcceptedAsset(mocks.USDTBaseDenom)
	s.Require().True(i >= 0)

	amountToSwap := hundredUSDT.Amount.Sub(fee.Amount)
	amountToReserves := assetSettings.ReservePortion.MulInt(amountToSwap).TruncateInt()
	amountToLeverage := amountToSwap.Sub(amountToReserves)

	usdtPrice, err := prices[0].PriceByBaseDenom(mocks.USDTBaseDenom)
	s.Require().NoError(err)
	returned := usdtPrice.SwapRate.MulInt(amountToSwap).TruncateInt()

	s.executeSwap(umeeAPIEndpoint, valAddr.String(), hundredUSDT, mocks.MeUSDDenom)

	expectedBalance := mocks.EmptyUSDIndexBalances(mocks.MeUSDDenom)
	expectedBalance.MetokenSupply.Amount = expectedBalance.MetokenSupply.Amount.Add(returned)
	usdtBalance, i := expectedBalance.AssetBalance(mocks.USDTBaseDenom)
	s.Require().True(i >= 0)
	usdtBalance.Fees = usdtBalance.Fees.Add(fee.Amount)
	usdtBalance.Reserved.
}

func (s *E2ETest) checkMetokenBalance(
	umeeAPIEndpoint, denom string,
	expectedBalance metoken.IndexBalances,
) []metoken.IndexPrices {
	var prices []metoken.IndexPrices
	s.Require().Eventually(
		func() bool {
			resp, err := s.QueryMetokenBalances(umeeAPIEndpoint, denom)
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

func (s *E2ETest) getIndex(umeeAPIEndpoint, denom string) metoken.Index {
	index := metoken.Index{}
	s.Require().Eventually(
		func() bool {
			resp, err := s.QueryMetokenIndexes(umeeAPIEndpoint, denom)
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

func (s *E2ETest) executeSwap(umeeAPIEndpoint, umeeAddr string, asset sdk.Coin, meTokenDenom string) {
	s.Require().Eventually(
		func() bool {
			err := s.TxSwap(umeeAPIEndpoint, umeeAddr, asset, meTokenDenom)
			if err != nil {
				fmt.Printf("ERROR: %s", err.Error())
				return false
			}

			return true
		},
		30*time.Second,
		500*time.Millisecond,
	)
}
