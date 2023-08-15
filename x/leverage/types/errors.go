package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

var (
	// 1XX = General Validation
	ErrEmptyAddress     = errors.Register(ModuleName, 100, "empty address")
	ErrNilAsset         = errors.Register(ModuleName, 101, "nil asset")
	ErrGetAmount        = errors.Register(ModuleName, 102, "retrieved invalid amount")
	ErrSetAmount        = errors.Register(ModuleName, 103, "cannot set invalid amount")
	ErrInvalidPriceMode = errors.Register(ModuleName, 104, "invalid price mode")

	// 2XX = Token Registry
	ErrNotRegisteredToken   = errors.Register(ModuleName, 200, "not a registered Token")
	ErrUToken               = errors.Register(ModuleName, 201, "denom should not be a uToken")
	ErrNotUToken            = errors.Register(ModuleName, 202, "denom should be a uToken")
	ErrSupplyNotAllowed     = errors.Register(ModuleName, 203, "supplying of Token disabled")
	ErrBorrowNotAllowed     = errors.Register(ModuleName, 204, "borrowing of Token disabled")
	ErrBlacklisted          = errors.Register(ModuleName, 205, "blacklisted Token")
	ErrCollateralWeightZero = errors.Register(
		ModuleName, 206,
		"collateral weight of Token is zero: can't be used as a collateral",
	)
	ErrDuplicateToken           = errors.Register(ModuleName, 207, "duplicate token")
	ErrEmptyAddAndUpdateTokens  = errors.Register(ModuleName, 208, "empty add and update tokens")
	ErrEmptyUpdateSpecialAssets = errors.Register(ModuleName, 209, "empty update special asset pairs")
	ErrDuplicatePair            = errors.Register(ModuleName, 210, "duplicate special asset pair")
	ErrProposedSetOrder         = errors.Register(ModuleName, 211, "asset sets not in ascending (weight) order")

	// 3XX = User Positions
	ErrInsufficientBalance    = errors.Register(ModuleName, 300, "insufficient balance")
	ErrInsufficientCollateral = errors.Register(ModuleName, 301, "insufficient collateral")
	ErrLiquidationRepayZero   = errors.Register(ModuleName, 303, "liquidation would repay zero tokens")
	ErrBondedCollateral       = errors.Register(ModuleName, 304, "collateral is bonded to incentive module")

	// 4XX = PriceByBaseDenom Sensitive
	ErrBadValue              = errors.Register(ModuleName, 400, "bad USD value")
	ErrInvalidOraclePrice    = errors.Register(ModuleName, 401, "invalid oracle price")
	ErrUndercollaterized     = errors.Register(ModuleName, 402, "borrow positions are undercollaterized")
	ErrLiquidationIneligible = errors.Register(ModuleName, 403, "borrower not eligible for liquidation")
	ErrNoHistoricMedians     = errors.Register(ModuleName, 405, "insufficient historic medians available")

	// 5XX = Market Conditions
	ErrLendingPoolInsufficient = errors.Register(ModuleName, 500, "lending pool insufficient")
	ErrMaxSupplyUtilization    = errors.Register(ModuleName, 501, "market would exceed MaxSupplyUtilization")
	ErrMinCollateralLiquidity  = errors.Register(ModuleName, 502, "market would fall below MinCollateralLiquidity")
	ErrMaxCollateralShare      = errors.Register(ModuleName, 503, "market would exceed MaxCollateralShare")
	ErrMaxSupply               = errors.Register(ModuleName, 504, "market would exceed MaxSupply")

	// 6XX = Internal Failsafes
	ErrInvalidUtilization      = errors.Register(ModuleName, 600, "invalid token utilization")
	ErrNegativeTotalBorrowed   = errors.Register(ModuleName, 601, "total borrowed was negative")
	ErrNegativeAPY             = errors.Register(ModuleName, 602, "negative APY")
	ErrNegativeTimeElapsed     = errors.Register(ModuleName, 603, "negative time elapsed since last interest time")
	ErrInvalidExchangeRate     = errors.Register(ModuleName, 604, "exchange rate less than one")
	ErrInconsistentTotalBorrow = errors.Register(ModuleName, 605, "total adjusted borrow inconsistency")
	ErrExcessiveTimeElapsed    = errors.Register(ModuleName, 606, "excessive time elapsed since last interest time")
	ErrIncentiveKeeperNotSet   = errors.Register(ModuleName, 607, "incentive keeper not set")

	// 7XX = Disabled Functionality
	ErrNotLiquidatorNode = errors.Register(ModuleName, 700, "node has disabled liquidator queries")
)
