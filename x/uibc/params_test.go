package uibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestValidateIBCTransferStatus(t *testing.T) {
	err := validateIBCTransferStatus(4)
	require.ErrorContains(t, err, "invalid ibc-transfer status")
	err = validateIBCTransferStatus(2)
	require.NoError(t, err)
}

func TestValidateQuotaDuration(t *testing.T) {
	err := validateQuotaDuration(-1)
	require.ErrorContains(t, err, "quota duration time must be positive")
	err = validateQuotaDuration(10)
	require.NoError(t, err)
}

func TestValidateQuota(t *testing.T) {
	err := validateQuota(sdk.NewDec(-1), "s")
	require.ErrorContains(t, err, "must be not negative")
	err = validateQuota(sdk.NewDec(100), "s")
	require.NoError(t, err)
}
