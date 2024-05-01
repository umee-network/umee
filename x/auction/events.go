package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/sdkutil"
)

func EmitFundRewardsAuction(ctx *sdk.Context, coins sdk.Coins) {
	sdkutil.Emit(ctx, &EventFundRewardsAuction{Assets: coins})
}
