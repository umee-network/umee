package query

import (
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	octypes "github.com/umee-network/umee/v2/x/oracle/types"
)

// HandleMarketSummary gets the exchange rates of all denoms.
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
