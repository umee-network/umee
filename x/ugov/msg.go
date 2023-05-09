package ugov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/checkers"
)

var (
	_ sdk.Msg = &MsgGovUpdateMinGasPrice{}
)

// ValidateBasic implements Msg
func (msg *MsgGovUpdateMinGasPrice) ValidateBasic() error {
	if err := checkers.ValidateAddr(msg.Authority, "authority"); err != nil {
		return err
	}

	return msg.MinGasPrice.Validate()
}

// GetSignBytes implements Msg
func (msg *MsgGovUpdateMinGasPrice) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}
