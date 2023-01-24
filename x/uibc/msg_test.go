package uibc

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"gotest.tools/assert"
)

func TestMsgGovUpdateQuota(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		authority     string
		description   string
		total         sdk.Dec
		perDenom      sdk.Dec
		quotaDuration time.Duration
		errExpected   bool
	}{
		{
			name:          "valid msg",
			title:         "update quota",
			authority:     authtypes.NewModuleAddress("gov").String(),
			description:   "desc",
			total:         sdk.MustNewDecFromStr("1000"),
			perDenom:      sdk.MustNewDecFromStr("1000"),
			quotaDuration: 100,
			errExpected:   false,
		},
		{
			name:          "invalid authority address in msg",
			title:         "update quota",
			authority:     authtypes.NewModuleAddress("govv").String(),
			description:   "desc",
			total:         sdk.MustNewDecFromStr("1000"),
			perDenom:      sdk.MustNewDecFromStr("1000"),
			quotaDuration: 100,
			errExpected:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMsgGovUpdateQuota(tc.authority, tc.title, tc.description, tc.total, tc.perDenom, tc.quotaDuration)
			err := m.ValidateBasic()
			if tc.errExpected {
				assert.ErrorContains(t, err, "invalid authority")
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestMsgGovSetIBCPause(t *testing.T) {
	tests := []struct {
		name           string
		title          string
		authority      string
		description    string
		IbcPauseStatus IBCTransferStatus
		errExpected    bool
	}{
		{
			name:           "valid msg",
			title:          "title",
			authority:      authtypes.NewModuleAddress("gov").String(),
			description:    "desc",
			IbcPauseStatus: 1,
			errExpected:    false,
		},
		{
			name:           "invalid authority address in msg",
			title:          "title",
			authority:      authtypes.NewModuleAddress("govv").String(),
			description:    "desc",
			IbcPauseStatus: 1,
			errExpected:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMsgGovSetIBCPause(tc.authority, tc.title, tc.description, tc.IbcPauseStatus)
			err := m.ValidateBasic()
			if tc.errExpected {
				assert.ErrorContains(t, err, "invalid authority")
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
