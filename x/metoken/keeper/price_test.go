package keeper

import (
	"testing"

	"github.com/umee-network/umee/v6/x/metoken/mocks"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v6/x/metoken"
)

func TestIndexPrices_Prices(t *testing.T) {
	o := NewOracleMock()
	l := NewLeverageMock()
	k := initMeUSDKeeper(t, nil, l, o)
	index, err := k.RegisteredIndex(mocks.MeUSDDenom)
	require.NoError(t, err)

	// inexisting asset case
	ip, err := k.Prices(index)
	require.NoError(t, err)
	_, err = ip.Price("inexistingAsset")
	require.ErrorContains(t, err, "not found")

	// confirm all the asset prices are set correctly
	price, err := ip.Price(mocks.USDTBaseDenom)
	require.NoError(t, err)
	require.Equal(t, price.Exponent, uint32(6))
	require.True(t, price.Price.Equal(mocks.USDTPrice))

	price, err = ip.Price(mocks.USDCBaseDenom)
	require.NoError(t, err)
	require.Equal(t, price.Exponent, uint32(6))
	require.True(t, price.Price.Equal(mocks.USDCPrice))

	price, err = ip.Price(mocks.ISTBaseDenom)
	require.NoError(t, err)
	require.Equal(t, price.Exponent, uint32(6))
	require.True(t, price.Price.Equal(mocks.ISTPrice))

	// case with 4960 meTokens minted
	// metoken_price = (supply1 * price1 + supply2 * price2 + supplyN * priceN) / metokens_minted
	// metoken_price = (1200 * 0.998 + 760 * 1.0 + 3000 * 1.02) / 4960 = 1.011612903225806452
	price, err = ip.Price(mocks.MeUSDDenom)
	require.NoError(t, err)
	require.Equal(t, price.Exponent, uint32(6))
	require.True(t, price.Price.Equal(sdk.MustNewDecFromStr("1.011612903225806452")))

	// case with no meTokens minted
	balance := mocks.EmptyUSDIndexBalances(mocks.MeUSDDenom)
	err = k.setIndexBalances(balance)
	require.NoError(t, err)

	// case with 0 meTokens minted
	// metoken_price = (price1 + price2 + priceN) / N
	// metoken_price = (0.998 + 1.0 + 1.02) / 3 = 1.006
	ip, err = k.Prices(index)
	require.NoError(t, err)
	price, err = ip.Price(mocks.MeUSDDenom)
	require.NoError(t, err)
	require.Equal(t, price.Exponent, uint32(6))
	require.True(t, price.Price.Equal(sdk.MustNewDecFromStr("1.006")))
}

func TestIndexPrices_Convert(t *testing.T) {
	o := NewOracleMock()
	l := NewLeverageMock()
	k := initMeUSDKeeper(t, nil, l, o)

	// same exponent cases
	index, err := k.RegisteredIndex(mocks.MeUSDDenom)
	require.NoError(t, err)
	ip, err := k.Prices(index)
	require.NoError(t, err)

	meTokenPrice, err := ip.Price(mocks.MeUSDDenom)
	require.NoError(t, err)

	// convert 20 USDC to meUSD
	usdcPrice, err := ip.Price(mocks.USDCBaseDenom)
	require.NoError(t, err)

	coin := sdk.NewCoin(mocks.USDCBaseDenom, sdkmath.NewInt(20_000000))
	result, err := ip.SwapRate(coin, mocks.MeUSDDenom)
	require.NoError(t, err)
	require.True(t, result.Equal(usdcPrice.Price.Quo(meTokenPrice.Price).MulInt(coin.Amount).TruncateInt()))

	// convert 130 meUSD to IST
	istPrice, err := ip.Price(mocks.ISTBaseDenom)
	require.NoError(t, err)

	coin = sdk.NewCoin(mocks.MeUSDDenom, sdkmath.NewInt(130_000000))
	result, err = ip.SwapRate(coin, mocks.ISTBaseDenom)
	require.NoError(t, err)
	require.True(t, result.Equal(meTokenPrice.Price.Quo(istPrice.Price).MulInt(coin.Amount).TruncateInt()))

	// diff exponent cases
	// change exponents in leverage
	usdtSettings := l.tokens[mocks.USDCBaseDenom]
	usdtSettings.Exponent = 8
	l.tokens[mocks.USDCBaseDenom] = usdtSettings
	istSettings := l.tokens[mocks.ISTBaseDenom]
	istSettings.Exponent = 4
	l.tokens[mocks.ISTBaseDenom] = istSettings

	// change supply given the new exponents
	supply, err := k.IndexBalances(mocks.MeUSDDenom)
	require.NoError(t, err)
	i, usdtBalance := supply.AssetBalance(mocks.USDCBaseDenom)
	require.True(t, i >= 0)
	usdtBalance.Reserved = usdtBalance.Reserved.Mul(sdkmath.NewInt(100))
	usdtBalance.Leveraged = usdtBalance.Leveraged.Mul(sdkmath.NewInt(100))
	supply.SetAssetBalance(usdtBalance)

	i, istBalance := supply.AssetBalance(mocks.ISTBaseDenom)
	require.True(t, i >= 0)
	istBalance.Reserved = istBalance.Reserved.Quo(sdkmath.NewInt(100))
	istBalance.Leveraged = istBalance.Leveraged.Quo(sdkmath.NewInt(100))
	supply.SetAssetBalance(istBalance)

	err = k.setIndexBalances(supply)

	ip, err = k.Prices(index)
	require.NoError(t, err)

	// convert 115 USDC to meUSD
	usdcPrice, err = ip.Price(mocks.USDCBaseDenom)
	require.NoError(t, err)

	coin = sdk.NewCoin(mocks.USDCBaseDenom, sdkmath.NewInt(115_000000))
	result, err = ip.SwapRate(coin, mocks.MeUSDDenom)
	require.NoError(t, err)
	require.True(
		t, result.Equal(
			usdcPrice.Price.Quo(meTokenPrice.Price).MulInt(coin.Amount).Mul(
				sdk.
					MustNewDecFromStr("0.01"),
			).TruncateInt(),
		),
	)

	// convert 1783.91827311 meUSD to IST
	istPrice, err = ip.Price(mocks.ISTBaseDenom)
	require.NoError(t, err)

	coin = sdk.NewCoin(mocks.MeUSDDenom, sdkmath.NewInt(1783_91827311))
	result, err = ip.SwapRate(coin, mocks.ISTBaseDenom)
	require.NoError(t, err)
	require.True(
		t, result.Equal(
			meTokenPrice.Price.Quo(istPrice.Price).MulInt(coin.Amount).QuoInt(
				sdkmath.NewInt(
					100,
				),
			).TruncateInt(),
		),
	)
}

func meUSDIndexPricesAdjustedToBalance(t *testing.T, balance metoken.IndexBalances) metoken.IndexPrices {
	i, usdtSupply := balance.AssetBalance(mocks.USDTBaseDenom)
	require.True(t, i >= 0)
	i, usdcSupply := balance.AssetBalance(mocks.USDCBaseDenom)
	require.True(t, i >= 0)
	i, istSupply := balance.AssetBalance(mocks.ISTBaseDenom)
	require.True(t, i >= 0)

	prices := metoken.NewIndexPrices()
	prices.SetPrice(mocks.USDTBaseDenom, mocks.USDTPrice, 6)
	prices.SetPrice(mocks.USDCBaseDenom, mocks.USDCPrice, 6)
	prices.SetPrice(mocks.ISTBaseDenom, mocks.ISTPrice, 6)
	prices.SetPrice(
		mocks.MeUSDDenom,
		mocks.USDTPrice.MulInt(usdtSupply.AvailableSupply()).Add(
			mocks.USDCPrice.MulInt(
				usdcSupply.
					AvailableSupply(),
			),
		).Add(mocks.ISTPrice.MulInt(istSupply.AvailableSupply())).QuoInt(balance.MetokenSupply.Amount),
		6,
	)

	return prices
}
