package message

import (
	lvtypes "github.com/umee-network/umee/v2/x/leverage/types"
)

// AssignedMsg defines the msg to be called.
type AssignedMsg uint16

const (
	// AssignedMsgLend represents the call to lend coins to the capital facility.
	AssignedMsgLend AssignedMsg = iota + 1
	// AssignedMsgWithdraw represents the call to withdraw previously loaned coins
	// from the capital facility.
	AssignedMsgWithdraw
	// AssignedMsgBorrowAsset represents the call to borrowing coins from the
	// capital facility.
	AssignedMsgBorrowAsset
	// AssignedMsgRepayAsset represents the call to repaying borrowed coins to
	// the capital facility.
	AssignedMsgRepayAsset
	// AssignedMsgLiquidate represents the call to repaying a different user's
	// borrowed coins to the capital facility in exchange for some of their
	// collateral.
	AssignedMsgLiquidate
)

// UmeeMsg wraps all the queries availables for cosmwasm smartcontracts.
type UmeeMsg struct {
	// Mandatory field to determine which msg to call.
	AssignedMsg AssignedMsg `json:"assigned_msg"`
	// Used to lending coins to the capital facility.
	LendAsset *lvtypes.MsgLendAsset `json:"lend_asset,omitempty"`
	// Used to withdraw previously loaned coins from the capital facility.
	WithdrawAsset *lvtypes.MsgWithdrawAsset `json:"withdraw_asset,omitempty"`
	// Used to borrowing coins from the capital facility.
	BorrowAsset *lvtypes.MsgBorrowAsset `json:"borrow_asset,omitempty"`
	// Used to repaying borrowed coins to the capital facility.
	RepayAsset *lvtypes.MsgRepayAsset `json:"repay_asset,omitempty"`
	// Used to repaying a different user's borrowed coins
	// to the capital facility in exchange for some of their collateral.
	Liquidate *lvtypes.MsgLiquidate `json:"liquidate,omitempty"`
}
