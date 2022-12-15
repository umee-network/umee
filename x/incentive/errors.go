package incentive

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// 0 = TODO
	ErrNotImplemented = sdkerrors.Register(ModuleName, 0, "not implemented")

	// 1XX = General
	ErrInvalidProgramID = sdkerrors.Register(ModuleName, 100, "invalid program ID")
	ErrNilAsset         = sdkerrors.Register(ModuleName, 101, "nil asset")
	ErrInvalidTier      = sdkerrors.Register(ModuleName, 102, "invalid unbonding tier")
	ErrEmptyAddress     = sdkerrors.Register(ModuleName, 103, "empty address")

	// 2XX = Params
	ErrUnbondingTierOrder   = sdkerrors.Register(ModuleName, 200, "unbonding tier lock durations out of order")
	ErrUnbondingWeightOrder = sdkerrors.Register(ModuleName, 201, "unbonding tier weights out of order")

	// 3XX = Gov Proposal
	ErrNonzeroRemainingRewards = sdkerrors.Register(ModuleName, 300, "remaining rewards must be zero in proposal")
	ErrNonzeroFundedRewards    = sdkerrors.Register(ModuleName, 301, "funded rewards must be zero in proposal")
)
