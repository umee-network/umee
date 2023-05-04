package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v4/tests/tsdk"
	"github.com/umee-network/umee/v4/x/uibc"
)

const (
	umee = "UUMEE"
	atom = "ATOM"
)

// creates keeper without external dependencies (app, leverage etc...)
func initKeeper(t *testing.T, l uibc.Leverage, o uibc.Oracle) TestKeeper {
	cdc := codec.NewProtoCodec(nil)
	storeKey := storetypes.NewMemoryStoreKey("quota")
	k := NewKeeper(cdc, storeKey, nil, l, o)
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)
	return TestKeeper{k, t, &ctx}
}

// creates keeper without simple mock of leverage and oracle, providing token settings and
// prices for umee and atom
func initUmeeKeeper(t *testing.T) TestKeeper {
	lmock := NewLeverageKeeperMock(umee, atom)
	omock := NewOracleMock(umee, sdk.NewDec(2))
	omock.prices[atom] = sdk.NewDec(10)
	return initKeeper(t, lmock, omock)
}

type TestKeeper struct {
	Keeper
	t   *testing.T
	ctx *sdk.Context
}

func (k TestKeeper) checkOutflows(denom string, perToken, total int64) {
	o, err := k.GetTokenOutflows(*k.ctx, denom)
	require.NoError(k.t, err)
	require.Equal(k.t, sdk.NewDec(perToken), o.Amount)

	d := k.GetTotalOutflow(*k.ctx)
	require.Equal(k.t, sdk.NewDec(total), d)
}

func (k TestKeeper) setQuotaParams(perToken, total int64) {
	err := k.SetParams(*k.ctx,
		uibc.Params{TokenQuota: sdk.NewDec(perToken), TotalQuota: sdk.NewDec(total)})
	require.NoError(k.t, err)
}
