//go:build experimental
// +build experimental

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
	s := initKeeperTestSuite(t)
	ctx := s.ctx

	tests := []struct {
		name        string
		msg         uibc.MsgGovUpdateQuota
		errExpected bool
		errMsg      string
	}{
		{
			name: "invlaid authority address in msg",
			msg: uibc.MsgGovUpdateQuota{
				Title:       "title",
				Description: "desc",
				Authority:   authtypes.NewModuleAddress("govv").String(),
				Total:       sdk.NewDec(10),
				PerDenom:    sdk.NewDec(1),
			},
			errExpected: true,
			errMsg:      "expected gov account as only signer for proposal message",
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
			errMsg:      "total quota must be greater than or equal to per_denom quota",
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
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.msgServer.GovUpdateQuota(ctx, &tc.msg)
			if tc.errExpected {
				assert.ErrorContains(t, err, tc.errMsg)
			} else {
				assert.NilError(t, err)
				// check the update quota params
				paramsRes, err := s.queryClient.Params(ctx, &uibc.QueryParams{})
				assert.NilError(t, err)
				assert.Equal(t, paramsRes.Params.QuotaDuration, tc.msg.QuotaDuration)
				assert.Equal(t, true, paramsRes.Params.TokenQuota.Equal(tc.msg.PerDenom))
				assert.Equal(t, true, paramsRes.Params.TotalQuota.Equal(tc.msg.Total))
			}
		})
	}
}

func TestMsgServer_GovSetIBCPause(t *testing.T) {
	s := initKeeperTestSuite(t)
	ctx := s.ctx

	tests := []struct {
		name        string
		msg         uibc.MsgGovSetIBCPause
		errExpected bool
		errMsg      string
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
			errMsg:      "expected gov account as only signer for proposal message",
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
			errMsg:      "invalid ibc-transfer status",
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
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.msgServer.GovSetIBCPause(ctx, &tc.msg)
			if tc.errExpected {
				assert.ErrorContains(t, err, tc.errMsg)
			} else {
				assert.NilError(t, err)
				// check the update ibc-transfer pause status
				params, err := s.queryClient.Params(ctx, &uibc.QueryParams{})
				assert.NilError(t, err)
				assert.Equal(t, params.Params.IbcPause, tc.msg.IbcPauseStatus)
			}
		})
	}
}
