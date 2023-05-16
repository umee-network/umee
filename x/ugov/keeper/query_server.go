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
func (q Querier) MinGasPrice(ctx context.Context, _ *ugov.QueryMinGasPrice) (*ugov.QueryMinGasPriceResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	p := q.Keeper(&sdkCtx).MinGasPrice()
	return &ugov.QueryMinGasPriceResponse{MinGasPrice: *p}, nil
}
