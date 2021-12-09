package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DenomsFromCoins returns all the denominations in a set of Coins.
func DenomsFromCoins(coins sdk.Coins) []string {
	denoms := make([]string, len(coins))
	for i, c := range coins {
		denoms[i] = c.Denom
	}

	return denoms
}
