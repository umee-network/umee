package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInternal                = sdkerrors.Register(ModuleName, 1, "internal")
	ErrDuplicate               = sdkerrors.Register(ModuleName, 2, "duplicate")
	ErrInvalid                 = sdkerrors.Register(ModuleName, 3, "invalid")
	ErrTimeout                 = sdkerrors.Register(ModuleName, 4, "timeout")
	ErrUnknown                 = sdkerrors.Register(ModuleName, 5, "unknown")
	ErrEmpty                   = sdkerrors.Register(ModuleName, 6, "empty")
	ErrOutdated                = sdkerrors.Register(ModuleName, 7, "outdated")
	ErrUnsupported             = sdkerrors.Register(ModuleName, 8, "unsupported")
	ErrNonContiguousEventNonce = sdkerrors.Register(ModuleName, 9, "non contiguous event nonce")
	ErrNoUnbatchedTxsFound     = sdkerrors.Register(ModuleName, 10, "no unbatched txs found")
	ErrResetDelegateKeys       = sdkerrors.Register(ModuleName, 11, "can not set orchestrator addresses more than once")
	ErrSupplyOverflow          = sdkerrors.Register(ModuleName, 12, "supply cannot exceed max ERC20 value")
)
