package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/x/oracle/types"
)

var (
	umeeDenom = types.Denom{
		BaseDenom:   app.BondDenom,
		SymbolDenom: app.DisplayDenom,
		Exponent:    6,
	}
	lunaDenom = types.Denom{
		BaseDenom:   denomLunaIBC,
		SymbolDenom: "ULUNA",
		Exponent:    6,
	}
	atomDenom = types.Denom{
		BaseDenom:   denomAtomIBC,
		SymbolDenom: "UATOM",
		Exponent:    6,
	}
)

func TestDenomString(t *testing.T) {
	testCases := []struct {
		denom       types.Denom
		expectedStr string
	}{
		{
			denom:       umeeDenom,
			expectedStr: "base_denom: uumee\nsymbol_denom: UMEE\nexponent: 6\n",
		},
		{
			denom:       lunaDenom,
			expectedStr: "base_denom: ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0\nsymbol_denom: ULUNA\nexponent: 6\n",
		},
		{
			denom:       atomDenom,
			expectedStr: "base_denom: ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2\nsymbol_denom: UATOM\nexponent: 6\n",
		},
	}

	for _, testCase := range testCases {
		require.Equal(t, testCase.expectedStr, testCase.denom.String())
	}
}

func TestDenomEqual(t *testing.T) {
	testCases := []struct {
		denom         types.Denom
		denomCompared types.Denom
		equal         bool
	}{
		{
			denom:         umeeDenom,
			denomCompared: umeeDenom,
			equal:         true,
		},
		{
			denom:         umeeDenom,
			denomCompared: lunaDenom,
			equal:         false,
		},
		{
			denom:         lunaDenom,
			denomCompared: lunaDenom,
			equal:         true,
		},
		{
			denom:         atomDenom,
			denomCompared: atomDenom,
			equal:         true,
		},
		{
			denom:         atomDenom,
			denomCompared: lunaDenom,
			equal:         false,
		},
	}

	for _, testCase := range testCases {
		require.Equal(t, testCase.equal, testCase.denom.Equal(&testCase.denomCompared))
	}
}

func TestDenomListString(t *testing.T) {
	testCases := []struct {
		denomList   types.DenomList
		expectedStr string
	}{
		{
			denomList:   types.DenomList{umeeDenom},
			expectedStr: "base_denom: uumee\nsymbol_denom: UMEE\nexponent: 6",
		},
		{
			denomList:   types.DenomList{atomDenom, lunaDenom},
			expectedStr: "base_denom: ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2\nsymbol_denom: UATOM\nexponent: 6\n\nbase_denom: ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0\nsymbol_denom: ULUNA\nexponent: 6",
		},
	}

	for _, testCase := range testCases {
		require.Equal(t, testCase.expectedStr, testCase.denomList.String())
	}
}

func TestDenomListContains(t *testing.T) {
	testCases := []struct {
		denomList    types.DenomList
		denomSymbol  string
		symbolInList bool
	}{
		{
			denomList:    types.DenomList{umeeDenom},
			denomSymbol:  umeeDenom.SymbolDenom,
			symbolInList: true,
		},
		{
			denomList:    types.DenomList{umeeDenom},
			denomSymbol:  lunaDenom.SymbolDenom,
			symbolInList: false,
		},
		{
			denomList:    types.DenomList{umeeDenom, atomDenom},
			denomSymbol:  lunaDenom.SymbolDenom,
			symbolInList: false,
		},
		{
			denomList:    types.DenomList{umeeDenom, atomDenom},
			denomSymbol:  atomDenom.SymbolDenom,
			symbolInList: true,
		},
		{
			denomList:    types.DenomList{umeeDenom, atomDenom, lunaDenom},
			denomSymbol:  lunaDenom.SymbolDenom,
			symbolInList: true,
		},
	}

	for _, testCase := range testCases {
		require.Equal(t, testCase.symbolInList, testCase.denomList.Contains(testCase.denomSymbol))
	}
}
