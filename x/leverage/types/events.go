package types

// Event types and attributes for the leverage module
const (
	EventTypeLoanAsset            = "loan_asset"
	EventTypeWithdrawLoanedAsset  = "withdraw_loaned_asset"
	EventTypeSetCollateralSetting = "set_collateral_setting"
	EventTypeBorrowAsset          = "borrow_asset"
	EventTypeRepayBorrowedAsset   = "repay_borrowed_asset"

	EventAttrModule   = ModuleName
	EventAttrLender   = "lender"
	EventAttrBorrower = "borrower"
	EventAttrDenom    = "denom"
	EventAttrEnable   = "enabled"
)
