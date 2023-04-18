package msg

import (
	lvtypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// AssignedMsg defines the msg to be called.
type AssignedMsg uint16

const (
	// AssignedMsgSupply represents the call to supply coins to the capital facility.
	AssignedMsgSupply AssignedMsg = iota + 0
	// AssignedMsgWithdraw represents the call to withdraw previously loaned coins
	// from the capital facility.
	AssignedMsgWithdraw
	// AssignedMsgCollateralize represents the call to enable an amount of
	// selected uTokens as collateral.
	AssignedMsgCollateralize
	// AssignedMsgDecollateralize represents the call to disable amount of
	// an selected uTokens as collateral.
	AssignedMsgDecollateralize
	// AssignedMsgBorrow represents the call to borrowing coins from the
	// capital facility.
	AssignedMsgBorrow
	// AssignedMsgRepay represents the call to repaying borrowed coins to
	// the capital facility.
	AssignedMsgRepay
	// AssignedMsgLiquidate represents the call to repaying a different user's
	// borrowed coins to the capital facility in exchange for some of their
	// collateral.
	AssignedMsgLiquidate
	// AssignedMsgSupplyCollateral represents the call to supply and collateralize their assets.
	AssignedMsgSupplyCollateral
	AssignedMsgMaxWithdraw
)

// UmeeMsg wraps all the messages availables for cosmwasm smartcontracts.
type UmeeMsg struct {
	// Mandatory field to determine which msg to call.
	AssignedMsg AssignedMsg `json:"assigned_msg"`
	// Used to supply coins to the capital facility.
	Supply *lvtypes.MsgSupply `json:"supply,omitempty"`
	// Used to withdraw previously loaned coins from the capital facility.
	Withdraw *lvtypes.MsgWithdraw `json:"withdraw,omitempty"`
	// Used to enable an amount of selected uTokens as collateral.
	Collateralize *lvtypes.MsgCollateralize `json:"collateralize,omitempty"`
	// Used to disable amount of an selected uTokens as collateral.
	Decollateralize *lvtypes.MsgDecollateralize `json:"decollateralize,omitempty"`
	// Used to borrowing coins from the capital facility.
	Borrow *lvtypes.MsgBorrow `json:"borrow,omitempty"`
	// Used to repaying borrowed coins to the capital facility.
	Repay *lvtypes.MsgRepay `json:"repay,omitempty"`
	// Used to repaying a different user's borrowed coins
	// to the capital facility in exchange for some of their collateral.
	Liquidate *lvtypes.MsgLiquidate `json:"liquidate,omitempty"`
	// Used to do supply and collateralize their assets.
	SupplyCollateral *lvtypes.MsgSupplyCollateral `json:"supply_collateralize,omitempty"`
	// Used to do withdraw maximum assets by supplier.
	AssignedMsgMaxWithdraw *lvtypes.MsgMaxWithdraw `json:"max_withdraw,omitempty"`
}
