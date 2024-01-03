package quota

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/x/uibc"
)

// returns mapping of token denom (according to x/bank) to token symbol name.
func (k Keeper) coinsWithTokenSymbols(ctx sdk.Context, coins sdk.DecCoins) []uibc.DecCoinSymbol {
	tokenSymbols := map[string]string{}
	for _, t := range k.leverage.GetAllRegisteredTokens(ctx) {
		tokenSymbols[t.BaseDenom] = t.SymbolDenom
	}
	cs := make([]uibc.DecCoinSymbol, len(coins))
	for i := range coins {
		cs[i].Denom = coins[i].Denom
		cs[i].Amount = coins[i].Amount
		cs[i].Symbol = tokenSymbols[coins[i].Denom]
	}
	return cs
}
