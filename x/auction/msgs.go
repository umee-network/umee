package auction

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/util/checkers"
)

var (
	_ sdk.Msg = &MsgGovSetRewardsParams{}
	_ sdk.Msg = &MsgRewardsBid{}
)

const minBidDuration = 3600 // 1h in seconds

//
// MsgGovSetRewardsParams
//

// ValidateBasic implements Msg
func (msg *MsgGovSetRewardsParams) ValidateBasic() error {
	errs := checkers.AssertGovAuthority(msg.Authority)
	return errors.Join(errs, checkers.NumberMin(msg.Params.BidDuration, minBidDuration, "bid_duration"))
}

func (msg *MsgGovSetRewardsParams) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}

// LegacyMsg.Type implementations
func (msg MsgGovSetRewardsParams) Route() string { return "" }
func (msg MsgGovSetRewardsParams) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg *MsgGovSetRewardsParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

//
// MsgRewardsBid
//

// ValidateBasic implements Msg
func (msg *MsgRewardsBid) ValidateBasic() error {
	errs := checkers.ValidateAddr(msg.Sender, "sender")
	errs = errors.Join(errs, checkers.NumberPositive(msg.Id, "auction ID"))
	return errors.Join(errs, checkers.BigNumPositive(msg.Amount, "bid_amount"))
}

func (msg *MsgRewardsBid) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Sender)
}

// LegacyMsg.Type implementations
func (msg MsgRewardsBid) Route() string { return "" }
func (msg MsgRewardsBid) Type() string  { return sdk.MsgTypeURL(&msg) }
func (msg *MsgRewardsBid) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}
