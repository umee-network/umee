package keeper

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/metoken"
)

func (k Keeper) Bond() error {
	if k.ctx.BlockTime().Before(k.getNextBondingTime()) {
		return nil
	}

	if params := k.incentiveKeeper.GetParams(*k.ctx); params.UnbondingDuration != 0 {
		return nil
	}

	if ok, err := k.incentiveKeeper.HasOngoingPrograms(*k.ctx); err != nil {
		k.Logger().Debug(
			"bonding: couldn't get ongoing programs",
			"error", err.Error(),
			"block_time", k.ctx.BlockTime(),
		)
		return nil
	} else if !ok {
		return nil
	}

	for _, index := range k.GetAllRegisteredIndexes() {
		for _, aa := range index.AcceptedAssets {
			if ok, err := k.incentiveKeeper.HasOngoingProgramsByDenom(*k.ctx, aa.Denom); err != nil {
				k.Logger().Debug(
					"bonding: couldn't get ongoing programs for denom",
					"denom", aa.Denom,
					"error", err.Error(),
					"block_time", k.ctx.BlockTime(),
				)
			} else if !ok {
				continue
			}

			err := k.bondByDenom(aa.Denom)
			// every error returned here will be non-recoverable
			if err != nil {
				return err
			}
		}
	}

	//TODO: confirm bonding period = zero
	//      get ongoing campaigns
	//      filter by denom
	//      if find some: collateralize and bond

	return nil
}

func (k Keeper) bondByDenom(assetDenom string) error {
	meTokenAddr := authtypes.NewModuleAddress(metoken.ModuleName)
	uTokenDenom := coin.ToUTokenDenom(assetDenom)
	collateral := k.leverageKeeper.GetCollateral(*k.ctx, meTokenAddr, uTokenDenom)
	supplied, err := k.leverageKeeper.GetSupplied(*k.ctx, meTokenAddr, uTokenDenom)
	if err != nil {
		k.Logger().Debug(
			"bonding: couldn't get supplied for denom",
			"denom", uTokenDenom,
			"error", err.Error(),
			"block_time", k.ctx.BlockTime(),
		)
		return nil
	}

	if collateral.IsLT(supplied) {
		toCollateralize := supplied.Sub(collateral)
		recoverable, err := k.leverageKeeper.CollateralizeForModule(*k.ctx, metoken.ModuleName, toCollateralize)
		if err != nil {
			if recoverable {
				k.Logger().Debug(
					"bonding: couldn't collateralize for module",
					"toCollateralize", toCollateralize.String(),
					"error", err.Error(),
					"block_time", k.ctx.BlockTime(),
				)
				return nil
			}

			return err
		}
	}

	bonded := k.incentiveKeeper.GetBonded(*k.ctx, meTokenAddr, uTokenDenom)
	if bonded.IsGTE(supplied) {
		return nil
	}

	toBond := supplied.Sub(bonded)
	recoverable, err := k.incentiveKeeper.BondForModule(*k.ctx, metoken.ModuleName, toBond)
	if err != nil {
		if recoverable {
			k.Logger().Debug(
				"bonding: couldn't bond for module",
				"toBond", toBond.String(),
				"error", err.Error(),
				"block_time", k.ctx.BlockTime(),
			)
			return nil
		}

		return err
	}

	return nil
}
