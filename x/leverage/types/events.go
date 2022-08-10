package types

// Event types and attributes for the leverage module
const (
	// Messages
	EventTypeSupply             = "supply"
	EventTypeWithdraw           = "withdraw"
	EventTypeCollateralize      = "collateralize"
	EventTypeDecollateralize    = "decollateralize"
	EventTypeBorrow             = "borrow"
	EventTypeRepayBorrowedAsset = "repay"
	EventTypeLiquidate          = "liquidate"

	// EndBlock
	EventTypeRepayBadDebt      = "repay_bad_debt"
	EventTypeReservesExhausted = "reserves_exhausted"
	EventTypeInterestAccrual   = "interest_accrual"
	EventTypeFundOracle        = "fund_oracle"

	EventAttrModule      = ModuleName
	EventAttrSupplier    = "supplier"
	EventAttrBorrower    = "borrower"
	EventAttrLiquidator  = "liquidator"
	EventAttrDenom       = "denom"
	EventAttrSupplied    = "supplied"
	EventAttrReceived    = "received"
	EventAttrRedeemed    = "redeemed"
	EventAttrAttempted   = "attempted"
	EventAttrRepaid      = "repaid"
	EventAttrLiquidated  = "liquidated"
	EventAttrReward      = "reward"
	EventAttrInterest    = "total_interest"
	EventAttrBlockHeight = "block_height"
	EventAttrUnixTime    = "unix_time"
	EventAttrReserved    = "reserved"
)
