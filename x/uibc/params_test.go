package uibc

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"gotest.tools/v3/assert"
)

func TestValidateIBCTransferStatus(t *testing.T) {
	err := validateIBCTransferStatus(10)
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
	err := validateQuota(sdkmath.LegacyNewDec(-1), "s")
	assert.ErrorContains(t, err, "must be not negative")
	err = validateQuota(sdkmath.LegacyNewDec(100), "s")
	assert.NilError(t, err)
}
