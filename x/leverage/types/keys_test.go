package types_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/x/leverage/types"
)

func TestAddressFromKey(t *testing.T) {
	address := sdk.AccAddress([]byte("addr________________"))
	key := types.KeyAdjustedBorrow(address, appparams.BondDenom)
	expectedAddress := types.AddressFromKey(key, types.KeyPrefixAdjustedBorrow)

	assert.DeepEqual(t, address, expectedAddress)

	address = sdk.AccAddress([]byte("anotherAddr________________"))
	key = types.KeyCollateralAmountNoDenom(address)
	expectedAddress = types.AddressFromKey(key, types.KeyPrefixAdjustedBorrow)

	assert.DeepEqual(t, address, expectedAddress)
}

func TestDenomFromKeyWithAddress(t *testing.T) {
	address := sdk.AccAddress([]byte("addr________________"))
	denom := appparams.BondDenom
	key := types.KeyAdjustedBorrow(address, denom)
	expectedDenom := types.DenomFromKeyWithAddress(key, types.KeyPrefixAdjustedBorrow)

	assert.Equal(t, denom, expectedDenom)

	uDenom := fmt.Sprintf("u%s", denom)
	key = types.KeyCollateralAmount(address, uDenom)
	expectedDenom = types.DenomFromKeyWithAddress(key, types.KeyPrefixCollateralAmount)

	assert.Equal(t, uDenom, expectedDenom)
}

func TestDenomFromKey(t *testing.T) {
	denom := appparams.BondDenom
	key := types.KeyReserveAmount(denom)
	expectedDenom := types.DenomFromKey(key, types.KeyPrefixReserveAmount)

	assert.Equal(t, denom, expectedDenom)

	uDenom := fmt.Sprintf("u%s", denom)
	key = types.KeyReserveAmount(uDenom)
	expectedDenom = types.DenomFromKey(key, types.KeyPrefixReserveAmount)

	assert.Equal(t, uDenom, expectedDenom)
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
			types.KeyRegisteredToken("uumee"),
			[][]byte{
				{0x01},     // prefix
				uumeebytes, // uumee
				{0x00},     // null terminator
			},
			"registered token key (uumee)",
		},
		{
			types.KeyRegisteredToken("ibc/abcd"),
			[][]byte{
				{0x01},       // prefix
				ibcabcdbytes, // ibc/abcd
				{0x00},       // null terminator
			},
			"registered token key (ibc/abcd)",
		},
		{
			types.KeyAdjustedBorrow(addr, "uumee"),
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
			types.KeyAdjustedBorrow(addr, "ibc/abcd"),
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
			types.KeyAdjustedBorrowNoDenom(addr),
			[][]byte{
				{0x02},    // prefix
				{0x14},    // address length prefix = 20
				addrbytes, // addr________________
			},
			"adjusted borrow key (no denom)",
		},
		{
			types.KeyCollateralAmount(addr, "u/ibc/abcd"),
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
			types.KeyCollateralAmountNoDenom(addr),
			[][]byte{
				{0x04},    // prefix
				{0x14},    // address length prefix = 20
				addrbytes, // addr________________
			},
			"collateral amount key (no denom)",
		},
		{
			types.KeyReserveAmount("ibc/abcd"),
			[][]byte{
				{0x05},       // prefix
				ibcabcdbytes, // ibc/abcd
				{0x00},       // null terminator
			},
			"reserve amount key",
		},
		{
			types.KeyBadDebt("u/ibc/abcd", addr),
			[][]byte{
				{0x07},    // prefix
				{0x14},    // address length prefix = 20
				addrbytes, // addr________________
				uibcbytes, // u/ibc/abcd
				{0x00},    // null terminator
			},
			"bad debt key",
		},
		{
			types.KeyInterestScalar("ibc/abcd"),
			[][]byte{
				{0x08},       // prefix
				ibcabcdbytes, // ibc/abcd
				{0x00},       // null terminator
			},
			"interest scalar key",
		},
		{
			types.KeyAdjustedTotalBorrow("ibc/abcd"),
			[][]byte{
				{0x09},       // prefix
				ibcabcdbytes, // ibc/abcd
				{0x00},       // null terminator
			},
			"adjusted total borrow key",
		},
		{
			types.KeyUTokenSupply("u/ibc/abcd"),
			[][]byte{
				{0x0A},    // prefix
				uibcbytes, // u/ibc/abcd
				{0x00},    // null terminator
			},
			"uToken supply key",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			expectedKey := []byte{}
			for _, e := range tc.expected {
				expectedKey = append(expectedKey, e...)
			}
			assert.DeepEqual(
				t,
				expectedKey,
				tc.actual,
			)
		})
	}
}
