package ibctransfer

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrIBCPauseStatus          = sdkerrors.Register(ModuleName, 1, "invalid ibc pause status")
	ErrRateLimitExceeded       = sdkerrors.Register(ModuleName, 2, "rate limit exceeded")
	ErrBadMessage              = sdkerrors.Register(ModuleName, 3, "bad message")
	ErrContractError           = sdkerrors.Register(ModuleName, 4, "contract error")
	ErrBadPacket               = sdkerrors.Register(ModuleName, 5, "bad packet")
	ErrNoRateLimitsForIBCDenom = sdkerrors.Register(ModuleName, 6, "no rate limits for ibc denom")
)
