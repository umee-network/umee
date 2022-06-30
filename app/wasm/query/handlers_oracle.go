package query

import (
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	octypes "github.com/umee-network/umee/v2/x/oracle/types"
)

// HandleExchangeRates gets the exchange rates of all denoms.
func (umeeQuery UmeeQuery) HandleExchangeRates(
	ctx sdk.Context,
	queryServer octypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.ExchangeRates(sdk.WrapSDKContext(ctx), umeeQuery.ExchangeRates)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Exchange Rates", err)}
	}

	return MarshalResponse(resp)
}

// HandleActiveExchangeRates gets all active denoms.
func (umeeQuery UmeeQuery) HandleActiveExchangeRates(
	ctx sdk.Context,
	queryServer octypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.ActiveExchangeRates(sdk.WrapSDKContext(ctx), umeeQuery.ActiveExchangeRates)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{
			Kind: fmt.Sprintf("error %+v to assigned query Active Exchange Rates", err),
		}
	}

	return MarshalResponse(resp)
}

// HandleFeederDelegation gets all the feeder delegation of a validator.
func (umeeQuery UmeeQuery) HandleFeederDelegation(
	ctx sdk.Context,
	queryServer octypes.QueryServer,
) ([]byte, error) {
	resp, err := queryServer.FeederDelegation(sdk.WrapSDKContext(ctx), umeeQuery.FeederDelegation)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v to assigned query Feeder Delegation", err)}
	}

	return MarshalResponse(resp)
}
