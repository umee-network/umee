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
