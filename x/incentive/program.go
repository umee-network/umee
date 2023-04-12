package incentive

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// validateProposedIncentiveProgram runs IncentiveProgram.Validate and also checks additional requirements applying
// to incentive programs which have not yet been funded or passed by governance
func validateProposedIncentiveProgram(program IncentiveProgram) error {
	if program.ID != 0 {
		return ErrInvalidProgramID.Wrapf("%d", program.ID)
	}
	if !program.RemainingRewards.IsZero() {
		return ErrNonzeroRemainingRewards.Wrap(program.RemainingRewards.String())
	}
	if program.Funded {
		return ErrProposedFundedProgram
	}
	return program.Validate()
}

// validatePassedIncentiveProgram runs IncentiveProgram.Validate and also checks additional requirements applying
// to incentive programs which have already been passed by governance
func validatePassedIncentiveProgram(program IncentiveProgram) error {
	if program.ID == 0 {
		return ErrInvalidProgramID.Wrapf("%d", program.ID)
	}
	return program.Validate()
}

// Validate performs validation on an IncentiveProgram type returning an error
// if the program is invalid.
func (ip IncentiveProgram) Validate() error {
	if err := sdk.ValidateDenom(ip.UDenom); err != nil {
		return err
	}
	if !leveragetypes.HasUTokenPrefix(ip.UDenom) {
		// only allow uToken denoms as bonded denoms
		return errors.Wrap(leveragetypes.ErrNotUToken, ip.UDenom)
	}

	if err := ip.TotalRewards.Validate(); err != nil {
		return err
	}
	if leveragetypes.HasUTokenPrefix(ip.TotalRewards.Denom) {
		// only allow base token denoms as rewards
		return errors.Wrap(leveragetypes.ErrUToken, ip.TotalRewards.Denom)
	}
	if ip.TotalRewards.IsZero() {
		return ErrProgramWithoutRewards
	}

	if err := ip.RemainingRewards.Validate(); err != nil {
		return err
	}
	if ip.RemainingRewards.Denom != ip.TotalRewards.Denom {
		return ErrProgramRewardMismatch
	}
	if !ip.Funded && ip.RemainingRewards.IsPositive() {
		return ErrNonfundedProgramRewards
	}

	if ip.Duration <= 0 {
		return errors.Wrapf(ErrInvalidProgramDuration, "%d", ip.Duration)
	}
	if ip.StartTime <= 0 {
		return errors.Wrapf(ErrInvalidProgramStart, "%d", ip.Duration)
	}

	return nil
}
