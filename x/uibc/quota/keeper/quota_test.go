//go:build experimental
// +build experimental

package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/uibc"
)

func (s *KeeperTestSuite) TestGetQuotas() {
	ctx, k := s.ctx, s.app.UIbcQuotaKeeper

	quotas, err := k.GetQuotaOfIBCDenoms(ctx)
	s.Require().NoError(err)
	s.Require().Equal(len(quotas), 0)

	setQuotas := []uibc.Quota{
		{
			IbcDenom:   "test_uumee",
			OutflowSum: sdk.MustNewDecFromStr("10000"),
		},
	}

	err = k.SetDenomQuotas(ctx, setQuotas)
	s.Require().NoError(err)
	quotas, err = k.GetQuotaOfIBCDenoms(ctx)
	s.Require().NoError(err)
	s.Require().Equal(len(quotas), len(setQuotas))

	// get the quota of denom
	quota, err := k.GetQuotaByDenom(ctx, setQuotas[0].IbcDenom)
	s.Require().NoError(err)
	s.Require().Equal(quota.IbcDenom, setQuotas[0].IbcDenom)
}

func (s *KeeperTestSuite) TestGetLocalDenom() {
	k := s.app.UIbcQuotaKeeper
	out := k.GetLocalDenom("umee")
	s.Require().Equal("umee", out)
}
