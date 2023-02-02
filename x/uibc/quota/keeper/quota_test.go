//go:build experimental
// +build experimental

package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/uibc"
	"gotest.tools/v3/assert"
)

func TestGetQuotas(t *testing.T) {
	s := initIntegrationSuite(t)
	ctx, k := s.ctx, s.app.UIbcQuotaKeeper

	quotas, err := k.GetQuotaOfIBCDenoms(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(quotas), 0)

	setQuotas := []uibc.Quota{
		{
			IbcDenom:   "test_uumee",
			OutflowSum: sdk.MustNewDecFromStr("10000"),
		},
	}

	k.SetDenomQuotas(ctx, setQuotas)
	quotas, err = k.GetQuotaOfIBCDenoms(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(quotas), len(setQuotas))

	// get the quota of denom
	quota, err := k.GetQuotaByDenom(ctx, setQuotas[0].IbcDenom)
	assert.NilError(t, err)
	assert.Equal(t, quota.IbcDenom, setQuotas[0].IbcDenom)
}

func TestGetLocalDenom(t *testing.T) {
	s := initIntegrationSuite(t)
	k := s.app.UIbcQuotaKeeper
	out := k.GetLocalDenom("umee")
	assert.Equal(t, "umee", out)
}

func TestResetQuota(t *testing.T) {
	s := initIntegrationSuite(t)
	ctx, k := s.ctx, s.app.UIbcQuotaKeeper

	umeeQuota := uibc.Quota{IbcDenom: "uumee", OutflowSum: sdk.NewDec(1000)}
	k.SetDenomQuota(ctx, umeeQuota)
	// get the qutoa
	q, err := k.GetQuotaByDenom(ctx, umeeQuota.IbcDenom)
	assert.NilError(t, err)
	assert.Equal(t, q.GetIbcDenom(), umeeQuota.IbcDenom)
	assert.DeepEqual(t, q.OutflowSum, umeeQuota.OutflowSum)

	// reset the quota
	k.ResetQuota(ctx)

	// check the quota after reset
	q, err = k.GetQuotaByDenom(ctx, umeeQuota.IbcDenom)
	assert.NilError(t, err)
	assert.DeepEqual(t, q.OutflowSum, sdk.NewDec(0))
}
