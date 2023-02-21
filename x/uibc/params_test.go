package uibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestValidateIBCTransferStatus(t *testing.T) {
	err := validateIBCTransferStatus(4)
	assert.ErrorContains(t, err, "invalid ibc-transfer status")
	err = validateIBCTransferStatus(2)
	assert.NilError(t, err)
}

func TestValidateQuotaDuration(t *testing.T) {
	err := validateQuotaDuration(-1)
	assert.ErrorContains(t, err, "quota duration time must be positive")
	err = validateQuotaDuration(10)
	assert.NilError(t, err)
}

func TestValidateQuota(t *testing.T) {
	err := validateQuota(sdk.NewDec(-1), "s")
	assert.ErrorContains(t, err, "must be not negative")
	err = validateQuota(sdk.NewDec(100), "s")
	assert.NilError(t, err)
}
