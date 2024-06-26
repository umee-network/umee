package uibc

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/umee-network/umee/v6/tests/accs"
	"github.com/umee-network/umee/v6/tests/tcheckers"
	"github.com/umee-network/umee/v6/util/checkers"
)

func TestMsgGovUpdateQuota(t *testing.T) {
	t.Parallel()
	validMsg := MsgGovUpdateQuota{
		Authority:                   checkers.GovModuleAddr,
		Description:                 "",
		Total:                       sdkmath.LegacyMustNewDecFromStr("1000"),
		PerDenom:                    sdkmath.LegacyMustNewDecFromStr("1000"),
		InflowOutflowQuotaBase:      sdkmath.LegacyMustNewDecFromStr("500"),
		InflowOutflowTokenQuotaBase: sdkmath.LegacyMustNewDecFromStr("500"),
		InflowOutflowQuotaRate:      sdkmath.LegacyMustNewDecFromStr("5"),
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
	invalidTotalQuota.PerDenom = sdkmath.LegacyNewDec(10)
	invalidTotalQuota.Total = sdkmath.LegacyNewDec(2)

	invalidInflowOutflow := validMsg
	invalidInflowOutflow.InflowOutflowTokenQuotaBase = sdkmath.LegacyMustNewDecFromStr("501")

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
		tcheckers.ErrorContains(t, tc.msg.ValidateBasic(), tc.errMsg, tc.name)
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
		tcheckers.ErrorContains(t, tc.msg.ValidateBasic(), tc.errMsg, tc.name)
	}
}

func TestMsgGovTMsgGovToggleICS20Hooks(t *testing.T) {
	t.Parallel()
	validMsg := MsgGovToggleICS20Hooks{
		Authority:   checkers.GovModuleAddr,
		Description: "",
		Enabled:     true,
	}

	validMsg2 := validMsg
	validMsg2.Enabled = false

	validEmergencyGroup := validMsg
	validEmergencyGroup.Authority = accs.Alice.String()
	validEmergencyGroup.Description = "not empty"

	invalidAuthority := validMsg
	invalidAuthority.Authority = "ABC"

	tests := []struct {
		name   string
		msg    MsgGovToggleICS20Hooks
		errMsg string
	}{
		{
			name:   "valid msg",
			msg:    validMsg,
			errMsg: "",
		}, {
			name:   "valid msg2",
			msg:    validMsg2,
			errMsg: "",
		}, {
			name:   "valid msg with other authority",
			msg:    validEmergencyGroup,
			errMsg: "",
		}, {
			name:   "invalid authority address in msg",
			msg:    invalidAuthority,
			errMsg: "invalid bech32",
		},
	}

	for _, tc := range tests {
		tcheckers.ErrorContains(t, tc.msg.ValidateBasic(), tc.errMsg, tc.name)
	}

}
