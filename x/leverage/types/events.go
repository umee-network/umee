package types

// Event types and attributes for the leverage module
const (
	EventTypeLoanAsset           = "loan_asset"
	EventTypeWithdrawLoanedAsset = "withdraw_loaned_asset"

	EventAttrModule = ModuleName
	EventAttrLender = "lender"
)
