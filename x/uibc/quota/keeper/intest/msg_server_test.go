package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v4/x/uibc"
)

func TestMsgServer_GovUpdateQuota(t *testing.T) {
	s := initTestSuite(t)

	tests := []struct {
		name   string
		msg    uibc.MsgGovUpdateQuota
		errMsg string
	}{
		{
			name: "invalid authority address in msg",
			msg: uibc.MsgGovUpdateQuota{
				Title:       "title",
				Description: "desc",
				Authority:   authtypes.NewModuleAddress("govv").String(),
				Total:       sdk.NewDec(10),
				PerDenom:    sdk.NewDec(1),
			},
			errMsg: "expected gov account as only signer for proposal message",
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
			errMsg: "total quota must be greater than or equal to per_denom quota",
		},
		{
			name: "valid msg",
			msg: uibc.MsgGovUpdateQuota{
				Title:         "title",
				Description:   "desc",
				Authority:     authtypes.NewModuleAddress("gov").String(),
				QuotaDuration: time.Duration(time.Minute * 100),
				PerDenom:      sdk.NewDec(1000),
				Total:         sdk.NewDec(10000),
			},
			errMsg: "",
		},
		{
			name: "valid msg with new update <update the new params again>",
			msg: uibc.MsgGovUpdateQuota{
				Title:         "override new params",
				Description:   "desc",
				Authority:     authtypes.NewModuleAddress("gov").String(),
				QuotaDuration: time.Duration(time.Minute * 1000),
				PerDenom:      sdk.NewDec(10000),
				Total:         sdk.NewDec(100000),
			},
			errMsg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.msgServer.GovUpdateQuota(s.ctx, &tc.msg)
			if tc.errMsg == "" {
				assert.NilError(t, err)
				// check the update quota params
				paramsRes, err := s.queryClient.Params(s.ctx, &uibc.QueryParams{})
				assert.NilError(t, err)
				assert.Equal(t, paramsRes.Params.QuotaDuration, tc.msg.QuotaDuration)
				assert.DeepEqual(t, paramsRes.Params.TokenQuota, tc.msg.PerDenom)
				assert.DeepEqual(t, paramsRes.Params.TotalQuota, tc.msg.Total)
			} else {
				assert.ErrorContains(t, err, tc.errMsg)
			}
		})
	}
}

func TestMsgServer_GovSetIBCStatus(t *testing.T) {
	s := initTestSuite(t)

	tests := []struct {
		name   string
		msg    uibc.MsgGovSetIBCSQuotaStatus
		errMsg string
	}{
		{
			name: "invalid authority address in msg",
			msg: uibc.MsgGovSetIBCSQuotaStatus{
				Title:       "title",
				Description: "desc",
				Authority:   authtypes.NewModuleAddress("govv").String(),
				QuotaStatus: 1,
			},
			errMsg: "expected gov account as only signer for proposal message",
		}, {
			name: "invalid ibc-transfer status in msg",
			msg: uibc.MsgGovSetIBCSQuotaStatus{
				Title:       "title",
				Description: "desc",
				Authority:   authtypes.NewModuleAddress("gov").String(),
				QuotaStatus: 5,
			},
			errMsg: "invalid ibc-transfer status",
		}, {
			name: "valid in msg <enable the ibc-transfer pause",
			msg: uibc.MsgGovSetIBCSQuotaStatus{
				Title:       "title",
				Description: "desc",
				Authority:   authtypes.NewModuleAddress("gov").String(),
				QuotaStatus: 2,
			},
			errMsg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.msgServer.GovSetIBCQuotaStatus(s.ctx, &tc.msg)
			if tc.errMsg == "" {
				assert.NilError(t, err)
				// check the update ibc-transfer pause status
				params, err := s.queryClient.Params(s.ctx, &uibc.QueryParams{})
				assert.NilError(t, err)
				assert.Equal(t, params.Params.QuotaStatus, tc.msg.QuotaStatus)
			} else {
				assert.ErrorContains(t, err, tc.errMsg)
			}
		})
	}
}
