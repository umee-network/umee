package migrations

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MigrateOldExgRatesToExgRatesWithTimestamp will migrate old exchnage rate of denoms into new exchange rate format with
// timestamp into store.
func (m Migrator) MigrateOldExgRatesToExgRatesWithTimestamp(ctx sdk.Context) {
	m.keeper.IterateOldExchangeRates(ctx, func(s string, d sdk.Dec) bool {
		ctx.Logger().Info("Migrating old exchange rate to new exchange rate format", "denom", s)
		m.keeper.SetExchangeRateWithTimestamp(ctx, s, d, ctx.BlockTime())
		return false
	})
}
