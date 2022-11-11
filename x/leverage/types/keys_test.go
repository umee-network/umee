package types_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/umee-network/umee/v3/app/params"
	"github.com/umee-network/umee/v3/x/leverage/types"
)

func TestAddressFromKey(t *testing.T) {
	address := sdk.AccAddress([]byte("addr________________"))
	key := types.CreateAdjustedBorrowKey(address, appparams.BondDenom)
	expectedAddress := types.AddressFromKey(key, types.KeyPrefixAdjustedBorrow)

	require.Equal(t, address, expectedAddress)

	address = sdk.AccAddress([]byte("anotherAddr________________"))
	key = types.CreateCollateralAmountKeyNoDenom(address)
	expectedAddress = types.AddressFromKey(key, types.KeyPrefixAdjustedBorrow)

	require.Equal(t, address, expectedAddress)
}

func TestDenomFromKeyWithAddress(t *testing.T) {
	address := sdk.AccAddress([]byte("addr________________"))
	denom := appparams.BondDenom
	key := types.CreateAdjustedBorrowKey(address, denom)
	expectedDenom := types.DenomFromKeyWithAddress(key, types.KeyPrefixAdjustedBorrow)

	require.Equal(t, denom, expectedDenom)

	uDenom := fmt.Sprintf("u%s", denom)
	key = types.CreateCollateralAmountKey(address, uDenom)
	expectedDenom = types.DenomFromKeyWithAddress(key, types.KeyPrefixCollateralAmount)

	require.Equal(t, uDenom, expectedDenom)
}

func TestDenomFromKey(t *testing.T) {
	denom := appparams.BondDenom
	key := types.CreateReserveAmountKey(denom)
	expectedDenom := types.DenomFromKey(key, types.KeyPrefixReserveAmount)

	require.Equal(t, denom, expectedDenom)

	uDenom := fmt.Sprintf("u%s", denom)
	key = types.CreateReserveAmountKey(uDenom)
	expectedDenom = types.DenomFromKey(key, types.KeyPrefixReserveAmount)

	require.Equal(t, uDenom, expectedDenom)
}

func TestGetKeys(t *testing.T) {
	type testCase struct {
		actual      []byte
		expected    [][]byte
		description string
	}

	addr := sdk.AccAddress("addr________________") // length: 20
	addrbytes := []byte{0x61, 0x64, 0x64, 0x72, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f}
	uumeebytes := []byte{0x75, 0x75, 0x6d, 0x65, 0x65}                              // uumee
	ibcabcdbytes := []byte{0x69, 0x62, 0x63, 0x2f, 0x61, 0x62, 0x63, 0x64}          // ibc/abcd
	uibcbytes := []byte{0x75, 0x2f, 0x69, 0x62, 0x63, 0x2f, 0x61, 0x62, 0x63, 0x64} // u/ibc/abcd

	testCases := []testCase{
		{
			types.CreateRegisteredTokenKey("uumee"),
			[][]byte{
				{0x01},     // prefix
				uumeebytes, // uumee
				{0x00},     // null terminator
			},
			"registered token key (uumee)",
		},
		{
			types.CreateRegisteredTokenKey("ibc/abcd"),
			[][]byte{
				{0x01},       // prefix
				ibcabcdbytes, // ibc/abcd
				{0x00},       // null terminator
			},
			"registered token key (ibc/abcd)",
		},
		{
			types.CreateAdjustedBorrowKey(addr, "uumee"),
			[][]byte{
				{0x02},     // prefix
				{0x14},     // address length prefix = 20
				addrbytes,  // addr________________
				uumeebytes, // uumee
				{0x00},     // null terminator
			},
			"adjusted borrow key (uumee)",
		},
		{
			types.CreateAdjustedBorrowKey(addr, "ibc/abcd"),
			[][]byte{
				{0x02},       // prefix
				{0x14},       // address length prefix = 20
				addrbytes,    // addr________________
				ibcabcdbytes, // ibc/abcd
				{0x00},       // null terminator
			},
			"adjusted borrow key (ibc)",
		},
		{
			types.CreateAdjustedBorrowKeyNoDenom(addr),
			[][]byte{
				{0x02},    // prefix
				{0x14},    // address length prefix = 20
				addrbytes, // addr________________
			},
			"adjusted borrow key (no denom)",
		},
		{
			types.CreateCollateralAmountKey(addr, "u/ibc/abcd"),
			[][]byte{
				{0x04},    // prefix
				{0x14},    // address length prefix = 20
				addrbytes, // addr________________
				uibcbytes, // u/ibc/abcd
				{0x00},    // null terminator
			},
			"collateral amount key",
		},
		{
			types.CreateCollateralAmountKeyNoDenom(addr),
			[][]byte{
				{0x04},    // prefix
				{0x14},    // address length prefix = 20
				addrbytes, // addr________________
			},
			"collateral amount key (no denom)",
		},
		{
			types.CreateReserveAmountKey("ibc/abcd"),
			[][]byte{
				{0x05},       // prefix
				ibcabcdbytes, // ibc/abcd
				{0x00},       // null terminator
			},
			"reserve amount key",
		},
		{
			types.CreateBadDebtKey("u/ibc/abcd", addr),
			[][]byte{
				{0x07},    // prefix
				{0x14},    // address length prefix = 20
				addrbytes, // addr________________
				uibcbytes, // u/ibc/abcd
				{0x00},    // null terminator
			},
			"bad debt key",
		},
	}
	for _, tc := range testCases {
		expectedKey := []byte{}
		for _, e := range tc.expected {
			expectedKey = append(expectedKey, e...)
		}
		require.Equalf(
			t,
			expectedKey,
			tc.actual,
			tc.description,
		)
	}
}
