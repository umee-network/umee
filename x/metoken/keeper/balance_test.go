package keeper

import (
	"testing"

	"github.com/umee-network/umee/v6/x/metoken/mocks"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestUnitBalance(t *testing.T) {
	k := initSimpleKeeper(t)

	_, err := k.IndexBalances("inexistingMetoken")
	require.ErrorIs(t, err, sdkerrors.ErrNotFound)

	balance := mocks.ValidUSDIndexBalances(mocks.MeUSDDenom)
	err = k.setIndexBalances(balance)

	balance2, err := k.IndexBalances(balance.MetokenSupply.Denom)
	require.NoError(t, err)
	require.Equal(t, balance, balance2)
}
