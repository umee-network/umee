package metoken

import (
	"testing"

	"github.com/umee-network/umee/v6/util/coin"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestGenesisState_Validate(t *testing.T) {
	invalidRegistry := *DefaultGenesisState()
	invalidRegistry.Registry = []Index{
		NewIndex("token", sdkmath.ZeroInt(), 6, Fee{}, nil),
	}

	invalidBalance := *DefaultGenesisState()
	invalidBalance.Balances = []IndexBalances{
		{
			MetokenSupply: sdk.Coin{
				Denom:  "test",
				Amount: sdkmath.ZeroInt(),
			},
			AssetBalances: nil,
		},
	}

	tcs := []struct {
		name   string
		g      GenesisState
		errMsg string
	}{
		{"default genesis", *DefaultGenesisState(), ""},
		{
			"invalid registry",
			invalidRegistry,
			"meToken denom token should have the following format: me/<TokenName>",
		},
		{
			"invalid balances",
			invalidBalance,
			"meToken denom test should have the following format: me/<TokenName>",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				err := tc.g.Validate()
				if tc.errMsg != "" {
					assert.ErrorContains(t, err, tc.errMsg)
				} else {
					assert.NilError(t, err)
				}
			},
		)
	}
}

func TestIndexBalances_Validate(t *testing.T) {
	zeroInt := sdkmath.ZeroInt()
	negativeIntOne := sdkmath.NewInt(-1)

	invalidMetokenDenom := validIndexBalances()
	invalidMetokenDenom.MetokenSupply = sdk.NewCoin("test", sdkmath.ZeroInt())

	invalidMetokenAmount := validIndexBalances()
	invalidMetokenAmount.MetokenSupply = coin.Negative1("me/Token")

	invalidAssetInLeverage := validIndexBalances()
	invalidAssetInLeverage.SetAssetBalance(NewAssetBalance("USDT", negativeIntOne, zeroInt, zeroInt, zeroInt))

	invalidAssetInReserves := validIndexBalances()
	invalidAssetInReserves.SetAssetBalance(NewAssetBalance("USDT", zeroInt, negativeIntOne, zeroInt, zeroInt))

	invalidAssetInFees := validIndexBalances()
	invalidAssetInFees.SetAssetBalance(NewAssetBalance("USDT", zeroInt, zeroInt, negativeIntOne, zeroInt))

	invalidAssetInInterest := validIndexBalances()
	invalidAssetInInterest.SetAssetBalance(NewAssetBalance("USDT", zeroInt, zeroInt, zeroInt, negativeIntOne))

	duplicatedBalance := validIndexBalances()
	duplicatedBalance.AssetBalances = append(duplicatedBalance.AssetBalances, NewZeroAssetBalance("USDT"))

	tcs := []struct {
		name   string
		ib     IndexBalances
		errMsg string
	}{
		{"valid index balance", validIndexBalances(), ""},
		{
			"invalid meToken denom",
			invalidMetokenDenom,
			"meToken denom test should have the following format: me/<TokenName>",
		},
		{
			"invalid meToken amount",
			invalidMetokenAmount,
			"negative coin amount",
		},
		{
			"invalid assetInLeverage",
			invalidAssetInLeverage,
			"asset balance cannot be negative",
		},
		{
			"invalid assetInReserves",
			invalidAssetInReserves,
			"asset balance cannot be negative",
		},
		{
			"invalid assetInFee",
			invalidAssetInFees,
			"asset balance cannot be negative",
		},
		{
			"invalid assetInInterest",
			invalidAssetInInterest,
			"asset balance cannot be negative",
		},
		{
			"duplicated balance",
			duplicatedBalance,
			"duplicated balance",
		},
		{
			"valid index balance",
			NewIndexBalances(
				sdk.NewCoin("me/USD", sdkmath.ZeroInt()), []AssetBalance{
					NewZeroAssetBalance("USDT"),
				},
			),
			"",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				err := tc.ib.Validate()
				if tc.errMsg != "" {
					assert.ErrorContains(t, err, tc.errMsg)
				} else {
					assert.NilError(t, err)
				}
			},
		)
	}
}

func validIndexBalances() IndexBalances {
	zeroInt := sdkmath.ZeroInt()
	return IndexBalances{
		MetokenSupply: coin.Zero("me/USD"),
		AssetBalances: []AssetBalance{
			NewAssetBalance(
				"USDT",
				zeroInt,
				zeroInt,
				zeroInt,
				zeroInt,
			),
		},
	}
}
