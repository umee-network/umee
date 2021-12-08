package types

// Event types and attributes for the leverage module
const (
	EventTypeLoanAsset            = "loan_asset"
	EventTypeWithdrawLoanedAsset  = "withdraw_loaned_asset"
	EventTypeSetCollateralSetting = "set_collateral_setting"
	EventTypeBorrowAsset          = "borrow_asset"
	EventTypeRepayBorrowedAsset   = "repay_borrowed_asset"
	EventTypeLiquidate            = "liquidate_borrow_position"

	EventAttrModule     = ModuleName
	EventAttrLender     = "lender"
	EventAttrBorrower   = "borrower"
	EventAttrLiquidator = "liquidator"
	EventAttrDenom      = "denom"
	EventAttrEnable     = "enabled"
	EventAttrAttempted  = "attempted"
	EventAttrReward     = "reward"
)
