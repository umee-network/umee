package intest

import (
	"testing"

	"gotest.tools/v3/assert"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/metoken"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

func TestQuerier_Params(t *testing.T) {
	s := initTestSuite(t, nil, nil)
	querier, ctx := s.queryClient, s.ctx

	resp, err := querier.Params(ctx, nil)
	assert.NilError(t, err)
	assert.Check(t, resp.Params.RebalancingFrequency > 0)
	assert.Check(t, resp.Params.ClaimingFrequency > 0)
}

func TestQuerier_Indexes(t *testing.T) {
	index1 := mocks.StableIndex(mocks.MeUSDDenom)
	index2 := mocks.StableIndex("me/EUR")

	s := initTestSuite(t, []metoken.Index{index1, index2}, nil)
	querier, ctx := s.queryClient, s.ctx

	tcs := []struct {
		name          string
		denom         string
		expIndexCount int
		expErr        string
	}{
		{
			"get all indexes",
			"",
			2,
			"",
		},
		{
			"get index found",
			mocks.MeUSDDenom,
			1,
			"",
		},
		{
			"get index not found",
			"me/Test",
			0,
			"not found",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				resp, err := querier.Indexes(
					ctx, &metoken.QueryIndexes{
						MetokenDenom: tc.denom,
					},
				)
				if len(tc.expErr) == 0 {
					assert.NilError(t, err)
					assert.Check(t, tc.expIndexCount == len(resp.Registry))
				} else {
					assert.ErrorContains(t, err, tc.expErr)
				}
			},
		)
	}
}

func TestQuerier_Balances(t *testing.T) {
	index1 := mocks.StableIndex(mocks.MeUSDDenom)
	index2 := mocks.NonStableIndex(mocks.MeNonStableDenom)

	balance1 := mocks.ValidUSDIndexBalances(mocks.MeUSDDenom)
	balance2 := mocks.EmptyNonStableIndexBalances(mocks.MeNonStableDenom)

	s := initTestSuite(t, []metoken.Index{index1, index2}, []metoken.IndexBalances{balance1, balance2})
	querier, ctx := s.queryClient, s.ctx

	tcs := []struct {
		name            string
		denom           string
		expBalanceCount int
		expErr          string
	}{
		{
			"get all balances",
			"",
			2,
			"",
		},
		{
			"get balance found",
			mocks.MeUSDDenom,
			1,
			"",
		},
		{
			"get balance not found",
			"me/Test",
			0,
			"not found",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				resp, err := querier.IndexBalances(
					ctx, &metoken.QueryIndexBalances{
						MetokenDenom: tc.denom,
					},
				)
				if len(tc.expErr) > 0 {
					assert.ErrorContains(t, err, tc.expErr)
				} else {
					assert.NilError(t, err)
					assert.Check(t, tc.expBalanceCount == len(resp.IndexBalances))
					assert.Check(t, tc.expBalanceCount == len(resp.Prices))
				}
			},
		)
	}
}

type feeTestCase struct {
	name  string
	asset sdk.Coin
	denom string
}

func TestQuerier_SwapFee_meUSD(t *testing.T) {
	index := mocks.StableIndex(mocks.MeUSDDenom)
	balances := mocks.ValidUSDIndexBalances(mocks.MeUSDDenom)

	s := initTestSuite(t, []metoken.Index{index}, []metoken.IndexBalances{balances})
	querier, ctx := s.queryClient, s.ctx

	// set prices
	prices := metoken.EmptyIndexPrices(index)
	prices.SetPrice(
		metoken.AssetPrice{
			BaseDenom:   mocks.USDTBaseDenom,
			SymbolDenom: mocks.USDTSymbolDenom,
			Price:       mocks.USDTPrice,
			Exponent:    6,
		},
	)
	prices.SetPrice(
		metoken.AssetPrice{
			BaseDenom:   mocks.USDCBaseDenom,
			SymbolDenom: mocks.USDCSymbolDenom,
			Price:       mocks.USDCPrice,
			Exponent:    6,
		},
	)
	prices.SetPrice(
		metoken.AssetPrice{
			BaseDenom:   mocks.ISTBaseDenom,
			SymbolDenom: mocks.ISTSymbolDenom,
			Price:       mocks.ISTPrice,
			Exponent:    6,
		},
	)

	totalValue := sdk.ZeroDec()
	values := make(map[string]sdk.Dec)
	for _, balance := range balances.AssetBalances {
		// calculate total asset supply (leveraged + reserved)
		assetSupply := balance.AvailableSupply()
		// get asset PriceByBaseDenom
		assetPrice, err := prices.PriceByBaseDenom(balance.Denom)
		assert.NilError(t, err)
		// calculate asset value
		assetValue := assetPrice.Price.MulInt(assetSupply)

		// add asset value to the totalValue
		totalValue = totalValue.Add(assetValue)
		// calculate every asset balance value
		values[balance.Denom] = assetValue
	}

	tcs := []feeTestCase{
		{
			name:  "10 USDT swap",
			asset: sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(10_000000)),
			denom: mocks.MeUSDDenom,
		},
		{
			name:  "750 USDC swap",
			asset: sdk.NewCoin(mocks.USDCBaseDenom, sdkmath.NewInt(750_000000)),
			denom: mocks.MeUSDDenom,
		},
		{
			name:  "1876 IST swap",
			asset: sdk.NewCoin(mocks.ISTBaseDenom, sdkmath.NewInt(1876_000000)),
			denom: mocks.MeUSDDenom,
		},
	}

	for _, tc := range tcs {
		req := &metoken.QuerySwapFee{
			Asset:        tc.asset,
			MetokenDenom: tc.denom,
		}
		denom := tc.asset.Denom

		resp, err := querier.SwapFee(ctx, req)
		assert.NilError(t, err)

		// current_allocation = asset_value / total_value
		currentAllocation := values[denom].Quo(totalValue)
		aa, i := index.AcceptedAsset(denom)
		assert.Check(t, i >= 0)
		targetAllocation := aa.TargetAllocation

		// swap_delta_allocation = (current_allocation - target_allocation) / target_allocation
		swapDeltaAllocation := currentAllocation.Sub(targetAllocation).Quo(targetAllocation)

		// fee = delta_allocation * balanced_fee + balanced_fee
		fee := swapDeltaAllocation.Mul(index.Fee.BalancedFee).Add(index.Fee.BalancedFee)

		// swap_fee = fee * amount
		result := fee.MulInt(tc.asset.Amount).TruncateInt()

		assert.Check(t, result.Equal(resp.Asset.Amount))
	}
}

func TestQuerier_RedeemFee_meUSD(t *testing.T) {
	index := mocks.StableIndex(mocks.MeUSDDenom)
	balances := mocks.ValidUSDIndexBalances(mocks.MeUSDDenom)

	s := initTestSuite(t, []metoken.Index{index}, []metoken.IndexBalances{balances})
	querier, ctx, app := s.queryClient, s.ctx, s.app

	// set prices
	prices, err := app.MetokenKeeperB.Keeper(&ctx).Prices(index)
	assert.NilError(t, err)

	tcs := []feeTestCase{
		{
			name:  "20 meUSD to USDT redemption",
			asset: sdk.NewCoin(mocks.MeUSDDenom, sdkmath.NewInt(20_000000)),
			denom: mocks.USDTBaseDenom,
		},
		{
			name:  "444 meUSD to USDC redemption",
			asset: sdk.NewCoin(mocks.MeUSDDenom, sdkmath.NewInt(444_000000)),
			denom: mocks.USDCBaseDenom,
		},
		{
			name:  "1267 meUSD to IST redemption",
			asset: sdk.NewCoin(mocks.MeUSDDenom, sdkmath.NewInt(1267_000000)),
			denom: mocks.ISTBaseDenom,
		},
	}

	for _, tc := range tcs {
		req := &metoken.QueryRedeemFee{
			Metoken:    tc.asset,
			AssetDenom: tc.denom,
		}
		resp, err := querier.RedeemFee(ctx, req)
		assert.NilError(t, err)

		price, err := prices.PriceByBaseDenom(tc.denom)
		assert.NilError(t, err)
		balance, i := balances.AssetBalance(tc.denom)
		assert.Check(t, i >= 0)
		supply := balance.AvailableSupply()

		// current_allocation = asset_value / total_value
		currentAllocation := price.Price.MulInt(supply).Quo(prices.Price.MulInt(balances.MetokenSupply.Amount))
		aa, i := index.AcceptedAsset(tc.denom)
		assert.Check(t, i >= 0)
		targetAllocation := aa.TargetAllocation

		// redeem_delta_allocation = (target_allocation - current_allocation) / target_allocation
		redeemDeltaAllocation := targetAllocation.Sub(currentAllocation).Quo(targetAllocation)

		// fee = delta_allocation * balanced_fee + balanced_fee
		fee := redeemDeltaAllocation.Mul(index.Fee.BalancedFee).Add(index.Fee.BalancedFee)

		// exchange_rate = metoken_price / asset_price
		exchangeRate := prices.Price.Quo(price.Price)

		// asset_to_redeem = exchange_rate * asset_amount
		toRedeem := exchangeRate.MulInt(tc.asset.Amount).TruncateInt()

		// total_fee = asset_to_redeem * fee
		totalFee := fee.MulInt(toRedeem).TruncateInt()

		assert.Check(t, totalFee.Equal(resp.Asset.Amount))
	}
}

func TestQuerier_IndexPrices(t *testing.T) {
	// Within these cases we are testing grpc functionality,
	// Exact prices are tested in their unit tests:
	// https://github.com/umee-network/umee/blob/main/x/metoken/keeper/price_test.go
	stableIndex := mocks.StableIndex(mocks.MeUSDDenom)
	nonStableIndex := mocks.NonStableIndex(mocks.MeNonStableDenom)

	stableBalance := mocks.EmptyUSDIndexBalances(mocks.MeUSDDenom)
	nonStableBalance := mocks.EmptyNonStableIndexBalances(mocks.MeNonStableDenom)

	s := initTestSuite(
		t,
		[]metoken.Index{stableIndex, nonStableIndex},
		[]metoken.IndexBalances{stableBalance, nonStableBalance},
	)
	querier, ctx := s.queryClient, s.ctx

	tcs := []struct {
		name          string
		denom         string
		expPriceCount int
		expErr        string
	}{
		{
			"invalid meToken denom",
			"invalidDenom",
			0,
			"should have the following format: me/<TokenName>",
		},
		{
			"index not found",
			"me/NotFound",
			0,
			"not found",
		},
		{
			"get meUSD price",
			mocks.MeUSDDenom,
			1,
			"",
		},
		{
			"get all prices",
			"",
			2,
			"",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				resp, err := querier.IndexPrices(
					ctx, &metoken.QueryIndexPrices{
						MetokenDenom: tc.denom,
					},
				)
				if len(tc.expErr) == 0 {
					assert.NilError(t, err)
					assert.Check(t, tc.expPriceCount == len(resp.Prices))
					for _, i := range resp.Prices {
						for _, a := range i.Assets {
							assert.Check(t, a.SwapRate.GT(sdk.ZeroDec()))
							assert.Check(t, a.RedeemRate.GT(sdk.ZeroDec()))
						}
					}
				} else {
					assert.ErrorContains(t, err, tc.expErr)
				}
			},
		)
	}
}
