package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNotUtoken        = sdkerrors.Register(ModuleName, 1200, "not a uToken denom")
	ErrProgramEnded     = sdkerrors.Register(ModuleName, 1201, "incentive program has ended")
	ErrInvalidTier      = sdkerrors.Register(ModuleName, 1202, "tier must be between 1 and 3")
	ErrInvalidProgramID = sdkerrors.Register(ModuleName, 1203, "program id must be greater than zero")
)
