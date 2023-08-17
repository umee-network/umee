package keeper

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/metoken"
	lerrors "github.com/umee-network/umee/v6/x/metoken/errors"
)

// RebalanceReserves checks if the portion of reserves is below the desired and transfer the missing amount from
// x/leverage to x/metoken reserves, or vice versa.
func (k Keeper) RebalanceReserves() error {
	if k.ctx.BlockTime().Before(k.getNextRebalancingTime()) {
		return nil
	}

	indexes := k.GetAllRegisteredIndexes()
	for _, index := range indexes {
		balances, err := k.IndexBalances(index.Denom)
		if err != nil {
			return err
		}

		// if no meToken were minted, there is nothing to check
		if !balances.MetokenSupply.IsPositive() {
			continue
		}

		updatedBalances := make([]metoken.AssetBalance, 0)
		for _, balance := range balances.AssetBalances {
			if balance.AvailableSupply().IsPositive() {
				assetSettings, i := index.AcceptedAsset(balance.Denom)
				if i < 0 {
					k.Logger().Debug(
						"rebalancing reserves: failed getting accepted asset",
						"asset", balance.Denom,
						"index", index.Denom,
						"block_time", k.ctx.BlockTime(),
					)
					continue
				}

				// Calculate the desired reserves amount
				desiredReserves := assetSettings.ReservePortion.MulInt(balance.AvailableSupply()).TruncateInt()
				// In case the x/metoken module has fewer reserves than required,
				// transfer the missing amount from x/leverage to x/metoken
				if desiredReserves.GT(balance.Reserved) {
					missingReserves := desiredReserves.Sub(balance.Reserved)
					tokensWithdrawn, err := k.withdrawFromLeverage(sdk.NewCoin(balance.Denom, missingReserves))
					if err != nil {
						var leverageError *lerrors.LeverageError
						if errors.As(err, &leverageError) && leverageError.IsRecoverable {
							k.Logger().Debug(
								"rebalancing reserves: couldn't withdraw from leverage",
								"error", err.Error(),
								"index", index.Denom,
								"block_time", k.ctx.BlockTime(),
							)
							continue
						}

						return err
					}

					balance.Reserved = balance.Reserved.Add(tokensWithdrawn.Amount)
					balance.Leveraged = balance.Leveraged.Sub(tokensWithdrawn.Amount)
					updatedBalances = append(updatedBalances, balance)

				} else if desiredReserves.LT(balance.Reserved) {
					// In case the x/metoken module has more reserves than required,
					// transfer the extra amount to x/leverage
					extraReserves := balance.Reserved.Sub(desiredReserves)
					tokenSupplied, err := k.supplyToLeverage(sdk.NewCoin(balance.Denom, extraReserves))
					if err != nil {
						var leverageError *lerrors.LeverageError
						if errors.As(err, &leverageError) && leverageError.IsRecoverable {
							k.Logger().Debug(
								"rebalancing reserves: couldn't supply to leverage",
								"error", err.Error(),
								"index", index.Denom,
								"block_time", k.ctx.BlockTime(),
							)
							continue
						}

						return err
					}

					balance.Reserved = balance.Reserved.Sub(tokenSupplied)
					balance.Leveraged = balance.Leveraged.Add(tokenSupplied)
					updatedBalances = append(updatedBalances, balance)
				}
			}
		}

		if err = k.updateBalances(balances, updatedBalances); err != nil {
			return err
		}
	}

	k.setNextRebalancingTime(k.ctx.BlockTime().Add(time.Duration(k.GetParams().RebalancingFrequency) * time.Second))

	return nil
}
