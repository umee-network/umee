package msg

import (
	lvtypes "github.com/umee-network/umee/v5/x/leverage/types"
)

// UmeeMsg wraps all the messages availables for cosmwasm smartcontracts.
type UmeeMsg struct {
	// Used to supply coins to the capital facility.
	Supply *lvtypes.MsgSupply `json:"supply,omitempty"`
	// Used to withdraw previously loaned coins from the capital facility.
	Withdraw *lvtypes.MsgWithdraw `json:"withdraw,omitempty"`
	// Used to do withdraw maximum assets by supplier.
	MaxWithdraw *lvtypes.MsgMaxWithdraw `json:"max_withdraw,omitempty"`
	// Used to enable an amount of selected uTokens as collateral.
	Collateralize *lvtypes.MsgCollateralize `json:"collateralize,omitempty"`
	// Used to disable amount of an selected uTokens as collateral.
	Decollateralize *lvtypes.MsgDecollateralize `json:"decollateralize,omitempty"`
	// Used to borrowing coins from the capital facility.
	Borrow *lvtypes.MsgBorrow `json:"borrow,omitempty"`
	// Used to borrowing coins from the capital facility.
	MaxBorrow *lvtypes.MsgMaxBorrow `json:"max_borrow,omitempty"`
	// Used to repaying borrowed coins to the capital facility.
	Repay *lvtypes.MsgRepay `json:"repay,omitempty"`
	// Used to repaying a different user's borrowed coins
	// to the capital facility in exchange for some of their collateral.
	Liquidate *lvtypes.MsgLiquidate `json:"liquidate,omitempty"`
	// Used to do supply and collateralize their assets.
	SupplyCollateral *lvtypes.MsgSupplyCollateral `json:"supply_collateral,omitempty"`
}
