package auction

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/tests/accs"
	"github.com/umee-network/umee/v6/tests/tcheckers"
	"github.com/umee-network/umee/v6/util/checkers"
)

func TestMsgGovSetRewardsParams(t *testing.T) {
	t.Parallel()
	validMsg := MsgGovSetRewardsParams{
		Authority: checkers.GovModuleAddr,
		Params:    RewardsParams{BidDuration: 3600 * 12}, // 12h
	}

	invalidAuth := validMsg
	invalidAuth.Authority = accs.Bob.String()

	invalidBidDuration1 := validMsg
	invalidBidDuration1.Params.BidDuration = 0

	invalidBidDuration2 := validMsg
	invalidBidDuration2.Params.BidDuration = 1200

	tcs := []struct {
		name   string
		msg    MsgGovSetRewardsParams
		errMsg string
	}{
		{"valid msg", validMsg, ""},
		{"wrong gov auth", invalidAuth, "expected gov account"},
		{"bid duration 0", invalidBidDuration1, "must be at least"},
		{"bid duration 1200", invalidBidDuration2, "must be at least"},
	}
	for _, tc := range tcs {
		tcheckers.ErrorContains(t, tc.msg.ValidateBasic(), tc.errMsg, tc.name)
	}
}

func TestMsgRewardsBid(t *testing.T) {
	t.Parallel()
	validMsg := MsgRewardsBid{
		Sender: accs.Alice.String(),
		Id:     12,
		Amount: sdk.NewInt(123),
	}

	invalid := validMsg
	invalid.Sender = "not a valid acc"
	invalid.Id = 0
	invalid.Amount = sdk.ZeroInt()

	invalidAmount := validMsg
	invalidAmount.Amount = sdk.NewInt(-100)

	tcs := []struct {
		name   string
		msg    MsgRewardsBid
		errMsg string
	}{
		{"valid msg", validMsg, ""},
		{"invalid sender", invalid, "sender"},
		{"invalid ID", invalid, "auction ID"},
		{"amount zero", invalid, "bid_amount: must be positive"},
		{"amount negative", invalidAmount, "bid_amount: must be positive"},
	}
	for _, tc := range tcs {
		tcheckers.ErrorContains(t, tc.msg.ValidateBasic(), tc.errMsg, tc.name)
	}
}
