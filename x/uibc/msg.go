package uibc

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v6/util/checkers"
)

var (
	_ sdk.Msg = &MsgGovUpdateQuota{}
	_ sdk.Msg = &MsgGovSetIBCStatus{}
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
	if err := checkers.AssertGovAuthority(msg.Authority); err != nil {
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

	return nil
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
	if err := checkers.AssertGovAuthority(msg.Authority); err != nil {
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
