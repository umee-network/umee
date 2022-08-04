package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// InitGenesis initializes the x/leverage module state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, token := range genState.Registry {
		if err := k.SetTokenSettings(ctx, token); err != nil {
			panic(err)
		}
	}

	for _, borrow := range genState.AdjustedBorrows {
		borrower, err := sdk.AccAddressFromBech32(borrow.Address)
		if err != nil {
			panic(err)
		}

		if err = k.setAdjustedBorrow(ctx, borrower, borrow.Amount); err != nil {
			panic(err)
		}
	}

	for _, collateral := range genState.Collateral {
		borrower, err := sdk.AccAddressFromBech32(collateral.Address)
		if err != nil {
			panic(err)
		}

		if err = k.setCollateralAmount(ctx, borrower, collateral.Amount); err != nil {
			panic(err)
		}
	}

	for _, reserve := range genState.Reserves {
		if err := k.setReserveAmount(ctx, reserve); err != nil {
			panic(err)
		}
	}

	if err := k.SetLastInterestTime(ctx, genState.LastInterestTime); err != nil {
		panic(err)
	}

	for _, badDebt := range genState.BadDebts {
		borrower, err := sdk.AccAddressFromBech32(badDebt.Address)
		if err != nil {
			panic(err)
		}

		if err := k.setBadDebtAddress(ctx, borrower, badDebt.Denom, true); err != nil {
			panic(err)
		}
	}

	for _, rate := range genState.InterestScalars {
		if err := k.setInterestScalar(ctx, rate.Denom, rate.Scalar); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the x/leverage module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return types.NewGenesisState(
		k.GetParams(ctx),
		k.GetAllRegisteredTokens(ctx),
		k.getAllAdjustedBorrows(ctx),
		k.getAllCollateral(ctx),
		k.GetAllReserves(ctx),
		k.GetLastInterestTime(ctx),
		k.getAllBadDebts(ctx),
		k.getAllInterestScalars(ctx),
		k.GetAllUTokenSupply(ctx),
	)
}

// getAllAdjustedBorrows returns all borrows across all borrowers and asset types. Uses the
// AdjustedBorrow struct found in GenesisState, which stores amount scaled by InterestScalar.
func (k Keeper) getAllAdjustedBorrows(ctx sdk.Context) []types.AdjustedBorrow {
	prefix := types.KeyPrefixAdjustedBorrow
	borrows := []types.AdjustedBorrow{}

	iterator := func(key, val []byte) error {
		addr := types.AddressFromKey(key, prefix)
		denom := types.DenomFromKeyWithAddress(key, prefix)

		var amount sdk.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled borrow amount should never happen
			return err
		}

		borrows = append(borrows, types.NewAdjustedBorrow(addr.String(), sdk.NewDecCoin(denom, amount)))
		return nil
	}

	err := k.iterate(ctx, prefix, iterator)
	if err != nil {
		panic(err)
	}

	return borrows
}

// getAllCollateral returns all collateral across all borrowers and asset types. Uses the
// CollateralAmount struct found in GenesisState, which stores borrower address as a string.
func (k Keeper) getAllCollateral(ctx sdk.Context) []types.Collateral {
	prefix := types.KeyPrefixCollateralAmount
	collateral := []types.Collateral{}

	iterator := func(key, val []byte) error {
		addr := types.AddressFromKey(key, prefix)
		denom := types.DenomFromKeyWithAddress(key, prefix)

		var amount sdk.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled collateral amount should never happen
			return err
		}

		collateral = append(collateral, types.NewCollateral(addr.String(), sdk.NewCoin(denom, amount)))
		return nil
	}

	err := k.iterate(ctx, prefix, iterator)
	if err != nil {
		panic(err)
	}

	return collateral
}

// getAllBadDebts gets bad debt instances across all borrowers. Uses the
// BadDebt struct found  in GenesisState.
func (k Keeper) getAllBadDebts(ctx sdk.Context) []types.BadDebt {
	prefix := types.KeyPrefixBadDebt
	badDebts := []types.BadDebt{}

	iterator := func(key, _ []byte) error {
		addr := types.AddressFromKey(key, prefix)
		denom := types.DenomFromKeyWithAddress(key, prefix)

		badDebts = append(badDebts, types.NewBadDebt(addr.String(), denom))

		return nil
	}

	err := k.iterate(ctx, prefix, iterator)
	if err != nil {
		panic(err)
	}

	return badDebts
}

// getAllInterestScalars returns all interest scalars. Uses the InterestScalar struct found
// in GenesisState.
func (k Keeper) getAllInterestScalars(ctx sdk.Context) []types.InterestScalar {
	prefix := types.KeyPrefixInterestScalar
	interestScalars := []types.InterestScalar{}

	iterator := func(key, val []byte) error {
		denom := types.DenomFromKey(key, prefix)

		var scalar sdk.Dec
		if err := scalar.Unmarshal(val); err != nil {
			// improperly marshaled interest scalar should never happen
			return err
		}

		interestScalars = append(interestScalars, types.NewInterestScalar(denom, scalar))
		return nil
	}

	err := k.iterate(ctx, prefix, iterator)
	if err != nil {
		panic(err)
	}

	return interestScalars
}
