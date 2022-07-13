package types

// Event types and attributes for the leverage module
const (
	EventTypeSupply             = "supply"
	EventTypeWithdraw           = "withdraw"
	EventTypeCollateralize      = "collateralize"
	EventTypeDecollateralize    = "decollateralize"
	EventTypeBorrow             = "borrow"
	EventTypeRepayBorrowedAsset = "repay"
	EventTypeLiquidate          = "liquidate"
	EventTypeRepayBadDebt       = "repay_bad_debt"
	EventTypeReservesExhausted  = "reserves_exhausted"
	EventTypeInterestAccrual    = "interest_accrual"
	EventTypeFundOracle         = "fund_oracle"

	EventAttrModule      = ModuleName
	EventAttrSupplier    = "supplier"
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
