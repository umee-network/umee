package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	appparams "github.com/umee-network/umee/v3/app/params"
)

const (
	// UTokenPrefix defines the uToken denomination prefix for all uToken types.
	UTokenPrefix                  = "u/"
	DefaultExponent               = uint32(6)
	DefaultEnableMsgSupply        = true
	DefaultEnableMsgBorrow        = true
	DefaultBlacklist              = false
	UmeeDenom              string = appparams.BondDenom
	UmeeSymbol             string = "umee"
	AtomDenom              string = "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9"
	AtomSymbol             string = "atom"
)

var (
	DefaultReserveFactor          = sdk.MustNewDecFromStr("0.100000000000000000")
	DefaultCollateralWeight       = sdk.MustNewDecFromStr("0.050000000000000000")
	DefaultLiquidationThreshold   = sdk.MustNewDecFromStr("0.050000000000000000")
	DefaultBaseBorrowRate         = sdk.MustNewDecFromStr("0.020000000000000000")
	DefaultKinkBorrowRate         = sdk.MustNewDecFromStr("0.200000000000000000")
	DefaultMaxBorrowRate          = sdk.MustNewDecFromStr("1.50000000000000000")
	DefaultKinkUtilization        = sdk.MustNewDecFromStr("0.200000000000000000")
	DefaultLiquidationIncentive   = sdk.MustNewDecFromStr("0.180000000000000000")
	DefaultMaxCollateralShare     = sdk.MustNewDecFromStr("1.00000000000000000")
	DefaultMaxSupplyUtilization   = sdk.MustNewDecFromStr("1.00000000000000000")
	DefaultMinCollateralLiquidity = sdk.MustNewDecFromStr("0.000000000000000000")
	DefaultMaxSupply              = sdk.NewInt(100_000000_000000)
)

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

	// Reserve factor is non-negative and less than 1.
	if t.ReserveFactor.IsNegative() || t.ReserveFactor.GTE(sdk.OneDec()) {
		return fmt.Errorf("invalid reserve factor: %s", t.ReserveFactor)
	}
	// Collateral weight is non-negative and less than 1.
	if t.CollateralWeight.IsNegative() || t.CollateralWeight.GTE(sdk.OneDec()) {
		return fmt.Errorf("invalid collateral rate: %s", t.CollateralWeight)
	}
	// Liquidation threshold is at least collateral weight, but less than 1.
	if t.LiquidationThreshold.LT(t.CollateralWeight) || t.LiquidationThreshold.GTE(sdk.OneDec()) {
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

	if t.MaxCollateralShare.IsNegative() || t.MaxCollateralShare.GT(sdk.OneDec()) {
		return sdkerrors.ErrInvalidRequest.Wrap("Token.MaxCollateralShare must be between 0 and 1")
	}

	if t.MaxSupplyUtilization.IsNegative() || t.MaxSupplyUtilization.GT(sdk.OneDec()) {
		return sdkerrors.ErrInvalidRequest.Wrap("Token.MaxSupplyUtilization must be between 0 and 1")
	}

	if t.MinCollateralLiquidity.IsNegative() || t.MaxSupplyUtilization.GT(sdk.OneDec()) {
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
		return sdkerrors.Wrap(ErrSupplyNotAllowed, t.BaseDenom)
	}
	return nil
}

// AssertBorrowEnabled returns an error if a Token cannot be borrowed.
func (t Token) AssertBorrowEnabled() error {
	if !t.EnableMsgBorrow {
		return sdkerrors.Wrap(ErrBorrowNotAllowed, t.BaseDenom)
	}
	return nil
}

// AssertNotBlacklisted returns an error if a Token is blacklisted.
func (t Token) AssertNotBlacklisted() error {
	if t.Blacklist {
		return sdkerrors.Wrap(ErrBlacklisted, t.BaseDenom)
	}
	return nil
}

func NewDefaultRegistryToken(baseDenom, symbolDenom string) Token {
	return Token{
		BaseDenom:              baseDenom,
		SymbolDenom:            symbolDenom,
		Exponent:               DefaultExponent,
		ReserveFactor:          DefaultReserveFactor,
		CollateralWeight:       DefaultCollateralWeight,
		LiquidationThreshold:   DefaultLiquidationThreshold,
		BaseBorrowRate:         DefaultBaseBorrowRate,
		KinkBorrowRate:         DefaultKinkBorrowRate,
		MaxBorrowRate:          DefaultMaxBorrowRate,
		KinkUtilization:        DefaultKinkUtilization,
		LiquidationIncentive:   DefaultLiquidationIncentive,
		EnableMsgSupply:        DefaultEnableMsgSupply,
		EnableMsgBorrow:        DefaultEnableMsgBorrow,
		Blacklist:              DefaultBlacklist,
		MaxCollateralShare:     DefaultMaxCollateralShare,
		MaxSupplyUtilization:   DefaultMaxSupplyUtilization,
		MinCollateralLiquidity: DefaultMinCollateralLiquidity,
		MaxSupply:              DefaultMaxSupply,
	}
}

func DefaultRegistry() []Token {
	return []Token{
		NewDefaultRegistryToken(
			UmeeDenom,
			UmeeSymbol,
		),
		NewDefaultRegistryToken(
			AtomDenom,
			AtomSymbol,
		),
	}
}
