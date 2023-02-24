package incentive

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

var (
	// 0 = TODO
	ErrNotImplemented = errors.Register(ModuleName, 0, "not implemented")

	// 1XX = General
	ErrInvalidProgramID = errors.Register(ModuleName, 100, "invalid program ID")
	ErrNilAsset         = errors.Register(ModuleName, 101, "nil asset")
	ErrInvalidTier      = errors.Register(ModuleName, 102, "invalid unbonding tier")
	ErrEmptyAddress     = errors.Register(ModuleName, 103, "empty address")

	// 2XX = Params
	ErrUnbondingTierOrder   = errors.Register(ModuleName, 200, "unbonding tier lock durations out of order")
	ErrUnbondingWeightOrder = errors.Register(ModuleName, 201, "unbonding tier weights out of order")

	// 3XX = Gov Proposal
	ErrNonzeroRemainingRewards = errors.Register(ModuleName, 300, "remaining rewards must be zero in proposal")
	ErrNonzeroFundedRewards    = errors.Register(ModuleName, 301, "funded rewards must be zero in proposal")
	ErrEmptyProposal           = errors.Register(ModuleName, 302, "proposal contains no incentive programs")

	// 5XX = Misc
	ErrDecreaseNextProgramID  = errors.Register(ModuleName, 500, "cannot decrease NextProgramID")
	ErrDecreaseLastRewardTime = errors.Register(ModuleName, 501, "cannot decrease LastRewardTime")
)
