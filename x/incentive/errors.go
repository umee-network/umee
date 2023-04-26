package incentive

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

var (
	// 1XX = General
	ErrInvalidProgramID        = errors.Register(ModuleName, 100, "invalid program ID")
	ErrNilAsset                = errors.Register(ModuleName, 101, "nil asset")
	ErrEmptyAddress            = errors.Register(ModuleName, 102, "empty address")
	ErrInvalidProgramStatus    = errors.Register(ModuleName, 103, "invalid program status")
	ErrInvalidProgramDuration  = errors.Register(ModuleName, 104, "invalid program duration")
	ErrInvalidProgramStart     = errors.Register(ModuleName, 105, "invalid program start time")
	ErrProgramRewardMismatch   = errors.Register(ModuleName, 106, "program total and remaining reward denoms mismatched")
	ErrNonfundedProgramRewards = errors.Register(ModuleName, 107, "nonzero remaining rewards on a non-funded program")
	ErrProgramWithoutRewards   = errors.Register(ModuleName, 108, "incentive program must have nonzero rewards")
	ErrInvalidUnbonding        = errors.Register(ModuleName, 109, "invalid unbonding")

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
	ErrDecreaseNextProgramID  = errors.Register(ModuleName, 500, "cannot decrease NextProgramID")
	ErrDecreaseLastRewardTime = errors.Register(ModuleName, 501, "cannot decrease LastRewardTime")
)
