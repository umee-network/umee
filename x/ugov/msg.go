package ugov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	"github.com/umee-network/umee/v5/util/checkers"
)

var (
	_ sdk.Msg = &MsgGovUpdateMinGasPrice{}

	// amino
	_ legacytx.LegacyMsg = &MsgGovUpdateMinGasPrice{}
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

// Route implements LegacyMsg.Route
func (msg MsgGovUpdateMinGasPrice) Route() string { return "" }

// GetSignBytes implements the LegacyMsg.GetSignBytes
func (msg MsgGovUpdateMinGasPrice) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSignBytes implements the LegacyMsg.Type
func (msg MsgGovUpdateMinGasPrice) Type() string { return sdk.MsgTypeURL(&msg) }
