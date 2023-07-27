package keeper

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/metoken"
	lerrors "github.com/umee-network/umee/v5/x/metoken/errors"
)

// ClaimLeverageInterest sends accrued interest from x/leverage module to x/metoken account.
func (k Keeper) ClaimLeverageInterest() error {
	if k.ctx.BlockTime().Before(k.getNextInterestClaimTime()) {
		return nil
	}

	leverageLiquidity, err := k.leverageKeeper.GetAllSupplied(*k.ctx, ModuleAddr())
	if err != nil {
		return err
	}

	indexes := k.GetAllRegisteredIndexes()
	for _, index := range indexes {
		balances, err := k.IndexBalances(index.Denom)
		if err != nil {
			return err
		}

		updatedBalances := make([]metoken.AssetBalance, 0)
		for _, balance := range balances.AssetBalances {
			if balance.Leveraged.IsPositive() {
				found, liquidity := leverageLiquidity.Find(balance.Denom)
				if !found {
					continue
				}

				// If there is more liquidity in x/leverage than expected, we claim the delta,
				// which is the accrued interest
				if liquidity.Amount.GT(balance.Leveraged) {
					accruedInterest := sdk.NewCoin(balance.Denom, liquidity.Amount.Sub(balance.Leveraged))
					tokensWithdrawn, err := k.withdrawFromLeverage(accruedInterest)
					if err != nil {
						var leverageError *lerrors.LeverageError
						if errors.As(err, &leverageError) && leverageError.IsRecoverable {
							k.Logger().Debug(
								"claiming interest: couldn't withdraw from leverage",
								"error", err.Error(),
								"block_time", k.ctx.BlockTime(),
							)
							continue
						}

						return err
					}

					balance.Interest = balance.Interest.Add(tokensWithdrawn.Amount)
					updatedBalances = append(updatedBalances, balance)
				}
			}
		}

		if err = k.updateBalances(balances, updatedBalances); err != nil {
			return err
		}
	}

	k.setNextInterestClaimTime(k.ctx.BlockTime().Add(time.Duration(k.GetParams().ClaimingFrequency) * time.Second))

	return nil
}
