package uibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"gopkg.in/yaml.v3"

	"github.com/umee-network/umee/v3/util/checkers"
)

var (
	_ sdk.Msg = &MsgGovUpdateQuota{}
	_ sdk.Msg = &MsgGovUpdateTransferStatus{}
)

// GetTitle returns the title of the proposal.
func (msg *MsgGovUpdateQuota) GetTitle() string { return msg.Title }

// GetDescription returns the description of the proposal.
func (msg *MsgGovUpdateQuota) GetDescription() string { return msg.Description }

// Route implements Msg
func (msg MsgGovUpdateQuota) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgGovUpdateQuota) Type() string { return sdk.MsgTypeURL(&msg) }

// String implements the Stringer interface.
func (msg *MsgGovUpdateQuota) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg *MsgGovUpdateQuota) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	if msg.Total.IsNil() {
		return ErrInvalidQuota.Wrap("total quota shouldn't empty")
	}

	if msg.Total.IsNegative() {
		return ErrInvalidQuota.Wrapf("total quota shouldn't be negative %s", msg.Total.String())
	}

	if msg.PerDenom.IsNil() {
		return ErrInvalidQuota.Wrap("quota per denom shouldn't empty")
	}

	if msg.PerDenom.IsNegative() {
		return ErrInvalidQuota.Wrapf("quota per denom shouldn't be negative %s", msg.Total.String())
	}

	return checkers.ValidateProposal(msg.Title, msg.Description, msg.Authority)
}

// GetSignBytes implements Msg
func (msg *MsgGovUpdateQuota) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg *MsgGovUpdateQuota) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

func NewUpdateIBCTransferPauseStatus(authority, title, description string,
	ibcPauseStatus IBCTransferStatus,
) *MsgGovUpdateTransferStatus {
	return &MsgGovUpdateTransferStatus{
		Title:          title,
		Description:    description,
		Authority:      authority,
		IbcPauseStatus: ibcPauseStatus,
	}
}

// GetTitle returns the title of the proposal.
func (msg *MsgGovUpdateTransferStatus) GetTitle() string { return msg.Title }

// GetDescription returns the description of the proposal.
func (msg *MsgGovUpdateTransferStatus) GetDescription() string { return msg.Description }

// Route implements Msg
func (msg MsgGovUpdateTransferStatus) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgGovUpdateTransferStatus) Type() string { return sdk.MsgTypeURL(&msg) }

// String implements the Stringer interface.
func (msg *MsgGovUpdateTransferStatus) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg *MsgGovUpdateTransferStatus) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	if err := validateIBCTransferStatus(msg.IbcPauseStatus); err != nil {
		return err
	}

	return checkers.ValidateProposal(msg.Title, msg.Description, msg.Authority)
}

// GetSignBytes implements Msg
func (msg *MsgGovUpdateTransferStatus) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg *MsgGovUpdateTransferStatus) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

func (q *Quota) Validate() error {
	if len(q.IbcDenom) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("ibc denom shouldn't be empty")
	}

	if q.Expires == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("ibc denom quota expires shouldn't be empty")
	}

	if q.OutflowSum.IsNil() {
		return ErrInvalidQuota.Wrap("ibc denom quota expires shouldn't be empty")
	}

	if q.OutflowSum.IsNegative() {
		return ErrInvalidQuota.Wrap("ibc denom quota expires shouldn't be empty")
	}

	return nil
}
