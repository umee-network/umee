package uibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/tests/accs"
	"github.com/umee-network/umee/v6/util/checkers"
	"gotest.tools/v3/assert"
)

func TestMsgGovUpdateQuota(t *testing.T) {
	t.Parallel()
	validMsg := MsgGovUpdateQuota{
		Authority:                   checkers.GovModuleAddr,
		Description:                 "",
		Total:                       sdk.MustNewDecFromStr("1000"),
		PerDenom:                    sdk.MustNewDecFromStr("1000"),
		InflowOutflowQuotaBase:      sdk.MustNewDecFromStr("500"),
		InflowOutflowTokenQuotaBase: sdk.MustNewDecFromStr("500"),
		InflowOutflowQuotaRate:      sdk.MustNewDecFromStr("5"),
		QuotaDuration:               100,
	}

	validEmergencyGroup := validMsg
	validEmergencyGroup.Authority = accs.Alice.String()
	validEmergencyGroup.Description = "not empty"

	invalidDesc := validMsg
	invalidDesc.Description = "not empty description"

	invalidDesc2 := validEmergencyGroup
	invalidDesc2.Description = ""

	invalidTotalQuota := validMsg
	invalidTotalQuota.PerDenom = sdk.NewDec(10)
	invalidTotalQuota.Total = sdk.NewDec(2)

	invalidInflowOutflow := validMsg
	invalidInflowOutflow.InflowOutflowTokenQuotaBase = sdk.MustNewDecFromStr("501")

	tests := []struct {
		name   string
		msg    MsgGovUpdateQuota
		errMsg string
	}{
		{
			name:   "valid msg",
			msg:    validMsg,
			errMsg: "",
		}, {
			name:   "valid msg other authority",
			msg:    validEmergencyGroup,
			errMsg: "",
		}, {
			name:   "invalid description",
			msg:    invalidDesc,
			errMsg: "description must be empty",
		}, {
			name:   "invalid description2",
			msg:    invalidDesc2,
			errMsg: "description must be not empty",
		}, {
			name:   "invalid total quota with respect to per denom",
			msg:    invalidTotalQuota,
			errMsg: "total quota must be greater than or equal to per_denom quota",
		}, {
			name:   "invalid inflow outflow quota base with respect to per denom",
			msg:    invalidInflowOutflow,
			errMsg: "inflow_outflow_quota_base must be greater than",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.errMsg == "" {
				assert.NilError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errMsg)
			}
		})
	}
}

func TestMsgGovSetIBCStatus(t *testing.T) {
	t.Parallel()
	validMsg := MsgGovSetIBCStatus{
		Authority:   checkers.GovModuleAddr,
		Description: "",
		IbcStatus:   1,
	}

	validEmergencyGroup := validMsg
	validEmergencyGroup.Authority = accs.Alice.String()
	validEmergencyGroup.Description = "not empty"

	invalidAuthority := validMsg
	invalidAuthority.Authority = "ABC"

	invalidIBCStatus := validMsg
	invalidIBCStatus.IbcStatus = 10

	tests := []struct {
		msg    MsgGovSetIBCStatus
		name   string
		errMsg string
	}{
		{
			msg:    validMsg,
			name:   "valid msg",
			errMsg: "",
		}, {
			msg:    validEmergencyGroup,
			name:   "valid msg with other authority",
			errMsg: "",
		}, {
			name:   "invalid authority address in msg",
			msg:    invalidAuthority,
			errMsg: "invalid bech32",
		}, {
			name:   "invalid ibc pause status in msg",
			msg:    invalidIBCStatus,
			errMsg: "invalid ibc-transfer status",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.errMsg == "" {
				assert.NilError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errMsg)
			}
		})
	}
}
