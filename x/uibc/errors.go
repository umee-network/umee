package uibc

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	ErrQuotaExceeded      = sdkerrors.Register(ModuleName, 1, "quota transfer exceeded")
	ErrNoQuotaForIBCDenom = sdkerrors.Register(ModuleName, 2, "no quota for ibc denom")
)
