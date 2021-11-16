package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInternal                = sdkerrors.Register(ModuleName, 2, "internal")
	ErrDuplicate               = sdkerrors.Register(ModuleName, 3, "duplicate")
	ErrInvalid                 = sdkerrors.Register(ModuleName, 4, "invalid")
	ErrTimeout                 = sdkerrors.Register(ModuleName, 5, "timeout")
	ErrUnknown                 = sdkerrors.Register(ModuleName, 6, "unknown")
	ErrEmpty                   = sdkerrors.Register(ModuleName, 7, "empty")
	ErrOutdated                = sdkerrors.Register(ModuleName, 8, "outdated")
	ErrUnsupported             = sdkerrors.Register(ModuleName, 9, "unsupported")
	ErrNonContiguousEventNonce = sdkerrors.Register(ModuleName, 10, "non contiguous event nonce")
	ErrNoUnbatchedTxsFound     = sdkerrors.Register(ModuleName, 11, "no unbatched txs found")
	ErrSetOrchAddresses        = sdkerrors.Register(ModuleName, 12, "failed to set orchestrator addresses")
	ErrSupplyOverflow          = sdkerrors.Register(ModuleName, 13, "supply cannot exceed max ERC20 value")
	ErrInvalidERC20Event       = sdkerrors.Register(ModuleName, 14, "invalid ERC20 deployed event")
	ErrEmptyEthSig             = sdkerrors.Register(ModuleName, 15, "empty Ethereum signature")
)
