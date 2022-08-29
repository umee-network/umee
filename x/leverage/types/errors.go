package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// 1XX = General Validation
	ErrEmptyAddress = sdkerrors.Register(ModuleName, 100, "empty address")
	ErrNilAsset     = sdkerrors.Register(ModuleName, 101, "nil asset")

	// 2XX = Token Registry
	ErrNotRegisteredToken   = sdkerrors.Register(ModuleName, 200, "not a registered Token")
	ErrNotRegisteredUToken  = sdkerrors.Register(ModuleName, 201, "not a registered uToken")
	ErrUToken               = sdkerrors.Register(ModuleName, 202, "denom should not be a uToken")
	ErrNotUToken            = sdkerrors.Register(ModuleName, 203, "denom should be a uToken")
	ErrSupplyNotAllowed     = sdkerrors.Register(ModuleName, 204, "supplying of asset disabled")
	ErrBorrowNotAllowed     = sdkerrors.Register(ModuleName, 205, "borrowing of asset disabled")
	ErrBlacklisted          = sdkerrors.Register(ModuleName, 206, "base denom blacklisted")
	ErrCollateralWeightZero = sdkerrors.Register(ModuleName, 207, "token collateral weight is zero")

	// 3XX = User Positions
	ErrInsufficientBalance    = sdkerrors.Register(ModuleName, 300, "insufficient balance")
	ErrInsufficientCollateral = sdkerrors.Register(ModuleName, 301, "insufficient collateral")
	ErrDenomNotBorrowed       = sdkerrors.Register(ModuleName, 302, "denom not borrowed")
	ErrLiquidationRepayZero   = sdkerrors.Register(ModuleName, 303, "liquidation would repay zero tokens")

	// 4XX = Price Sensitive
	ErrBadValue              = sdkerrors.Register(ModuleName, 400, "bad USD value")
	ErrInvalidOraclePrice    = sdkerrors.Register(ModuleName, 401, "invalid oracle price")
	ErrUndercollaterized     = sdkerrors.Register(ModuleName, 402, "borrow positions are undercollaterized")
	ErrLiquidationIneligible = sdkerrors.Register(ModuleName, 403, "borrower not eligible for liquidation")

	// 5XX = Market Conditions
	ErrLendingPoolInsufficient = sdkerrors.Register(ModuleName, 500, "lending pool insufficient")
	ErrMaxSupplyUtilization    = sdkerrors.Register(ModuleName, 501, "market would exceed MaxSupplyUtilization")
	ErrMinCollateralLiquidity  = sdkerrors.Register(ModuleName, 502, "market would fall below MinCollateralLiquidity")
	ErrMaxCollateralShare      = sdkerrors.Register(ModuleName, 503, "market would exceed MaxCollateralShare")

	// 6XX = Internal Failsafes
	ErrInvalidUtilization      = sdkerrors.Register(ModuleName, 600, "invalid token utilization")
	ErrNegativeTotalBorrowed   = sdkerrors.Register(ModuleName, 601, "total borrowed was negative")
	ErrNegativeAPY             = sdkerrors.Register(ModuleName, 602, "negative APY")
	ErrNegativeTimeElapsed     = sdkerrors.Register(ModuleName, 603, "negative time elapsed since last interest time")
	ErrInvalidExchangeRate     = sdkerrors.Register(ModuleName, 604, "exchange rate less than one")
	ErrInconsistentTotalBorrow = sdkerrors.Register(ModuleName, 605, "total adjusted borrow inconsistency")
	ErrInvalidInteresrScalar   = sdkerrors.Register(ModuleName, 606, "interest scalar less than one")
)
