package auction

import (
	"strconv"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/tests/accs"
	"github.com/umee-network/umee/v6/tests/tcheckers"
	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/util/coin"
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
		Amount: coin.Umee(MinRewardsBid.Int64()),
	}

	invalid := validMsg
	invalid.Sender = "not a valid acc"
	invalid.Id = 0
	invalid.Amount.Amount = sdk.ZeroInt()

	invalidAmount1 := validMsg
	invalidAmount1.Amount.Amount = sdk.NewInt(-100)

	invalidDenom := validMsg
	invalidDenom.Amount.Denom = "other"

	errAmount := "bid amount must be at least " + MinRewardsBid.String()

	tcs := []struct {
		name   string
		msg    MsgRewardsBid
		errMsg string
	}{
		{"valid msg", validMsg, ""},
		{"invalid sender", invalid, "sender"},
		{"invalid ID", invalid, "auction ID"},
		{"amount zero", invalid, errAmount},
		{"amount negative", invalidAmount1, errAmount},
		{"wrong denom", invalidDenom, errAmount},
	}
	for _, tc := range tcs {
		tcheckers.ErrorContains(t, tc.msg.ValidateBasic(), tc.errMsg, tc.name)
	}
}

func TestValidateMin(t *testing.T) {
	t.Parallel()
	min := sdk.NewInt(123)
	expectedErr := "bid amount must be at least 123uumee"
	umeeNegative := coin.Umee(0)
	umeeNegative.Amount = math.NewInt(-1)
	tcs := []struct {
		amount sdk.Coin
		errMsg string
	}{
		{coin.Umee(123), ""},
		{coin.Umee(124), ""},

		{coin.Umee(122), expectedErr},
		{coin.Umee(0), expectedErr},
		{umeeNegative, expectedErr},
		{coin.New(coin.Dollar, 1), expectedErr},
		{coin.New(coin.Dollar, 124), expectedErr},
	}

	for i, tc := range tcs {
		err := ValidateMinRewardsBid(min, tc.amount)
		tcheckers.ErrorContains(t, err, tc.errMsg, strconv.Itoa(i))
	}
}
