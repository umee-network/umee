package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/umee-network/umee/v5/x/metoken"
)

// ClaimInterest sends accrued interest from x/leverage module to x/metoken account.
func (k Keeper) ClaimInterest() error {
	interestClaimTime, err := k.getNextInterestClaimTime()
	if err != nil {
		return err
	}

	if k.ctx.BlockTime().After(interestClaimTime) {
		indexes := k.GetAllRegisteredIndexes()
		if len(indexes) == 0 {
			return nil
		}

		leverageLiquidity, err := k.leverageKeeper.GetAllSupplied(
			*k.ctx,
			authtypes.NewModuleAddress(metoken.ModuleName),
		)
		if err != nil {
			return err
		}

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
	}

	return nil
}
