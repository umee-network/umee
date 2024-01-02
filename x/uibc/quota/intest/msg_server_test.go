package intest

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/tests/accs"
	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/x/uibc"
)

func TestMsgServer_GovUpdateQuota(t *testing.T) {
	s := initTestSuite(t)

	tests := []struct {
		name   string
		msg    uibc.MsgGovUpdateQuota
		errMsg string
	}{
		{
			name: "unauthorized to increase the quota",
			msg: uibc.MsgGovUpdateQuota{
				Description:                 "some description",
				Authority:                   accs.Alice.String(),
				QuotaDuration:               time.Duration(time.Minute * 100),
				PerDenom:                    sdk.NewDec(1000),
				Total:                       sdk.NewDec(10000),
				InflowOutflowQuotaBase:      sdk.NewDec(200),
				InflowOutflowTokenQuotaBase: sdk.NewDec(200),
				InflowOutflowQuotaRate:      sdk.NewDecWithPrec(1, 1),
			},
			errMsg: "unauthorized",
		},
		{
			name: "invalid quota in msg",
			msg: uibc.MsgGovUpdateQuota{
				Description:                 "",
				Authority:                   checkers.GovModuleAddr,
				QuotaDuration:               time.Duration(time.Minute * 100),
				PerDenom:                    sdk.NewDec(1000),
				Total:                       sdk.NewDec(100),
				InflowOutflowQuotaBase:      sdk.NewDec(200),
				InflowOutflowTokenQuotaBase: sdk.NewDec(200),
				InflowOutflowQuotaRate:      sdk.NewDecWithPrec(1, 1),
			},
			errMsg: "total quota must be greater than or equal to per_denom quota",
		},
		{
			name: "valid msg",
			msg: uibc.MsgGovUpdateQuota{
				Description:                 "",
				Authority:                   checkers.GovModuleAddr,
				QuotaDuration:               time.Duration(time.Minute * 100),
				PerDenom:                    sdk.NewDec(1000),
				Total:                       sdk.NewDec(10000),
				InflowOutflowQuotaBase:      sdk.NewDec(200),
				InflowOutflowTokenQuotaBase: sdk.NewDec(200),
				InflowOutflowQuotaRate:      sdk.NewDecWithPrec(1, 1),
			},
			errMsg: "",
		},
		{
			name: "valid update the new params again",
			msg: uibc.MsgGovUpdateQuota{
				Description:                 "",
				Authority:                   checkers.GovModuleAddr,
				QuotaDuration:               time.Duration(time.Minute * 1000),
				PerDenom:                    sdk.NewDec(10000),
				Total:                       sdk.NewDec(100000),
				InflowOutflowQuotaBase:      sdk.NewDec(200),
				InflowOutflowTokenQuotaBase: sdk.NewDec(200),
				InflowOutflowQuotaRate:      sdk.NewDecWithPrec(1, 1),
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
		msg    uibc.MsgGovSetIBCStatus
		errMsg string
	}{
		{
			name: "invalid authority address in msg",
			msg: uibc.MsgGovSetIBCStatus{
				Description: "desc",
				Authority:   accs.Alice.String(),
				IbcStatus:   1,
			},
			errMsg: "unauthorized",
		}, {
			name: "invalid ibc-transfer status in msg",
			msg: uibc.MsgGovSetIBCStatus{
				Description: "",
				Authority:   checkers.GovModuleAddr,
				IbcStatus:   10,
			},
			errMsg: "invalid ibc-transfer status",
		}, {
			name: "valid ibc-transfer pause",
			msg: uibc.MsgGovSetIBCStatus{
				Description: "",
				Authority:   checkers.GovModuleAddr,
				IbcStatus:   2,
			},
			errMsg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.msgServer.GovSetIBCStatus(s.ctx, &tc.msg)
			if tc.errMsg == "" {
				assert.NilError(t, err)
				// check the update ibc-transfer pause status
				params, err := s.queryClient.Params(s.ctx, &uibc.QueryParams{})
				assert.NilError(t, err)
				assert.Equal(t, params.Params.IbcStatus, tc.msg.IbcStatus)
			} else {
				assert.ErrorContains(t, err, tc.errMsg)
			}
		})
	}
}
