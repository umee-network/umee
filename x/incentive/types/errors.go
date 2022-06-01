package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNotUtoken    = sdkerrors.Register(ModuleName, 1200, "not a uToken denom")
	ErrProgramEnded = sdkerrors.Register(ModuleName, 1201, "incentive program has ended")
)
