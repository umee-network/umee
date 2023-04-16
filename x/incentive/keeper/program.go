package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/incentive"
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

	// Set program's ID to the next available value and store it in upcoming incentive programs
	id := k.getNextProgramID(ctx)
	program.ID = id
	if err := k.setIncentiveProgram(ctx, program, incentive.ProgramStatusUpcoming); err != nil {
		return err
	}

	// Increment module's NextProgramID
	return k.setNextProgramID(ctx, id+1)
}
