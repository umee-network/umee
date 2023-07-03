package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	appparams "github.com/umee-network/umee/v5/app/params"
)

const (
	// UTokenPrefix defines the uToken denomination prefix for all uToken types.
	UTokenPrefix = "u/"
)

var halfDec = sdk.MustNewDecFromStr("0.5")

// HasUTokenPrefix detects the uToken prefix on a denom.
func HasUTokenPrefix(denom string) bool {
	return strings.HasPrefix(denom, UTokenPrefix)
}

// ToUTokenDenom adds the uToken prefix to a denom. Returns an empty string
// instead if the prefix was already present.
func ToUTokenDenom(denom string) string {
	if HasUTokenPrefix(denom) {
		return ""
	}
	return UTokenPrefix + denom
}

// ValidateBaseDenom validates a denom and ensures it is not a uToken.
func ValidateBaseDenom(denom string) error {
	if err := sdk.ValidateDenom(denom); err != nil {
		return err
	}
	if HasUTokenPrefix(denom) {
		return ErrUToken.Wrap(denom)
	}
	return nil
}

// ToTokenDenom strips the uToken prefix from a denom, or returns an empty
// string if it was not present. Also returns an empty string if the prefix
// was repeated multiple times.
func ToTokenDenom(denom string) string {
	if !HasUTokenPrefix(denom) {
		return ""
	}
	s := strings.TrimPrefix(denom, UTokenPrefix)
	if HasUTokenPrefix(s) {
		// denom started with "u/u/"
		return ""
	}
	return s
}

// Validate performs validation on an Token type returning an error if the Token
// is invalid.
func (t Token) Validate() error {
	if err := sdk.ValidateDenom(t.BaseDenom); err != nil {
		return err
	}
	if HasUTokenPrefix(t.BaseDenom) {
		// prevent base asset denoms that start with "u/"
		return ErrUToken.Wrap(t.BaseDenom)
	}

	if err := sdk.ValidateDenom(t.SymbolDenom); err != nil {
		return err
	}
	if HasUTokenPrefix(t.SymbolDenom) {
		// prevent symbol denoms that start with "u/"
		return ErrUToken.Wrap(t.SymbolDenom)
	}

	one := sdk.OneDec()

	// Reserve factor is non-negative and less than 1.
	if t.ReserveFactor.IsNegative() || t.ReserveFactor.GTE(one) {
		return fmt.Errorf("invalid reserve factor: %s", t.ReserveFactor)
	}
	// Collateral weight is non-negative and less than 1.
	if t.CollateralWeight.IsNegative() || t.CollateralWeight.GTE(one) {
		return fmt.Errorf("invalid collateral rate: %s", t.CollateralWeight)
	}
	if t.LiquidationThreshold.LT(t.CollateralWeight) || t.LiquidationThreshold.GTE(one) {
		return fmt.Errorf(
			"liquidation threshold must be bigger or equal than collateral weight, got: %s",
			t.LiquidationThreshold,
		)
	}

	// Kink utilization rate ranges between 0 and 1, exclusive. This prevents
	// multiple interest rates being defined at exactly 0% or 100% utilization
	// e.g. kink at 0%, 2% base borrow rate, 4% borrow rate at kink.
	if !t.KinkUtilization.IsPositive() || t.KinkUtilization.GTE(one) {
		return fmt.Errorf("invalid kink utilization rate: %s", t.KinkUtilization)
	}

	// interest rates are non-negative; they do not need to have a maximum value
	if t.BaseBorrowRate.IsNegative() {
		return fmt.Errorf("invalid base borrow rate: %s", t.BaseBorrowRate)
	}
	if t.KinkBorrowRate.IsNegative() {
		return fmt.Errorf("invalid kink borrow rate: %s", t.KinkBorrowRate)
	}
	if t.MaxBorrowRate.IsNegative() {
		return fmt.Errorf("invalid max borrow rate: %s", t.MaxBorrowRate)
	}

	// Liquidation incentive is non-negative
	if t.LiquidationIncentive.IsNegative() {
		return fmt.Errorf("invalid liquidation incentive: %s", t.LiquidationIncentive)
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

	if t.MaxCollateralShare.IsNegative() || t.MaxCollateralShare.GT(one) {
		return sdkerrors.ErrInvalidRequest.Wrap("Token.MaxCollateralShare must be between 0 and 1")
	}

	if t.MaxSupplyUtilization.IsNegative() || t.MaxSupplyUtilization.GT(one) {
		return sdkerrors.ErrInvalidRequest.Wrap("Token.MaxSupplyUtilization must be between 0 and 1")
	}

	if t.MinCollateralLiquidity.IsNegative() || t.MinCollateralLiquidity.GT(one) {
		return sdkerrors.ErrInvalidRequest.Wrap("Token.MinCollateralLiquidity be between 0 and 1")
	}

	if t.MaxSupply.IsNegative() {
		return sdkerrors.ErrInvalidRequest.Wrap("Token.MaxSupply must not be negative")
	}

	return nil
}

// AssertSupplyEnabled returns an error if a Token cannot be supplied.
func (t Token) AssertSupplyEnabled() error {
	if !t.EnableMsgSupply {
		return ErrSupplyNotAllowed.Wrap(t.BaseDenom)
	}
	return nil
}

// AssertBorrowEnabled returns an error if a Token cannot be borrowed.
func (t Token) AssertBorrowEnabled() error {
	if !t.EnableMsgBorrow {
		return ErrBorrowNotAllowed.Wrap(t.BaseDenom)
	}
	return nil
}

// AssertNotBlacklisted returns an error if a Token is blacklisted.
func (t Token) AssertNotBlacklisted() error {
	if t.Blacklist {
		return ErrBlacklisted.Wrap(t.BaseDenom)
	}
	return nil
}

// BorrowFactor returns the minimum of 2.0 or 1 / collateralWeight.
func (t Token) BorrowFactor() sdk.Dec {
	if t.CollateralWeight.LTE(halfDec) {
		return sdk.MustNewDecFromStr("2.0")
	}
	return sdk.OneDec().Quo(t.CollateralWeight)
}

func defaultUmeeToken() Token {
	return Token{
		BaseDenom:       appparams.BondDenom,
		SymbolDenom:     "UMEE",
		Exponent:        6,
		EnableMsgSupply: true,
		EnableMsgBorrow: true,
		Blacklist:       false,
		// Reserves
		ReserveFactor: sdk.MustNewDecFromStr("0.10"),
		// Interest rate model
		BaseBorrowRate:  sdk.MustNewDecFromStr("0.05"),
		KinkBorrowRate:  sdk.MustNewDecFromStr("0.10"),
		MaxBorrowRate:   sdk.MustNewDecFromStr("0.80"),
		KinkUtilization: sdk.MustNewDecFromStr("0.50"),
		// Collateral
		CollateralWeight:     sdk.MustNewDecFromStr("0.35"),
		LiquidationThreshold: sdk.MustNewDecFromStr("0.50"),
		// Liquidation
		LiquidationIncentive: sdk.MustNewDecFromStr("0.10"),
		// Market limits
		MaxCollateralShare:     sdk.MustNewDecFromStr("1.00"),
		MaxSupplyUtilization:   sdk.MustNewDecFromStr("0.90"),
		MinCollateralLiquidity: sdk.MustNewDecFromStr("0.3"),
		MaxSupply:              sdk.NewInt(1000_000000_000000),
	}
}

func DefaultRegistry() []Token {
	return []Token{
		defaultUmeeToken(),
	}
}
