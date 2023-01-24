//go:build experimental
// +build experimental

package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/umee-network/umee/v4/x/uibc"
)

func (s *IntegrationTestSuite) TestMsgServer_GovUpdateQuota() {
	ctx := s.ctx

	tests := []struct {
		name        string
		msg         uibc.MsgGovUpdateQuota
		errExpected bool
	}{
		{
			name: "invlaid authority address in msg",
			msg: uibc.MsgGovUpdateQuota{
				Title:       "title",
				Description: "desc",
				Authority:   authtypes.NewModuleAddress("govv").String(),
			},
			errExpected: true,
		},
		{
			name: "invalid quota in msg",
			msg: uibc.MsgGovUpdateQuota{
				Title:         "title",
				Description:   "desc",
				Authority:     authtypes.NewModuleAddress("gov").String(),
				QuotaDuration: time.Duration(time.Minute * 100),
				PerDenom:      sdk.NewDec(1000),
				Total:         sdk.NewDec(100),
			},
			errExpected: true,
		},
		{
			name: "valid in msg",
			msg: uibc.MsgGovUpdateQuota{
				Title:         "title",
				Description:   "desc",
				Authority:     authtypes.NewModuleAddress("gov").String(),
				QuotaDuration: time.Duration(time.Minute * 100),
				PerDenom:      sdk.NewDec(1000),
				Total:         sdk.NewDec(10000),
			},
			errExpected: false,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := s.msgServer.GovUpdateQuota(ctx, &tc.msg)
			if tc.errExpected {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				// check the update quota params
				paramsRes, err := s.queryClient.Params(ctx, &uibc.QueryParams{})
				s.Require().NoError(err)
				s.Require().Equal(paramsRes.Params.QuotaDuration, tc.msg.QuotaDuration)
				s.Require().Equal(paramsRes.Params.TokenQuota, tc.msg.PerDenom)
				s.Require().Equal(paramsRes.Params.TotalQuota, tc.msg.Total)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestMsgServer_GovSetIBCPause() {
	ctx := s.ctx

	tests := []struct {
		name        string
		msg         uibc.MsgGovSetIBCPause
		errExpected bool
	}{
		{
			name: "invlaid authority address in msg",
			msg: uibc.MsgGovSetIBCPause{
				Title:          "title",
				Description:    "desc",
				Authority:      authtypes.NewModuleAddress("govv").String(),
				IbcPauseStatus: 1,
			},
			errExpected: true,
		},
		{
			name: "invalid ibc-transfer status in msg",
			msg: uibc.MsgGovSetIBCPause{
				Title:          "title",
				Description:    "desc",
				Authority:      authtypes.NewModuleAddress("gov").String(),
				IbcPauseStatus: 5,
			},
			errExpected: true,
		},
		{
			name: "invalid ibc-transfer status in msg <default: disabled (1)>",
			msg: uibc.MsgGovSetIBCPause{
				Title:          "title",
				Description:    "desc",
				Authority:      authtypes.NewModuleAddress("gov").String(),
				IbcPauseStatus: 1,
			},
			errExpected: true,
		},
		{
			name: "valid in msg <enable the ibc-transfer pause",
			msg: uibc.MsgGovSetIBCPause{
				Title:          "title",
				Description:    "desc",
				Authority:      authtypes.NewModuleAddress("gov").String(),
				IbcPauseStatus: 2,
			},
			errExpected: false,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := s.msgServer.GovSetIBCPause(ctx, &tc.msg)
			if tc.errExpected {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				// check the update ibc-transfer pause status
				params, err := s.queryClient.Params(ctx, &uibc.QueryParams{})
				s.Require().NoError(err)
				s.Require().Equal(params.Params.IbcPause, tc.msg.IbcPauseStatus)
			}
		})
	}
}
