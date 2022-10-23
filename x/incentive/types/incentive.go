package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	leveragetypes "github.com/umee-network/umee/v3/x/leverage/types"
)

// Validate performs validation on an IncentiveProgram type returning an error
// if the program is invalid.
func (ip IncentiveProgram) Validate() error {
	if err := sdk.ValidateDenom(ip.LockDenom); err != nil {
		return err
	}
	if !strings.HasPrefix(ip.LockDenom, leveragetypes.UTokenPrefix) {
		// only allow base asset denoms that start with "u/"
		return sdkerrors.Wrap(ErrNotUToken, ip.LockDenom)
	}

	// TODO: Finish validate logic

	return nil
}
