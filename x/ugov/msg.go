package ugov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	"github.com/umee-network/umee/v5/util/checkers"
)

var (
	_, _, _ sdk.Msg = &MsgGovUpdateMinGasPrice{}, &MsgGovSetEmergencyGroup{}, &MsgGovUpdateInflationParams{}

	// amino
	_, _, _ legacytx.LegacyMsg = &MsgGovUpdateMinGasPrice{}, &MsgGovSetEmergencyGroup{}, &MsgGovUpdateInflationParams{}
)

// ValidateBasic implements Msg
func (msg *MsgGovUpdateMinGasPrice) ValidateBasic() error {
	if err := checkers.IsGovAuthority(msg.Authority); err != nil {
		return err
	}
	return msg.MinGasPrice.Validate()
}

// GetSignBytes implements Msg
func (msg *MsgGovUpdateMinGasPrice) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// String implements Stringer interface
func (msg *MsgGovUpdateMinGasPrice) String() string {
	return fmt.Sprintf("<authority: %s, min_gas_price: %s>", msg.Authority, msg.MinGasPrice.String())
}

// LegacyMsg.Type implementations

func (msg MsgGovUpdateMinGasPrice) Route() string { return "" }

func (msg MsgGovUpdateMinGasPrice) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgGovUpdateMinGasPrice) Type() string { return sdk.MsgTypeURL(&msg) }

//
// MsgGovSetEmergencyGroup
//

// Msg interface implementation

func (msg *MsgGovSetEmergencyGroup) ValidateBasic() error {
	if err := checkers.IsGovAuthority(msg.Authority); err != nil {
		return err
	}
	_, err := sdk.AccAddressFromBech32(msg.EmergencyGroup)
	return err
}

// GetSignBytes implements Msg
func (msg *MsgGovSetEmergencyGroup) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// LegacyMsg.Type implementations
func (msg MsgGovSetEmergencyGroup) Route() string { return "" }

func (msg MsgGovSetEmergencyGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgGovSetEmergencyGroup) Type() string { return sdk.MsgTypeURL(&msg) }

//
// MsgGovUpdateInflationParams
//

// Msg interface implementation

func (msg *MsgGovUpdateInflationParams) ValidateBasic() error {
	if err := checkers.IsGovAuthority(msg.Authority); err != nil {
		return err
	}
	return msg.Params.Validate()
}

// GetSignBytes implements Msg
func (msg *MsgGovUpdateInflationParams) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// LegacyMsg.Type implementations
func (msg MsgGovUpdateInflationParams) Route() string { return "" }

func (msg MsgGovUpdateInflationParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgGovUpdateInflationParams) Type() string { return sdk.MsgTypeURL(&msg) }
