package uibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"gotest.tools/v3/assert"
)

func TestMsgGovUpdateQuota(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgGovUpdateQuota
		errMsg string
	}{
		{
			name: "valid msg",
			msg: MsgGovUpdateQuota{
				Title:         "update quota",
				Authority:     authtypes.NewModuleAddress("gov").String(),
				Description:   "desc",
				Total:         sdk.MustNewDecFromStr("1000"),
				PerDenom:      sdk.MustNewDecFromStr("1000"),
				QuotaDuration: 100,
			},
			errMsg: "",
		}, {
			name: "invalid authority address in msg",
			msg: MsgGovUpdateQuota{
				Title:         "update quota",
				Authority:     authtypes.NewModuleAddress("govv").String(),
				Description:   "desc",
				Total:         sdk.MustNewDecFromStr("1000"),
				PerDenom:      sdk.MustNewDecFromStr("1000"),
				QuotaDuration: 100,
			},
			errMsg: "invalid authority",
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

func TestMsgGovSetIBCPause(t *testing.T) {
	tests := []struct {
		msg    MsgGovSetIBCPause
		name   string
		errMsg string
	}{
		{
			msg: MsgGovSetIBCPause{
				Title:          "title",
				Authority:      authtypes.NewModuleAddress("gov").String(),
				Description:    "desc",
				IbcPauseStatus: 1,
			},
			name:   "valid msg",
			errMsg: "",
		}, {
			name: "invalid authority address in msg",
			msg: MsgGovSetIBCPause{
				Title:          "title",
				Authority:      authtypes.NewModuleAddress("govv").String(),
				Description:    "desc",
				IbcPauseStatus: 1,
			},
			errMsg: "invalid authority",
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
