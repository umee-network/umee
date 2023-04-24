package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/incentive"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// createIncentiveProgram saves an incentive program to upcoming programs after it
// passes governance, and also attempts to fund it from the module's community fund
// address if sufficient funds are available. The program is always added to upcoming
// even if funding fails or its start date has already passed, but an error is returned
// instead if it fails validation.
func (k Keeper) createIncentiveProgram(
	ctx sdk.Context,
	program incentive.IncentiveProgram,
	fromCommunityFund bool,
) error {
	if err := program.ValidateProposed(); err != nil {
		return err
	}

	addr := k.GetParams(ctx).CommunityFundAddress
	if fromCommunityFund {
		if addr != "" {
			communityAddress := sdk.MustAccAddressFromBech32(addr)
			// If the module has set a community fund address and the proposal
			// requested it, we can attempt to instantly fund the module when
			// the proposal passes.
			funds := k.bankKeeper.SpendableCoins(ctx, communityAddress)
			rewards := sdk.NewCoins(program.TotalRewards)
			if funds.IsAllGT(rewards) {
				// Community fund has the required tokens to fund the program
				err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, communityAddress, incentive.ModuleName, rewards)
				if err != nil {
					return err
				}
				// Set program's funded and remaining rewards to the amount just funded
				program.Funded = true
				program.RemainingRewards = program.TotalRewards
			} else {
				ctx.Logger().Error("incentive community fund insufficient. proposal will revert to manual funding.")
			}
		} else {
			ctx.Logger().Error("incentive community fund not set. proposal will revert to manual funding.")
		}
	}

	// If this is the first time this uToken has been incentivized, initialize its rewardAccumulator exponent.
	// Note that this interprets Exponent == 0 as needing initialization, but if an asset actually had exponent zero,
	// and had already been initialized, this would be a harmless no-op.
	if ra := k.getRewardAccumulator(ctx, program.UToken); ra.Exponent == 0 {
		token, err := k.leverageKeeper.GetTokenSettings(ctx, leveragetypes.ToTokenDenom(program.UToken))
		if err != nil {
			// unregistered tokens do not have uTokens, so they cannot be incentivized
			return err
		}
		// Set exponent, preserving all other fields in the reward accumulator in case they are not zero
		ra.Exponent = token.Exponent
		if err = k.setRewardAccumulator(ctx, ra); err != nil {
			return err
		}
	}

	// Set program's ID to the next available value and store it in upcoming incentive programs
	id := k.getNextProgramID(ctx)
	if id == 0 {
		return incentive.ErrInvalidProgramID.Wrap("next id was zero")
	}
	program.ID = id

	// Increment module's NextProgramID
	if err := k.setNextProgramID(ctx, id+1); err != nil {
		return err
	}
	return k.setIncentiveProgram(ctx, program, incentive.ProgramStatusUpcoming)
}

// moveIncentiveProgram moves an incentive program from one status to another.
// Valid status changes are upcoming -> [ongoing, completed] and ongoing -> upcoming
func (k Keeper) moveIncentiveProgram(ctx sdk.Context, id uint32, toStatus incentive.ProgramStatus) error {
	if id == 0 {
		return incentive.ErrInvalidProgramID.Wrap("zero")
	}
	program, fromStatus, err := k.getIncentiveProgram(ctx, id)
	if err != nil {
		return err
	}

	// enforces strict status change rules
	switch fromStatus {
	case incentive.ProgramStatusCompleted:
		return incentive.ErrInvalidProgramStatus.Wrap("cannot move program from completed status")
	case incentive.ProgramStatusOngoing:
		if toStatus != incentive.ProgramStatusCompleted {
			return incentive.ErrInvalidProgramStatus.Wrap("ongoing programs can only be moved to completed")
		}
	case incentive.ProgramStatusUpcoming:
		if toStatus != incentive.ProgramStatusOngoing && toStatus != incentive.ProgramStatusCompleted {
			return incentive.ErrInvalidProgramStatus.Wrap("upcoming programs can be moved to ongoing or completed")
		}
	default:
		return incentive.ErrInvalidProgramStatus
	}

	if err := k.deleteIncentiveProgram(ctx, id); err != nil {
		return err
	}

	// add program to new status
	return k.setIncentiveProgram(ctx, program, toStatus)
}

// sponsorIncentiveProgram sponsors an incentive program from an account and updates
// it in the store.
func (k Keeper) sponsorIncentiveProgram(ctx sdk.Context, sponsor sdk.AccAddress, id uint32) error {
	if id == 0 {
		return incentive.ErrInvalidProgramID.Wrap("zero")
	}

	// Error messages that follow are designed to promote third party usability, so they are more
	// verbose and situational than usual.
	program, status, err := k.getIncentiveProgram(ctx, id)
	if err != nil {
		return err
	}
	if status != incentive.ProgramStatusUpcoming {
		return incentive.ErrSponsorIneligible.Wrap("program exists but does not have status upcoming")
	}
	if program.Funded {
		return incentive.ErrSponsorIneligible.Wrap("program is already funded")
	}
	spendable := k.bankKeeper.SpendableCoins(ctx, sponsor).AmountOf(program.TotalRewards.Denom)
	if spendable.LT(program.TotalRewards.Amount) {
		return incentive.ErrSponsorInvalid.Wrapf("insufficient sponsor tokens: want %s, have %s",
			program.TotalRewards, spendable)
	}

	// transfer rewards from sponsor to incentive module
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx,
		sponsor,
		incentive.ModuleName,
		sdk.NewCoins(program.TotalRewards),
	)
	if err != nil {
		return err
	}

	// set update program state
	program.Funded = true
	program.RemainingRewards = program.TotalRewards

	// add program to new status
	return k.setIncentiveProgram(ctx, program, status)
}
