package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

// InitGenesis initializes the x/leverage module state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, token := range genState.Registry {
		util.Panic(k.SetTokenSettings(ctx, token))
	}

	for _, borrow := range genState.AdjustedBorrows {
		borrower, err := sdk.AccAddressFromBech32(borrow.Address)
		util.Panic(err)
		util.Panic(k.setAdjustedBorrow(ctx, borrower, borrow.Amount))
	}

	for _, collateral := range genState.Collateral {
		borrower, err := sdk.AccAddressFromBech32(collateral.Address)
		util.Panic(err)
		util.Panic(k.setCollateral(ctx, borrower, collateral.Amount))
	}

	for _, reserve := range genState.Reserves {
		util.Panic(k.setReserves(ctx, reserve))
	}

	util.Panic(k.setLastInterestTime(ctx, genState.LastInterestTime))

	for _, badDebt := range genState.BadDebts {
		borrower, err := sdk.AccAddressFromBech32(badDebt.Address)
		util.Panic(err)
		util.Panic(k.setBadDebtAddress(ctx, borrower, badDebt.Denom, true))
	}

	for _, rate := range genState.InterestScalars {
		util.Panic(k.setInterestScalar(ctx, rate.Denom, rate.Scalar))
	}

	for _, pair := range genState.SpecialPairs {
		util.Panic(k.SetSpecialAssetPair(ctx, pair))
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
		k.getLastInterestTime(ctx),
		k.getAllBadDebts(ctx),
		k.getAllInterestScalars(ctx),
		k.GetAllUTokenSupply(ctx),
		k.GetAllSpecialAssetPairs(ctx),
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

		var amount sdk.Dec
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled borrow amount should never happen
			return err
		}

		borrows = append(borrows, types.NewAdjustedBorrow(addr.String(), sdk.NewDecCoinFromDec(denom, amount)))
		return nil
	}

	util.Panic(k.iterate(ctx, prefix, iterator))

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

		var amount sdkmath.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled collateral amount should never happen
			return err
		}

		collateral = append(collateral, types.NewCollateral(addr.String(), sdk.NewCoin(denom, amount)))
		return nil
	}

	util.Panic(k.iterate(ctx, prefix, iterator))

	return collateral
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

	util.Panic(k.iterate(ctx, prefix, iterator))

	return interestScalars
}
