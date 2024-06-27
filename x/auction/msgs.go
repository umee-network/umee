package auction

import (
	"errors"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/util/checkers"
)

var (
	_ sdk.Msg = &MsgGovSetRewardsParams{}
	_ sdk.Msg = &MsgRewardsBid{}
)

const minBidDuration = 3600 // 1h in seconds

// MinRewardsBid is the minimum increase of the previous bid or the minimum bid if it's the
// first one. 50 UX = 50e6uumee
var MinRewardsBid = sdkmath.NewInt(50_000_000)

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
	errs = errors.Join(errs, ValidateMinRewardsBid(MinRewardsBid, msg.Amount))
	return errs
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

func ValidateMinRewardsBid(min sdkmath.Int, bid sdk.Coin) error {
	if bid.Amount.LT(min) || bid.Denom != appparams.BondDenom {
		return errors.New("bid amount must be at least " + min.String() + appparams.BondDenom)
	}
	return nil
}
