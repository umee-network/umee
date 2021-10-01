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
	ErrRepayNonexistentBorrow  = sdkerrors.Register(ModuleName, 1104, "repayment of nonexistent borrow")
)
