package types

// Event types and attributes for the leverage module
const (
	EventTypeLoanAsset            = "loan_asset"
	EventTypeWithdrawLoanedAsset  = "withdraw_loaned_asset"
	EventTypeSetCollateralSetting = "set_collateral_setting"
	EventTypeBorrowAsset          = "borrow_asset"
	EventTypeRepayBorrowedAsset   = "repay_borrowed_asset"
	EventTypeLiquidate            = "liquidate_borrow_position"
	EventTypeRepayBadDebt         = "repay_bad_debt"
	EventTypeReservesExhausted    = "reserves_exhausted"
	EventTypeInterestAccrual      = "interest_accrual"
	EventTypeFundOracle           = "fund_oracle"

	EventAttrModule      = ModuleName
	EventAttrLender      = "lender"
	EventAttrBorrower    = "borrower"
	EventAttrLiquidator  = "liquidator"
	EventAttrDenom       = "denom"
	EventAttrEnable      = "enabled"
	EventAttrAttempted   = "attempted"
	EventAttrReward      = "reward"
	EventAttrInterest    = "total_interest"
	EventAttrBlockHeight = "block_height"
	EventAttrUnixTime    = "unix_time"
	EventAttrReserved    = "reserved"
)
