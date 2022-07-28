package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/umee-network/umee/v2/math/bpmath"
)

const (
	// UTokenPrefix defines the uToken denomination prefix for all uToken types.
	UTokenPrefix       = "u/"
	maxBorrowRateLimit = bpmath.ONE * 10_000
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
		// prevent base asset denoms that start with "u/"
		return sdkerrors.Wrap(ErrInvalidAsset, t.BaseDenom)
	}

	if err := sdk.ValidateDenom(t.SymbolDenom); err != nil {
		return err
	}
	if strings.HasPrefix(t.SymbolDenom, UTokenPrefix) {
		// prevent symbol (ticker) denoms that start with "u/"
		return sdkerrors.Wrap(ErrInvalidAsset, t.SymbolDenom)
	}

	// Reserve factor and collateral weight range between 0 and 1, inclusive.
	if t.ReserveFactor.IsNegative() || t.ReserveFactor.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid reserve factor: %s", t.ReserveFactor)
	}

	if t.CollateralWeight.IsNegative() || t.CollateralWeight.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid collateral rate: %s", t.CollateralWeight)
	}

	// Liquidation threshold ranges between collateral weight and 1, inclusive.
	if t.LiquidationThreshold.LT(t.CollateralWeight) || t.LiquidationThreshold.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid liquidation threshold: %s", t.LiquidationThreshold)
	}

	// Kink utilization rate ranges between 0 and 1, exclusive. This prevents
	// multiple interest rates being defined at exactly 0% or 100% utilization
	// e.g. kink at 0%, 2% base borrow rate, 4% borrow rate at kink.
	if !t.KinkUtilization.IsPositive() || t.KinkUtilization.GTE(sdk.OneDec()) {
		return fmt.Errorf("invalid kink utilization rate: %s", t.KinkUtilization)
	}

	// interest rates are non-negative; they do not need to have a maximum value
	if t.BaseBorrowRate.IsNegative() {
		return fmt.Errorf("invalid base borrow rate: %s", t.BaseBorrowRate)
	}
	if t.KinkBorrowRate.IsNegative() {
		return fmt.Errorf("invalid kink borrow rate: %s", t.KinkBorrowRate)
	}
	if t.MaxBorrowRate > maxBorrowRateLimit {
		return fmt.Errorf("invalid max borrow rate: %d - must be in [0; %d] (value in basis points)", t.LiquidationIncentive, maxBorrowRateLimit)
	}

	// Liquidation incentive is non-negative
	if t.LiquidationIncentive > bpmath.ONE {
		return fmt.Errorf("invalid liquidation incentive: %d - must be in [0; 10000]", t.LiquidationIncentive)
	}

	// Blacklisted assets cannot have borrow or supply enabled
	if t.Blacklist {
		if t.EnableMsgBorrow {
			return fmt.Errorf("blacklisted assets cannot have borrowing enabled")
		}
		if t.EnableMsgSupply {
			return fmt.Errorf("blacklisted assets cannot have supplying enabled")
		}
	}

	if t.MaxCollateralShare.IsNegative() || t.MaxCollateralShare.GT(sdk.OneDec()) {
		return sdkerrors.ErrInvalidRequest.Wrap("Token.MaxCollateralShare must be between 0 and 1")
	}

	if t.MaxSupplyUtilization.IsNegative() || t.MaxSupplyUtilization.GT(sdk.OneDec()) {
		return sdkerrors.ErrInvalidRequest.Wrap("Token.MaxSupplyUtilization must be between 0 and 1")
	}

	if t.MinCollateralLiquidity.IsNegative() || t.MaxSupplyUtilization.GT(sdk.OneDec()) {
		return sdkerrors.ErrInvalidRequest.Wrap("Token.MinCollateralLiquidity be between 0 and 1")
	}

	return nil
}

// AssertSupplyEnabled returns an error if a token does not exist or cannot be supplied.
func (t Token) AssertSupplyEnabled() error {
	if !t.EnableMsgSupply {
		return sdkerrors.Wrap(ErrSupplyNotAllowed, t.BaseDenom)
	}
	return nil
}

// AssertBorrowEnabled returns an error if a token does not exist or cannot be borrowed.
func (t Token) AssertBorrowEnabled() error {
	if !t.EnableMsgBorrow {
		return sdkerrors.Wrap(ErrBorrowNotAllowed, t.BaseDenom)
	}
	return nil
}
