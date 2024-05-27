package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v6/tests/accs"
	"github.com/umee-network/umee/v6/tests/tsdk"
	"github.com/umee-network/umee/v6/x/metoken"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

// initSimpleKeeper creates a simple keeper without external dependencies.
func initSimpleKeeper(t *testing.T) Keeper {
	t.Parallel()
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	storeKey := storetypes.NewMemoryStoreKey("metoken")
	kb := NewBuilder(cdc, storeKey, nil, nil, nil, nil, accs.GenerateAddr("auction.Rewards"))
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)

	k := kb.Keeper(&ctx)
	err := k.SetParams(metoken.DefaultParams())
	require.NoError(t, err)
	return k
}

// initMeUSDKeeper creates a keeper with external dependencies and with meUSD index and balance inserted.
func initMeUSDKeeper(
	t *testing.T,
	bankKeeper metoken.BankKeeper,
	leverageKeeper metoken.LeverageKeeper,
	oracleKeeper metoken.OracleKeeper,
) Keeper {
	k := initSimpleKeeper(t)
	k.bankKeeper = bankKeeper
	k.leverageKeeper = leverageKeeper
	k.oracleKeeper = oracleKeeper

	index := mocks.StableIndex(mocks.MeUSDDenom)
	err := k.setRegisteredIndex(index)
	require.NoError(t, err)

	balance := mocks.ValidUSDIndexBalances(mocks.MeUSDDenom)
	err = k.setIndexBalances(balance)

	return k
}

func initMeUSDNoopKeper(t *testing.T) Keeper {
	return initMeUSDKeeper(t, NewBankMock(), NewLeverageMock(), NewOracleMock())
}
