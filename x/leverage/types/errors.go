package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidAsset = sdkerrors.Register(ModuleName, 1100, "invalid asset")
)
