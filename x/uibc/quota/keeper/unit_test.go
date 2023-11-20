package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v6/tests/tsdk"
	ugovmocks "github.com/umee-network/umee/v6/x/ugov/mocks"
	"github.com/umee-network/umee/v6/x/uibc"
)

const (
	umee = "UUMEE"
	atom = "ATOM"
)

// creates keeper without external dependencies (app, leverage etc...)
func initKeeper(t *testing.T, l uibc.Leverage, o uibc.Oracle) TestKeeper {
	eg := ugovmocks.NewSimpleEmergencyGroupBuilder()
	ir := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(ir)
	storeKey := storetypes.NewMemoryStoreKey("quota")
	kb := NewKeeperBuilder(cdc, storeKey, l, o, eg)
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)
	return TestKeeper{kb.Keeper(&ctx), t, &ctx}
}

// creates keeper without simple mock of leverage and oracle, providing token settings and
// prices for umee and atom
func initKeeperSimpleMock(t *testing.T) TestKeeper {
	lmock := NewLeverageKeeperMock(umee, atom)
	omock := NewOracleMock(umee, sdk.NewDec(2))
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

	d := k.GetOutflowSum()
	require.Equal(k.t, sdk.NewDec(total), d)
}

func (k TestKeeper) setQuotaParams(perToken, total int64) {
	dp := uibc.DefaultParams()
	dp.TokenQuota = sdk.NewDec(perToken)
	dp.TotalQuota = sdk.NewDec(total)
	err := k.SetParams(dp)
	require.NoError(k.t, err)
}
