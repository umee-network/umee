package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/umee-network/umee/v4/x/uibc"
	"github.com/umee-network/umee/v4/x/uibc/fixtures"
	"gotest.tools/v3/assert"
)

type MsgUpdate struct {
	denom   string
	outflow sdk.Dec
}

func TestCheckAndUpdateQuota(t *testing.T) {
	ctrl := gomock.NewController(t)
	o := fixtures.NewMockOracleKeeper(ctrl)
	ctx := sdk.Context{}
	denom := "umee"

	o.EXPECT().HistoricAvgPrice(gomock.Any(), denom).Return(sdk.MustNewDecFromStr("0.01"), nil)

	k := outflowKeeper{o, nil}
	prev, err := k.GetQuota(ctx, denom)
	assert.NilError(t, err)

	amount := sdkmath.NewIntFromUint64(10)
	k.checkAndUpdateQuota(sdk.Context{}, "u"+denom, amount, uibc.Params{})

	after, err := k.GetQuota(ctx, denom)
	assert.NilError(t, err)
	assert.Equal(t, after, prev.Add(amount))

}
