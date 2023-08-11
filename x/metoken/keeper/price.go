package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v6/x/metoken"
	otypes "github.com/umee-network/umee/v6/x/oracle/types"
)

var usdExponent = uint32(0)

// Prices calculates meToken price as an avg of median prices of all index accepted assets.
func (k Keeper) Prices(index metoken.Index) (metoken.IndexPrices, error) {
	indexPrices := metoken.NewIndexPrices()
	meTokenDenom := index.Denom

	supply, err := k.IndexBalances(index.Denom)
	if err != nil {
		return indexPrices, err
	}

	allPrices := k.oracleKeeper.AllMedianPrices(*k.ctx)

	// calculate the total assets value in the index balances
	totalAssetsUSDValue := sdk.ZeroDec()
	for _, aa := range index.AcceptedAssets {
		// get token settings from leverageKeeper to use the symbol_denom
		tokenSettings, err := k.leverageKeeper.GetTokenSettings(*k.ctx, aa.Denom)
		if err != nil {
			return indexPrices, err
		}

		assetPrice, err := latestPrice(allPrices, tokenSettings.SymbolDenom)
		if err != nil {
			return indexPrices, err
		}
		indexPrices.SetPrice(aa.Denom, assetPrice, tokenSettings.Exponent)

		// if no meTokens were minted, the totalAssetValue is the sum of all the assets prices.
		// otherwise is the sum of the value of all the assets in the index.
		if supply.MetokenSupply.IsZero() {
			totalAssetsUSDValue = totalAssetsUSDValue.Add(assetPrice)
		} else {
			i, balance := supply.AssetBalance(aa.Denom)
			if i < 0 {
				return indexPrices, sdkerrors.ErrNotFound.Wrapf("balance for denom %s not found", aa.Denom)
			}

			assetUSDValue, err := valueInUSD(balance.AvailableSupply(), assetPrice, tokenSettings.Exponent)
			if err != nil {
				return indexPrices, err
			}
			totalAssetsUSDValue = totalAssetsUSDValue.Add(assetUSDValue)
		}
	}

	if supply.MetokenSupply.IsZero() {
		// if no meTokens were minted, the meTokenPrice is totalAssetsUSDValue divided by accepted assets quantity
		indexPrices.SetPrice(
			meTokenDenom,
			totalAssetsUSDValue.QuoInt(sdkmath.NewInt(int64(len(index.AcceptedAssets)))),
			index.Exponent,
		)
	} else {
		// otherwise, the meTokenPrice is totalAssetsUSDValue divided by meTokens minted quantity
		meTokenPrice, err := priceInUSD(supply.MetokenSupply.Amount, totalAssetsUSDValue, index.Exponent)
		if err != nil {
			return indexPrices, err
		}
		indexPrices.SetPrice(meTokenDenom, meTokenPrice, index.Exponent)
	}

	return indexPrices, nil
}

// latestPrice from the list of medians, based on the block number.
func latestPrice(prices otypes.Prices, symbolDenom string) (sdk.Dec, error) {
	denomPrices := prices.FilterByDenom(symbolDenom)

	if len(denomPrices) == 0 {
		return sdk.Dec{}, fmt.Errorf("price not found in oracle for denom %s", symbolDenom)
	}

	return denomPrices[len(denomPrices)-1].ExchangeRateTuple.ExchangeRate, nil
}

// valueInUSD given a specific amount, price and exponent
func valueInUSD(amount sdkmath.Int, assetPrice sdk.Dec, assetExponent uint32) (sdk.Dec, error) {
	exponentFactor, err := metoken.ExponentFactor(assetExponent, usdExponent)
	if err != nil {
		return sdk.Dec{}, err
	}
	return exponentFactor.MulInt(amount).Mul(assetPrice), nil
}

// priceInUSD given a specific amount, totalValue and exponent
func priceInUSD(amount sdkmath.Int, totalValue sdk.Dec, assetExponent uint32) (sdk.Dec, error) {
	exponentFactor, err := metoken.ExponentFactor(assetExponent, usdExponent)
	if err != nil {
		return sdk.Dec{}, err
	}

	return totalValue.Quo(exponentFactor.MulInt(amount)), nil
}
