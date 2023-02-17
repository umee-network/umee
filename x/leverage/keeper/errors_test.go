package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/errors"

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

	require.Equal(t, false, nonOracleError(nil))
	require.Equal(t, false, nonOracleError(err1))
	require.Equal(t, false, nonOracleError(err2))
	require.Equal(t, false, nonOracleError(err3))
	require.Equal(t, true, nonOracleError(err4))
	require.Equal(t, true, nonOracleError(err5))
	require.Equal(t, true, nonOracleError(err6))
	require.Equal(t, true, nonOracleError(err7))
}
