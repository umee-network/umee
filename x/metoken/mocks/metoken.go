package mocks

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/util/coin"

	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/metoken"
	otypes "github.com/umee-network/umee/v6/x/oracle/types"
)

const (
	USDTBaseDenom    = "ibc/223420B0E8CF9CC47BCAB816AB3A20AE162EED27C1177F4B2BC270C83E11AD8D"
	USDTSymbolDenom  = "USDT"
	USDCBaseDenom    = "ibc/49788C29CD84E08D25CA7BE960BC1F61E88FEFC6333F58557D236D693398466A"
	USDCSymbolDenom  = "USDC"
	ISTBaseDenom     = "ibc/BA460328D9ABA27E643A924071FDB3836E4CE8084C6D2380F25EFAB85CF8EB11"
	ISTSymbolDenom   = "IST"
	WBTCBaseDenom    = "ibc/153B97FE395140EAAA2D7CAC537AF1804AEC5F0595CBC5F1603094018D158C0C"
	WBTCSymbolDenom  = "WBTC"
	ETHBaseDenom     = "ibc/04CE51E6E02243E565AE676DD60336E48D455F8AAD0611FA0299A22FDAC448D6"
	ETHSymbolDenom   = "ETH"
	CMSTBaseDenom    = "ibc/31FA0BA043524F2EFBC9AB0539C43708B8FC549E4800E02D103DDCECAC5FF40C"
	CMSTSymbolDenom  = "CMST"
	MeUSDDenom       = "me/USD"
	MeNonStableDenom = "me/NonStable"
	TestDenom1       = "testDenom1"
	BondDenom        = "uumee"
	MeBondDenom      = "me/" + BondDenom
)

var (
	USDTPrice = sdkmath.LegacyMustNewDecFromStr("0.998")
	USDCPrice = sdkmath.LegacyMustNewDecFromStr("1.0")
	ISTPrice  = sdkmath.LegacyMustNewDecFromStr("1.02")
	CMSTPrice = sdkmath.LegacyMustNewDecFromStr("0.998")
	WBTCPrice = sdkmath.LegacyMustNewDecFromStr("27268.938478585498709550")
	ETHPrice  = sdkmath.LegacyMustNewDecFromStr("1851.789229542837161069")
)

func StableIndex(denom string) metoken.Index {
	return metoken.NewIndex(
		denom,
		sdkmath.NewInt(1_000_000_000_000),
		6,
		ValidFee(),
		[]metoken.AcceptedAsset{
			acceptedAsset(USDTBaseDenom, "0.33"),
			acceptedAsset(USDCBaseDenom, "0.34"),
			acceptedAsset(ISTBaseDenom, "0.33"),
		},
	)
}

func NonStableIndex(denom string) metoken.Index {
	return metoken.NewIndex(
		denom,
		sdkmath.NewInt(1_000_000_000_000),
		8,
		ValidFee(),
		[]metoken.AcceptedAsset{
			acceptedAsset(CMSTBaseDenom, "0.33"),
			acceptedAsset(WBTCBaseDenom, "0.34"),
			acceptedAsset(ETHBaseDenom, "0.33"),
		},
	)
}

func BondIndex() metoken.Index {
	return metoken.Index{
		Denom:     MeBondDenom,
		MaxSupply: sdkmath.NewInt(1000000_00000),
		Exponent:  6,
		Fee:       ValidFee(),
		AcceptedAssets: []metoken.AcceptedAsset{
			metoken.NewAcceptedAsset(
				BondDenom, sdkmath.LegacyMustNewDecFromStr("0.2"),
				sdkmath.LegacyMustNewDecFromStr("1.0"),
			),
		},
	}
}

func BondBalance() metoken.IndexBalances {
	return metoken.IndexBalances{
		MetokenSupply: coin.Zero(MeBondDenom),
		AssetBalances: []metoken.AssetBalance{
			{
				Denom:     BondDenom,
				Leveraged: sdkmath.ZeroInt(),
				Reserved:  sdkmath.ZeroInt(),
				Fees:      sdkmath.ZeroInt(),
				Interest:  sdkmath.ZeroInt(),
			},
		},
	}
}

func acceptedAsset(denom, targetAllocation string) metoken.AcceptedAsset {
	return metoken.NewAcceptedAsset(denom, sdkmath.LegacyMustNewDecFromStr("0.2"), sdkmath.LegacyMustNewDecFromStr(targetAllocation))
}

func ValidFee() metoken.Fee {
	return metoken.NewFee(
		sdkmath.LegacyMustNewDecFromStr("0.01"),
		sdkmath.LegacyMustNewDecFromStr("0.2"),
		sdkmath.LegacyMustNewDecFromStr("0.5"),
	)
}

func EmptyUSDIndexBalances(denom string) metoken.IndexBalances {
	return metoken.NewIndexBalances(
		sdk.NewCoin(denom, sdkmath.ZeroInt()),
		[]metoken.AssetBalance{
			metoken.NewZeroAssetBalance(USDTBaseDenom),
			metoken.NewZeroAssetBalance(USDCBaseDenom),
			metoken.NewZeroAssetBalance(ISTBaseDenom),
		},
	)
}

func EmptyNonStableIndexBalances(denom string) metoken.IndexBalances {
	return metoken.NewIndexBalances(
		sdk.NewCoin(denom, sdkmath.ZeroInt()),
		[]metoken.AssetBalance{
			metoken.NewZeroAssetBalance(CMSTBaseDenom),
			metoken.NewZeroAssetBalance(WBTCBaseDenom),
			metoken.NewZeroAssetBalance(ETHBaseDenom),
		},
	)
}

func ValidUSDIndexBalances(denom string) metoken.IndexBalances {
	return metoken.NewIndexBalances(
		sdk.NewCoin(denom, sdkmath.NewInt(4960_000000)),
		[]metoken.AssetBalance{
			metoken.NewAssetBalance(
				USDTBaseDenom,
				sdkmath.NewInt(960_000000),
				sdkmath.NewInt(240_000000),
				sdkmath.NewInt(34_000000),
				sdkmath.ZeroInt(),
			),
			metoken.NewAssetBalance(
				USDCBaseDenom,
				sdkmath.NewInt(608_000000),
				sdkmath.NewInt(152_000000),
				sdkmath.NewInt(28_000000),
				sdkmath.ZeroInt(),
			),
			metoken.NewAssetBalance(
				ISTBaseDenom,
				sdkmath.NewInt(2400_000000),
				sdkmath.NewInt(600_000000),
				sdkmath.NewInt(76_000000),
				sdkmath.ZeroInt(),
			),
		},
	)
}

// ValidPrices return 24 medians, each one with different prices
func ValidPrices() otypes.Prices {
	prices := otypes.Prices{}
	usdtPrice := USDTPrice.Sub(sdkmath.LegacyMustNewDecFromStr("0.24"))
	usdcPrice := USDCPrice.Sub(sdkmath.LegacyMustNewDecFromStr("0.24"))
	istPrice := ISTPrice.Sub(sdkmath.LegacyMustNewDecFromStr("0.24"))
	cmstPrice := CMSTPrice.Sub(sdkmath.LegacyMustNewDecFromStr("0.24"))
	wbtcPrice := WBTCPrice.Sub(sdkmath.LegacyMustNewDecFromStr("0.24"))
	ethPrice := ETHPrice.Sub(sdkmath.LegacyMustNewDecFromStr("0.24"))
	for i := 1; i <= 24; i++ {
		median := otypes.Price{
			ExchangeRateTuple: otypes.NewExchangeRateTuple(
				USDTSymbolDenom,
				usdtPrice.Add(sdkmath.LegacyMustNewDecFromStr("0.01").MulInt(sdkmath.NewInt(int64(i)))),
			),
			BlockNum: uint64(i),
		}
		prices = append(prices, median)
		median = otypes.Price{
			ExchangeRateTuple: otypes.NewExchangeRateTuple(
				USDCSymbolDenom,
				usdcPrice.Add(sdkmath.LegacyMustNewDecFromStr("0.01").MulInt(sdkmath.NewInt(int64(i)))),
			),
			BlockNum: uint64(i),
		}
		prices = append(prices, median)
		median = otypes.Price{
			ExchangeRateTuple: otypes.NewExchangeRateTuple(
				ISTSymbolDenom,
				istPrice.Add(sdkmath.LegacyMustNewDecFromStr("0.01").MulInt(sdkmath.NewInt(int64(i)))),
			),
			BlockNum: uint64(i),
		}
		prices = append(prices, median)
		median = otypes.Price{
			ExchangeRateTuple: otypes.NewExchangeRateTuple(
				CMSTSymbolDenom,
				cmstPrice.Add(sdkmath.LegacyMustNewDecFromStr("0.01").MulInt(sdkmath.NewInt(int64(i)))),
			),
			BlockNum: uint64(i),
		}
		prices = append(prices, median)
		median = otypes.Price{
			ExchangeRateTuple: otypes.NewExchangeRateTuple(
				WBTCSymbolDenom,
				wbtcPrice.Add(sdkmath.LegacyMustNewDecFromStr("0.01").MulInt(sdkmath.NewInt(int64(i)))),
			),
			BlockNum: uint64(i),
		}
		prices = append(prices, median)
		median = otypes.Price{
			ExchangeRateTuple: otypes.NewExchangeRateTuple(
				ETHSymbolDenom,
				ethPrice.Add(sdkmath.LegacyMustNewDecFromStr("0.01").MulInt(sdkmath.NewInt(int64(i)))),
			),
			BlockNum: uint64(i),
		}
		prices = append(prices, median)
	}

	return prices
}

// ValidPricesFunc return mock func for x/oracle
func ValidPricesFunc() func(ctx sdk.Context) otypes.Prices {
	return func(ctx sdk.Context) otypes.Prices {
		return ValidPrices()
	}
}

func ValidToken(baseDenom, symbolDenom string, exponent uint32) ltypes.Token {
	maxSupply := sdkmath.NewInt(1000000_00000000)
	if baseDenom == ETHBaseDenom {
		maxSupply = sdkmath.ZeroInt()
	}
	return ltypes.Token{
		BaseDenom:              baseDenom,
		SymbolDenom:            symbolDenom,
		Exponent:               exponent,
		ReserveFactor:          sdkmath.LegacyMustNewDecFromStr("0.25"),
		CollateralWeight:       sdkmath.LegacyMustNewDecFromStr("0.5"),
		LiquidationThreshold:   sdkmath.LegacyMustNewDecFromStr("0.51"),
		BaseBorrowRate:         sdkmath.LegacyMustNewDecFromStr("0.01"),
		KinkBorrowRate:         sdkmath.LegacyMustNewDecFromStr("0.05"),
		MaxBorrowRate:          sdkmath.LegacyMustNewDecFromStr("1"),
		KinkUtilization:        sdkmath.LegacyMustNewDecFromStr("0.75"),
		LiquidationIncentive:   sdkmath.LegacyMustNewDecFromStr("0.05"),
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdkmath.LegacyMustNewDecFromStr("1"),
		MaxSupplyUtilization:   sdkmath.LegacyMustNewDecFromStr("1"),
		MinCollateralLiquidity: sdkmath.LegacyMustNewDecFromStr("0.05"),
		MaxSupply:              maxSupply,
		HistoricMedians:        24,
	}
}
