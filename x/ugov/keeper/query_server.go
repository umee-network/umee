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
	return &ugov.QueryMinGasPriceResponse{MinGasPrice: q.Keeper(&sdkCtx).MinGasPrice()}, nil
}

func (q Querier) EmergencyGroup(ctx context.Context, _ *ugov.QueryEmergencyGroup,
) (*ugov.QueryEmergencyGroupResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &ugov.QueryEmergencyGroupResponse{EmergencyGroup: q.Keeper(&sdkCtx).EmergencyGroup().String()}, nil
}

// InflationParams returns inflation rate change params
func (q Querier) InflationParams(ctx context.Context, _ *ugov.QueryInflationParams) (
	*ugov.QueryInflationParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &ugov.QueryInflationParamsResponse{InflationParams: q.Keeper(&sdkCtx).InflationParams()}, nil
}

// InflationCycleStart return when the inflation cycle is started
func (q Querier) InflationCycleStart(ctx context.Context, _ *ugov.QueryInflationCycleStart) (
	*ugov.QueryInflationCycleStartResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	icst, err := q.Keeper(&sdkCtx).GetInflationCycleStart()
	if err != nil {
		return nil, err
	}
	return &ugov.QueryInflationCycleStartResponse{Start: &icst}, nil
}
