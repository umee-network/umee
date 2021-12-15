package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// UTokenPrefix defines the uToken denomination prefix for all uToken types.
	UTokenPrefix = "u/"
)

// UTokenFromTokenDenom returns the uToken denom given a token denom.
func UTokenFromTokenDenom(tokenDenom string) string {
	return UTokenPrefix + tokenDenom
}

// Validate performs validation on an Token type returning an error if the Token
// is invalid.
func (t Token) Validate() error {
	if err := sdk.ValidateDenom(t.BaseDenom); err != nil {
		return err
	}
	if strings.HasPrefix(t.BaseDenom, UTokenPrefix) {
		// Prevents base asset denoms that start with "u/"
		return sdkerrors.Wrap(ErrInvalidAsset, t.BaseDenom)
	}

	// Reserve factor and collateral weight range between 0 and 1, inclusive.
	if t.ReserveFactor.IsNegative() || t.ReserveFactor.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid reserve factor: %s", t.ReserveFactor)
	}
	if t.CollateralWeight.IsNegative() || t.CollateralWeight.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid collateral rate: %s", t.CollateralWeight)
	}

	// Kink utilization rate ranges between 0 and 1, exclusive. This prevents multiple interest rates being
	// defined at exactly 0% or 100% utilization (e.g. kink at 0%, 2% base borrow rate, 4% borrow rate at kink.)
	if !t.KinkUtilizationRate.IsPositive() || t.KinkUtilizationRate.GTE(sdk.OneDec()) {
		return fmt.Errorf("invalid kink utilization rate: %s", t.KinkUtilizationRate)
	}

	// Interest rates are non-negative. They do not need to have a maximum value.
	if t.BaseBorrowRate.IsNegative() {
		return fmt.Errorf("invalid base borrow rate: %s", t.BaseBorrowRate)
	}
	if t.KinkBorrowRate.IsNegative() {
		return fmt.Errorf("invalid kink borrow rate: %s", t.KinkBorrowRate)
	}
	if t.MaxBorrowRate.IsNegative() {
		return fmt.Errorf("invalid max borrow rate: %s", t.MaxBorrowRate)
	}

	// Liquidation incentives are non-negative
	if t.LiquidationIncentive.IsNegative() {
		return fmt.Errorf("invalid liquidation incentive: %s", t.LiquidationIncentive)
	}

	return nil
}
