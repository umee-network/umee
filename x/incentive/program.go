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

// Validate performs validation on an IncentiveProgram type returning an error
// if the program is invalid.
func (ip IncentiveProgram) Validate() error {
	if err := sdk.ValidateDenom(ip.UToken); err != nil {
		return err
	}
	if !leveragetypes.HasUTokenPrefix(ip.UToken) {
		// only allow uToken denoms as bonded denoms
		return errors.Wrap(leveragetypes.ErrNotUToken, ip.UToken)
	}

	if err := ip.TotalRewards.Validate(); err != nil {
		return err
	}
	if leveragetypes.HasUTokenPrefix(ip.TotalRewards.Denom) {
		// only allow base token denoms as rewards
		return errors.Wrap(leveragetypes.ErrUToken, ip.TotalRewards.Denom)
	}

	if err := ip.RemainingRewards.Validate(); err != nil {
		return err
	}

	// TODO #1749: Finish validate logic

	return nil
}
