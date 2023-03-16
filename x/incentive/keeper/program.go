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
	if err := program.Validate(); err != nil {
		return err
	}

	addr := k.GetCommunityFundAddress(ctx)
	if fromCommunityFund && !addr.Empty() {
		// If the module has set a community fund address and the proposal
		// requested it, we can attempt to instantly fund the module when
		// the proposal passes.
		funds := k.bankKeeper.SpendableCoins(ctx, addr)
		rewards := sdk.NewCoins(program.TotalRewards)
		if funds.IsAllGT(rewards) {
			// Community fund has the required tokens to fund the program
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, incentive.ModuleName, rewards); err != nil {
				return err
			}
			// Set program's funded and remaining rewards to the amount just funded
			program.FundedRewards = program.TotalRewards
			program.RemainingRewards = program.TotalRewards
		}
	}

	// Set program's ID to the next available value and store it in upcoming incentive programs
	id := k.getNextProgramID(ctx)
	program.Id = id
	if err := k.SetIncentiveProgram(ctx, program, incentive.ProgramStatusUpcoming); err != nil {
		return err
	}

	// Increment module's NextProgramID
	return k.setNextProgramID(ctx, id+1)
}
