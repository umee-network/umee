package keeper

import (
	"strings"

	"cosmossdk.io/errors"

	"github.com/umee-network/umee/v4/util/decmath"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"
)

// nonOracleError returns true if an error is non-nil
// and also not one of ErrEmptyList, ErrUnknownDenom, or ErrNoHistoricMedians
// which are errors which can result from missing prices
func nonOracleError(err error) bool {
	if err == nil {
		return false
	}
	// check typed errors
	if errors.IsOf(err,
		leveragetypes.ErrInvalidOraclePrice,
		leveragetypes.ErrNoHistoricMedians,
		oracletypes.ErrUnknownDenom,
	) {
		return false
	}
	// this error needs to be checked by string comparison
	if strings.Contains(err.Error(), decmath.ErrEmptyList.Error()) {
		return false
	}
	return true
}
