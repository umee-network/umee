package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrIBCPauseStatus = sdkerrors.Register(ModuleName, 1, "invalid ibc pause status")
)
