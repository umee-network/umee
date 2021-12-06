package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidAsset            = sdkerrors.Register(ModuleName, 1100, "invalid asset")
	ErrInsufficientBalance     = sdkerrors.Register(ModuleName, 1101, "insufficient balance")
	ErrBorrowLimitLow          = sdkerrors.Register(ModuleName, 1102, "borrow limit too low")
	ErrLendingPoolInsufficient = sdkerrors.Register(ModuleName, 1103, "lending pool insufficient")
	ErrInvalidRepayment        = sdkerrors.Register(ModuleName, 1104, "invalid repayment")
	ErrInvalidAddress          = sdkerrors.Register(ModuleName, 1105, "invalid address")
	ErrNegativeTotalBorrowed   = sdkerrors.Register(ModuleName, 1106, "total borrowed was negative")
	ErrInvalidUtilization      = sdkerrors.Register(ModuleName, 1107, "invalid token utilization")
	ErrLiquidationIneligible   = sdkerrors.Register(ModuleName, 1108, "borrower not eligible for liquidation")
	ErrBadValue                = sdkerrors.Register(ModuleName, 1109, "bad USD value")
	ErrLiquidatorBalanceZero   = sdkerrors.Register(ModuleName, 1110, "liquidator base asset balance is zero")
)
