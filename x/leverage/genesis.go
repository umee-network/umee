package leverage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/x/leverage/keeper"
	"github.com/umee-network/umee/x/leverage/types"
)

// InitGenesis initializes the x/leverage module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, token := range genState.Registry {
		k.SetRegisteredToken(ctx, token)
	}

	for _, borrow := range genState.Borrows {
		borrower, err := sdk.AccAddressFromBech32(borrow.Address)
		if err != nil {
			panic(err)
		}

		if err = k.SetBorrow(ctx, borrower, borrow.Amount); err != nil {
			panic(err)
		}
	}

	for _, setting := range genState.CollateralSettings {
		borrower, err := sdk.AccAddressFromBech32(setting.Address)
		if err != nil {
			panic(err)
		}

		if err = k.SetCollateralSetting(ctx, borrower, setting.Denom, true); err != nil {
			panic(err)
		}
	}

	for _, collateral := range genState.Collateral {
		borrower, err := sdk.AccAddressFromBech32(collateral.Address)
		if err != nil {
			panic(err)
		}

		if err = k.SetCollateralAmount(ctx, borrower, collateral.Amount); err != nil {
			panic(err)
		}
	}

	for _, reserve := range genState.Reserves {
		if err := k.SetReserveAmount(ctx, reserve); err != nil {
			panic(err)
		}
	}

	if err := k.SetLastInterestTime(ctx, genState.LastInterestTime); err != nil {
		panic(err)
	}

	for _, rate := range genState.ExchangeRates {
		if err := k.SetExchangeRate(ctx, rate.Denom, rate.Amount); err != nil {
			panic(err)
		}
	}

	for _, badDebt := range genState.BadDebts {
		borrower, err := sdk.AccAddressFromBech32(badDebt.Address)
		if err != nil {
			panic(err)
		}

		k.SetBadDebtAddress(ctx, badDebt.Denom, borrower, true)
	}

	for _, rate := range genState.Borrow_APYs {
		if err := k.SetBorrowAPY(ctx, rate.Denom, rate.Amount); err != nil {
			panic(err)
		}
	}

	for _, rate := range genState.Lend_APYs {
		if err := k.SetLendAPY(ctx, rate.Denom, rate.Amount); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the x/leverage module's exported genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {

	tokens, err := k.GetAllRegisteredTokens(ctx)
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(
		k.GetParams(ctx),
		tokens,
		k.GetAllBorrows(ctx),
		k.GetAllCollateralSettings(ctx),
		k.GetAllCollateral(ctx),
		k.GetAllReserves(ctx),
		k.GetLastInterestTime(ctx),
		k.GetAllExchangeRates(ctx),
		k.GetAllBadDebts(ctx),
		k.GetAllBorrowAPY(ctx),
		k.GetAllLendAPY(ctx),
	)
}
