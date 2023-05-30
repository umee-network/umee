package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v5/tests/tsdk"
	"github.com/umee-network/umee/v5/x/uibc"
)

const (
	umee = "UUMEE"
	atom = "ATOM"
)

// creates keeper without external dependencies (app, leverage etc...)
func initKeeper(t *testing.T, l uibc.Leverage, o uibc.Oracle) TestKeeper {
	ir := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(ir)
	storeKey := storetypes.NewMemoryStoreKey("quota")
	kb := NewKeeperBuilder(cdc, storeKey, nil, l, o)
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)
	return TestKeeper{kb.Keeper(&ctx), t, &ctx}
}

// creates keeper without simple mock of leverage and oracle, providing token settings and
// prices for umee and atom
func initKeeperSimpleMock(t *testing.T) TestKeeper {
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
	o := k.GetTokenOutflows(denom)
	require.Equal(k.t, sdk.NewDec(perToken), o.Amount)

	d := k.GetTotalOutflow()
	require.Equal(k.t, sdk.NewDec(total), d)
}

func (k TestKeeper) setQuotaParams(perToken, total int64) {
	err := k.SetParams(uibc.Params{TokenQuota: sdk.NewDec(perToken), TotalQuota: sdk.NewDec(total)})
	require.NoError(k.t, err)
}
