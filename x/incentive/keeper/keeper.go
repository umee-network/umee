package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/v2/x/incentive/types"
)

type Keeper struct {
	cdc            codec.Codec
	storeKey       sdk.StoreKey
	paramSpace     paramtypes.Subspace
	bankKeeper     types.BankKeeper
	leverageKeeper types.LeverageKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	bk types.BankKeeper,
	lk types.LeverageKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		paramSpace:     paramSpace,
		bankKeeper:     bk,
		leverageKeeper: lk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ModuleBalance returns the amount of a given token held in the x/incentive module account
func (k Keeper) ModuleBalance(ctx sdk.Context, denom string) sdk.Int {
	return k.bankKeeper.SpendableCoins(ctx, authtypes.NewModuleAddress(types.ModuleName)).AmountOf(denom)
}

/*
// Claim attempts to claim any pending rewards belonging to an address.
func (k Keeper) LendAsset(ctx sdk.Context, addr sdk.AccAddress) error {
	if err := k.AssertLendEnabled(ctx, loan.Denom); err != nil {
		return err
	}

	// determine uToken amount to mint
	uToken, err := k.ExchangeToken(ctx, loan)
	if err != nil {
		return err
	}

	// send token balance to leverage module account
	loanTokens := sdk.NewCoins(loan)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, lenderAddr, types.ModuleName, loanTokens); err != nil {
		return err
	}

	// mint uToken and set new total uToken supply
	uTokens := sdk.NewCoins(uToken)
	if err = k.bankKeeper.MintCoins(ctx, types.ModuleName, uTokens); err != nil {
		return err
	}
	if err = k.setUTokenSupply(ctx, k.GetUTokenSupply(ctx, uToken.Denom).Add(uToken)); err != nil {
		return err
	}

	if k.GetCollateralSetting(ctx, lenderAddr, uToken.Denom) {
		// For uToken denoms enabled as collateral by this lender, the
		// minted uTokens stay in the module account and the keeper tracks the amount.
		currentCollateral := k.GetCollateralAmount(ctx, lenderAddr, uToken.Denom)
		if err = k.setCollateralAmount(ctx, lenderAddr, currentCollateral.Add(uToken)); err != nil {
			return err
		}
	} else if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lenderAddr, uTokens); err != nil {
		// For uToken denoms not enabled as collateral by this lender, the uTokens are sent to lender address
		return err
	}

	return nil
}
*/
