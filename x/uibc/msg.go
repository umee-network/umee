package uibc

import (
	"encoding/json"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/util/checkers"
)

var (
	_ sdk.Msg = &MsgGovUpdateQuota{}
	_ sdk.Msg = &MsgGovSetIBCStatus{}
	_ sdk.Msg = &MsgGovToggleICS20Hooks{}
)

//
// MsgGovUpdateQuota
//

// String implements the Stringer interface.
func (msg *MsgGovUpdateQuota) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg *MsgGovUpdateQuota) ValidateBasic() error {
	errs := checkers.Proposal(msg.Authority, msg.Description)
	errs = errors.Join(errs, checkers.DecPositive(msg.Total, "total quota"))
	errs = errors.Join(errs, checkers.DecPositive(msg.PerDenom, "per_denom quota"))
	errs = errors.Join(errs, checkers.DecNotNegative(msg.InflowOutflowQuotaBase,
		"inflow_outflow_quota_base"))
	errs = errors.Join(errs, checkers.DecNotNegative(msg.InflowOutflowTokenQuotaBase,
		"inflow_outflow_token_quota_base"))
	errs = errors.Join(errs, checkers.DecNotNegative(msg.InflowOutflowQuotaRate,
		"inflow_outflow_quota_rate"))
	if msg.Total.LT(msg.PerDenom) {
		errs = errors.Join(errs, errors.New("total quota must be greater than or equal to per_denom quota"))
	}
	if msg.InflowOutflowQuotaBase.LT(msg.InflowOutflowTokenQuotaBase) {
		errs = errors.Join(errs, errors.New(
			"inflow_outflow_quota_base must be greater than or equal than inflow_outflow_token_quota_base"))
	}

	return errs
}

func (msg *MsgGovUpdateQuota) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// LegacyMsg.Type implementations
func (msg MsgGovUpdateQuota) Route() string { return "" }
func (msg MsgGovUpdateQuota) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg *MsgGovUpdateQuota) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

//
// MsgGovSetIBCStatus
//

// String implements the Stringer interface.
func (msg *MsgGovSetIBCStatus) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg *MsgGovSetIBCStatus) ValidateBasic() error {
	if err := checkers.Proposal(msg.Authority, msg.Description); err != nil {
		return err
	}

	return validateIBCTransferStatus(msg.IbcStatus)
}

// GetSigners implements Msg
func (msg *MsgGovSetIBCStatus) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// LegacyMsg.Type implementations
func (msg MsgGovSetIBCStatus) Route() string { return "" }
func (msg MsgGovSetIBCStatus) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg *MsgGovSetIBCStatus) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

//
// MsgGovToggleICS20Hooks
//

// String implements the Stringer interface.
func (msg *MsgGovToggleICS20Hooks) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg *MsgGovToggleICS20Hooks) ValidateBasic() error {
	return checkers.Proposal(msg.Authority, msg.Description)
}

// GetSigners implements Msg
func (msg *MsgGovToggleICS20Hooks) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// LegacyMsg.Type implementations
func (msg MsgGovToggleICS20Hooks) Route() string { return "" }
func (msg MsgGovToggleICS20Hooks) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg *MsgGovToggleICS20Hooks) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}
