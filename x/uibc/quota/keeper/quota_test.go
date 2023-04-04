package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestResetQuota(t *testing.T) {
	s := initKeeperTestSuite(t)
	ctx, k := s.ctx, s.app.UIbcQuotaKeeper

	umeeQuota := sdk.NewInt64DecCoin("uumee", 1000)
	k.SetTokenOutflow(ctx, umeeQuota)
	q, err := k.GetOutflows(ctx, umeeQuota.Denom)
	assert.NilError(t, err)
	assert.DeepEqual(t, q, umeeQuota)

	// reset the quota
	k.ResetAllQuotas(ctx)

	// check the quota after reset
	q, err = k.GetOutflows(ctx, umeeQuota.Denom)
	assert.NilError(t, err)
	assert.DeepEqual(t, q.Amount, sdk.NewDec(0))
}
