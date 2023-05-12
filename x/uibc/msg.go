package uibc

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v4/util/checkers"
)

var (
	_ sdk.Msg = &MsgGovUpdateQuota{}
	_ sdk.Msg = &MsgGovSetIBCSQuotaStatus{}
)

// GetTitle returns the title of the proposal.
func (msg *MsgGovUpdateQuota) GetTitle() string { return msg.Title }

// GetDescription implements govv1b1.Content interface.
func (msg *MsgGovUpdateQuota) GetDescription() string { return msg.Description }

// Route implements Msg
func (msg MsgGovUpdateQuota) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgGovUpdateQuota) Type() string { return sdk.MsgTypeURL(&msg) }

// String implements the Stringer interface.
func (msg *MsgGovUpdateQuota) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg *MsgGovUpdateQuota) ValidateBasic() error {
	if err := checkers.ValidateAddr(msg.Authority, "authority"); err != nil {
		return err
	}

	if msg.Total.IsNil() || !msg.Total.IsPositive() {
		return sdkerrors.ErrInvalidRequest.Wrap("total quota must be positive")
	}

	if msg.PerDenom.IsNil() || !msg.PerDenom.IsPositive() {
		return sdkerrors.ErrInvalidRequest.Wrap("quota per denom must be positive")
	}

	if msg.Total.LT(msg.PerDenom) {
		return sdkerrors.ErrInvalidRequest.Wrap("total quota must be greater than or equal to per_denom quota")
	}

	return checkers.ValidateProposal(msg.Title, msg.Description, msg.Authority)
}

// GetSignBytes implements Msg
func (msg *MsgGovUpdateQuota) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements Msg
func (msg *MsgGovUpdateQuota) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// GetTitle implements govv1b1.Content interface.
func (msg *MsgGovSetIBCSQuotaStatus) GetTitle() string { return msg.Title }

// GetDescription implements govv1b1.Content interface.
func (msg *MsgGovSetIBCSQuotaStatus) GetDescription() string { return msg.Description }

// Route implements Msg
func (msg MsgGovSetIBCSQuotaStatus) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgGovSetIBCSQuotaStatus) Type() string { return sdk.MsgTypeURL(&msg) }

// String implements the Stringer interface.
func (msg *MsgGovSetIBCSQuotaStatus) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg *MsgGovSetIBCSQuotaStatus) ValidateBasic() error {
	if err := checkers.ValidateAddr(msg.Authority, "authority"); err != nil {
		return err
	}

	if err := validateIBCQuotaStatus(msg.QuotaStatus); err != nil {
		return err
	}

	return checkers.ValidateProposal(msg.Title, msg.Description, msg.Authority)
}

// GetSignBytes implements Msg
func (msg *MsgGovSetIBCSQuotaStatus) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg *MsgGovSetIBCSQuotaStatus) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}
