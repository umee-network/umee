package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/ugov"
)

var _ ugov.QueryServer = Querier{}

// Querier implements a QueryServer for the x/uibc module.
type Querier struct {
	Builder
}

// MinTxFees returns minimum transaction fees.
func (q Builder) MinTxFees(ctx context.Context, _ *ugov.QueryMinTxFees) (*ugov.QueryMinTxFeesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	fees := q.Keeper(&sdkCtx).GetFees()

	return &ugov.QueryMinTxFeesResponse{MinTxFees: fees}, nil
}
