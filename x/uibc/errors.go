package uibc

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrIBCPauseStatus     = sdkerrors.Register(ModuleName, 1, "invalid ibc pause status")
	ErrQuotaExceeded      = sdkerrors.Register(ModuleName, 2, "quota transfer exceeded")
	ErrBadMessage         = sdkerrors.Register(ModuleName, 3, "bad message")
	ErrContractError      = sdkerrors.Register(ModuleName, 4, "contract error")
	ErrBadPacket          = sdkerrors.Register(ModuleName, 5, "bad packet")
	ErrNoQuotaForIBCDenom = sdkerrors.Register(ModuleName, 6, "no quota for ibc denom")
	ErrInvalidIBCDenom    = sdkerrors.Register(ModuleName, 7, "invalid ibc denom")
	ErrInvalidQuota       = sdkerrors.Register(ModuleName, 8, "invalid quota")
)
