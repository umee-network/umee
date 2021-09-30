package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gopkg.in/yaml.v3"
)

// String implements the Stringer interface.
func (p UpdateAssetsProposal) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate performs validation on an Asset type returning an error if the Asset
// is invalid.
func (a Asset) Validate() error {
	if err := sdk.ValidateDenom(a.BaseTokenDenom); err != nil {
		return err
	}

	// TODO: Evaluate if we need additional constraints on the exchange rate.
	if a.ExchangeRate.IsNegative() || a.ExchangeRate.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid exchange rate: %s", a.ExchangeRate)
	}

	// TODO: Evaluate if we need additional constraints on the collateral rate.
	if a.CollateralWeight.IsNegative() || a.CollateralWeight.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid collateral rate: %s", a.CollateralWeight)
	}

	// TODO: Evaluate if we need additional constraints on the base borrow rate.
	if a.BaseBorrowRate.IsNegative() || a.BaseBorrowRate.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid base borrow rate: %s", a.BaseBorrowRate)
	}

	return nil
}
