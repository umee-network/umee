package checkers

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PositiveCoins checks if all coins are valida and amount is positive.
func PositiveCoins(note string, coins ...sdk.Coin) []error {
	var errs []error
	for i := range coins {
		if err := coins[i].Validate(); err != nil {
			errs = append(errs, fmt.Errorf("%s coin[%d]: %w", note, i, err))
		} else if !coins[i].Amount.IsPositive() {
			errs = append(errs, fmt.Errorf("%s coin[%d] amount must be positive", note, i))
		}
	}
	return errs
}
