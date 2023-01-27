package uibc

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v4/util/checkers"
)

var (
	_ sdk.Msg = &MsgGovUpdateQuota{}
	_ sdk.Msg = &MsgGovSetIBCPause{}
)

// GetTitle implements govv1b1.Content interface.
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
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	if msg.Total.IsNil() || !msg.Total.IsPositive() {
		return sdkerrors.ErrInvalidRequest.Wrap("total quota must be positive")
	}

	if msg.PerDenom.IsNil() || !msg.PerDenom.IsPositive() {
		return sdkerrors.ErrInvalidRequest.Wrap("quota per denom must be positive")
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
func (msg *MsgGovSetIBCPause) GetTitle() string { return msg.Title }

// GetDescription implements govv1b1.Content interface.
func (msg *MsgGovSetIBCPause) GetDescription() string { return msg.Description }

// Route implements Msg
func (msg MsgGovSetIBCPause) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgGovSetIBCPause) Type() string { return sdk.MsgTypeURL(&msg) }

// String implements the Stringer interface.
func (msg *MsgGovSetIBCPause) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}

// ValidateBasic implements Msg
func (msg *MsgGovSetIBCPause) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	if err := validateIBCTransferStatus(msg.IbcPauseStatus); err != nil {
		return err
	}

	return checkers.ValidateProposal(msg.Title, msg.Description, msg.Authority)
}

// GetSignBytes implements Msg
func (msg *MsgGovSetIBCPause) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg *MsgGovSetIBCPause) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

func (q *Quota) Validate() error {
	if len(q.IbcDenom) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("ibc denom shouldn't be empty")
	}

	if q.OutflowSum.IsNil() || q.OutflowSum.IsNegative() {
		return sdkerrors.ErrInvalidRequest.Wrap("ibc denom quota expires shouldn't be empty")
	}

	return nil
}
