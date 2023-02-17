package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestGetQuotas(t *testing.T) {
	s := initKeeperTestSuite(t)
	ctx, k := s.ctx, s.app.UIbcQuotaKeeper

	quotas, err := k.GetAllQuotas(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(quotas), 0)

	setQuotas := sdk.DecCoins{sdk.NewInt64DecCoin("test_uumee", 10000)}

	k.SetDenomQuotas(ctx, setQuotas)
	quotas, err = k.GetAllQuotas(ctx)
	assert.NilError(t, err)
	assert.DeepEqual(t, setQuotas, quotas)

	// get the quota of denom
	quota, err := k.GetQuota(ctx, setQuotas[0].Denom)
	assert.NilError(t, err)
	assert.Equal(t, quota.Denom, setQuotas[0].Denom)
}

func TestGetLocalDenom(t *testing.T) {
	s := initKeeperTestSuite(t)
	k := s.app.UIbcQuotaKeeper
	out := k.GetLocalDenom("umee")
	assert.Equal(t, "umee", out)
}

func TestResetQuota(t *testing.T) {
	s := initKeeperTestSuite(t)
	ctx, k := s.ctx, s.app.UIbcQuotaKeeper

	umeeQuota := sdk.NewInt64DecCoin("uumee", 1000)
	k.SetDenomQuota(ctx, umeeQuota)
	q, err := k.GetQuota(ctx, umeeQuota.Denom)
	assert.NilError(t, err)
	assert.DeepEqual(t, q, umeeQuota)

	// reset the quota
	k.ResetQuota(ctx)

	// check the quota after reset
	q, err = k.GetQuota(ctx, umeeQuota.Denom)
	assert.NilError(t, err)
	assert.DeepEqual(t, q.Amount, sdk.NewDec(0))
}
