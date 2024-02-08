package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/util/coin"
)

var (
	halfDec = sdk.MustNewDecFromStr("0.5")
	one     = sdk.OneDec()
)

// ValidateBaseDenom validates a denom and ensures it is not a uToken.
func ValidateBaseDenom(denom string) error {
	if err := sdk.ValidateDenom(denom); err != nil {
		return err
	}
	if coin.HasUTokenPrefix(denom) {
		return ErrUToken.Wrap(denom)
	}
	return nil
}

// Validate performs validation on an Token type returning an error if the Token
// is invalid.
func (t Token) Validate() error {
	if err := validateBaseDenoms(t.BaseDenom); err != nil {
		return fmt.Errorf("base_denom: %v", err)
	}
	if err := validateBaseDenoms(t.SymbolDenom); err != nil {
		return fmt.Errorf("symbol_denom: %v", err)
	}

	if err := checkers.DecInZeroOne(t.ReserveFactor, "reserve factor", false); err != nil {
		return err
	}
	if err := checkers.DecInZeroOne(t.CollateralWeight, "collateral weight", false); err != nil {
		return err
	}
	if err := checkers.DecInZeroOne(t.ReserveFactor, "reserve factor", false); err != nil {
		return err
	}
	if t.LiquidationThreshold.LT(t.CollateralWeight) || t.LiquidationThreshold.GTE(one) {
		return fmt.Errorf(
			"liquidation threshold must be bigger or equal than collateral weight, got: %s",
			t.LiquidationThreshold,
		)
	}

	// Kink utilization rate ranges between 0 and 1, inclusive.
	if t.KinkUtilization.IsNegative() || t.KinkUtilization.GT(one) {
		return fmt.Errorf("invalid kink utilization rate: %s", t.KinkUtilization)
	}
	// The following rules ensure the utilization:APY graph is continuous
	if t.KinkUtilization.GT(t.MaxSupplyUtilization) {
		return fmt.Errorf("kink utilization (%s) cannot be greater than than max supply utilization (%s)",
			t.KinkUtilization, t.MaxSupplyUtilization)
	}
	if t.KinkUtilization.Equal(t.MaxSupplyUtilization) && !t.MaxBorrowRate.Equal(t.KinkBorrowRate) {
		return fmt.Errorf(
			"since kink utilization equals max supply utilization, kink borrow rate must equal max borrow rate (%s)",
			t.MaxBorrowRate,
		)
	}
	if t.KinkUtilization.IsZero() && !t.KinkBorrowRate.Equal(t.BaseBorrowRate) {
		return fmt.Errorf(
			"since kink utilization equals zero, kink borrow rate must equal base borrow rate (%s)",
			t.BaseBorrowRate,
		)
	}
	if t.MaxSupplyUtilization.IsZero() && !t.MaxBorrowRate.Equal(t.BaseBorrowRate) {
		return fmt.Errorf(
			"since max supply utilization equals zero, max borrow rate must equal base borrow rate (%s)",
			t.BaseBorrowRate,
		)
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

// Validate performs validation on an SpecialAssetPair type
func (p SpecialAssetPair) Validate() error {
	if err := validateBaseDenoms(p.Collateral, p.Borrow); err != nil {
		return err
	}

	if p.CollateralWeight.IsNil() || p.LiquidationThreshold.IsNil() {
		return fmt.Errorf("nil collateral weight or liquidation threshold for asset pair (%s,%s)",
			p.Borrow, p.Collateral)
	}

	// Collateral Weight is non-negative and less than 1.
	if p.CollateralWeight.IsNegative() || p.CollateralWeight.GTE(sdk.OneDec()) {
		return fmt.Errorf("invalid collateral rate: %s", p.CollateralWeight)
	}

	// Liquidation Threshold ranges between collateral weight and 1.
	if p.LiquidationThreshold.LT(p.CollateralWeight) || p.LiquidationThreshold.GTE(sdk.OneDec()) {
		return fmt.Errorf("invalid liquidation threshold: %s", p.LiquidationThreshold)
	}

	return nil
}

// Validate performs validation on an SpecialAssetSet type
func (s SpecialAssetSet) Validate() error {
	if err := validateBaseDenoms(s.Assets...); err != nil {
		return err
	}

	denoms := map[string]bool{}
	for _, a := range s.Assets {
		if _, ok := denoms[a]; ok {
			return fmt.Errorf("duplicate special asset pair: %s", a)
		}
		denoms[a] = true
	}

	if s.CollateralWeight.IsNil() || s.LiquidationThreshold.IsNil() {
		return fmt.Errorf("nil collateral weight or liquidation threshold for asset set %s)", s.Assets)
	}

	// Collateral Weight is non-negative and less than 1.
	if s.CollateralWeight.IsNegative() || s.CollateralWeight.GTE(sdk.OneDec()) {
		return fmt.Errorf("invalid collateral rate: %s", s.CollateralWeight)
	}

	// Liquidation Threshold ranges between collateral weight and 1.
	if s.LiquidationThreshold.LT(s.CollateralWeight) || s.LiquidationThreshold.GTE(sdk.OneDec()) {
		return fmt.Errorf("invalid liquidation threshold: %s", s.LiquidationThreshold)
	}

	return nil
}

// validateBaseDenoms ensures that one or more strings are valid token denoms without the uToken prefix
func validateBaseDenoms(denoms ...string) error {
	for _, s := range denoms {
		if err := ValidateBaseDenom(s); err != nil {
			return err
		}
	}
	return nil
}
