package incentive

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

var (
	// 0 = TODO
	ErrNotImplemented = errors.Register(ModuleName, 0, "not implemented")

	// 1XX = General
	ErrInvalidProgramID     = errors.Register(ModuleName, 100, "invalid program ID")
	ErrNilAsset             = errors.Register(ModuleName, 101, "nil asset")
	ErrInvalidTier          = errors.Register(ModuleName, 102, "invalid unbonding tier")
	ErrEmptyAddress         = errors.Register(ModuleName, 103, "empty address")
	ErrInvalidProgramStatus = errors.Register(ModuleName, 104, "invalid program status")

	// 2XX = Params
	ErrUnbondingTierOrder   = errors.Register(ModuleName, 200, "unbonding tier lock durations out of order")
	ErrUnbondingWeightOrder = errors.Register(ModuleName, 201, "unbonding tier weights out of order")

	// 3XX = Gov Proposal
	ErrNonzeroRemainingRewards = errors.Register(ModuleName, 300, "remaining rewards must be zero in proposal")
	ErrProposedFundedProgram   = errors.Register(ModuleName, 301, "proposed program must have funded = false")
	ErrEmptyProposal           = errors.Register(ModuleName, 302, "proposal contains no incentive programs")

	// 4XX = Messages
	ErrSponsorIneligible      = errors.Register(ModuleName, 400, "incentive program not eligible for sponsorship")
	ErrSponsorInvalid         = errors.Register(ModuleName, 401, "incorrect funding for incentive program")
	ErrMaxUnbondings          = errors.Register(ModuleName, 402, "exceeds max concurrent unbondings for a single uToken")
	ErrInsufficientBonded     = errors.Register(ModuleName, 403, "insufficient bonded, but not already unbonding, uTokens")
	ErrInsufficientCollateral = errors.Register(ModuleName, 404, "insufficient collateral to create bond")

	// 5XX = Misc
	ErrDecreaseNextProgramID     = errors.Register(ModuleName, 500, "cannot decrease NextProgramID")
	ErrDecreaseLastRewardTime    = errors.Register(ModuleName, 501, "cannot decrease LastRewardTime")
	ErrChangeAccumulatorExponent = errors.Register(ModuleName, 502, "cannot change RewardAccumulator.Exponent once set")
)
