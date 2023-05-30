package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v5/x/ugov"
)

var _ ugov.QueryServer = Querier{}

// Querier implements a QueryServer for the x/uibc module.
type Querier struct {
	Builder
}

func NewQuerier(kb Builder) Querier {
	return Querier{kb}
}

// MinTxFees returns minimum transaction fees.
func (q Querier) MinGasPrice(ctx context.Context, _ *ugov.QueryMinGasPrice) (*ugov.QueryMinGasPriceResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &ugov.QueryMinGasPriceResponse{MinGasPrice: q.Keeper(&sdkCtx).MinGasPrice()},
		nil
}
