package keeper

import (
	"testing"

	"cosmossdk.io/errors"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v4/util/decmath"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"
)

func TestErrorMatching(t *testing.T) {
	// oracle errors
	err1 := errors.Wrap(decmath.ErrEmptyList, "denom: UMEE")
	err2 := oracletypes.ErrUnknownDenom.Wrap("UMEE")
	err3 := leveragetypes.ErrNoHistoricMedians.Wrapf(
		"requested %d, got %d",
		16,
		12,
	)
	// not oracle errors
	err4 := leveragetypes.ErrBlacklisted
	err5 := leveragetypes.ErrUToken
	err6 := leveragetypes.ErrNotRegisteredToken
	err7 := errors.New("foo", 1, "bar")

	assert.Equal(t, false, nonOracleError(nil))
	assert.Equal(t, false, nonOracleError(err1))
	assert.Equal(t, false, nonOracleError(err2))
	assert.Equal(t, false, nonOracleError(err3))
	assert.Equal(t, true, nonOracleError(err4))
	assert.Equal(t, true, nonOracleError(err5))
	assert.Equal(t, true, nonOracleError(err6))
	assert.Equal(t, true, nonOracleError(err7))
}
