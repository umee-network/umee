package uibc

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrIBCPauseStatus     = sdkerrors.Register(ModuleName, 1, "invalid ibc pause status")
	ErrQuotaExceeded      = sdkerrors.Register(ModuleName, 2, "quota transfer exceeded")
	ErrNoQuotaForIBCDenom = sdkerrors.Register(ModuleName, 3, "no quota for ibc denom")
	ErrInvalidIBCDenom    = sdkerrors.Register(ModuleName, 4, "invalid ibc denom")
	ErrInvalidQuota       = sdkerrors.Register(ModuleName, 5, "invalid quota")
	ErrInvalidAmount      = sdkerrors.Register(ModuleName, 6, "invalid transfer amount")
)
