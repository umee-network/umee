package uibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"gotest.tools/v3/assert"
)

func validMsgGovUpdateQuota() MsgGovUpdateQuota {
	return MsgGovUpdateQuota{
		Title:         "update quota",
		Authority:     authtypes.NewModuleAddress("gov").String(),
		Description:   "desc",
		Total:         sdk.MustNewDecFromStr("1000"),
		PerDenom:      sdk.MustNewDecFromStr("1000"),
		QuotaDuration: 100,
	}
}

func TestMsgGovUpdateQuota(t *testing.T) {
	t.Parallel()
	validMsg := validMsgGovUpdateQuota()

	invalidAuthority := validMsg
	invalidAuthority.Authority = authtypes.NewModuleAddress("govv").String()

	invalidTotalQuota := validMsg
	invalidTotalQuota.PerDenom = sdk.NewDec(10)
	invalidTotalQuota.Total = sdk.NewDec(2)

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
			name:   "invalid authority address in msg",
			msg:    invalidAuthority,
			errMsg: "expected gov account",
		}, {
			name:   "invalid total quota with respect to per denom",
			msg:    invalidTotalQuota,
			errMsg: "total quota must be greater than or equal to per_denom quota",
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

func validMsgGovSetIBCStatus() MsgGovSetIBCStatus {
	return MsgGovSetIBCStatus{
		Title:       "title",
		Authority:   authtypes.NewModuleAddress("gov").String(),
		Description: "desc",
		IbcStatus:   1,
	}
}

func TestMsgGovSetIBCStatus(t *testing.T) {
	t.Parallel()
	validMsg := validMsgGovSetIBCStatus()

	invalidAuthority := validMsg
	invalidAuthority.Authority = authtypes.NewModuleAddress("govv").String()

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
			name:   "invalid authority address in msg",
			msg:    invalidAuthority,
			errMsg: "expected gov account",
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
