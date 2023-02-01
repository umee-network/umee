package uibc

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrQuotaExceeded      = sdkerrors.Register(ModuleName, 1, "quota transfer exceeded")
	ErrNoQuotaForIBCDenom = sdkerrors.Register(ModuleName, 2, "no quota for ibc denom")
)
