package query

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

func (c *Client) WasmClient() wasmtypes.QueryClient {
	return wasmtypes.NewQueryClient(c.GrpcConn)
}

func (c *Client) QueryContract(contractAddr, query string) (*wasmtypes.QuerySmartContractStateResponse, error) {
	ctx, cancel := c.NewCtx()
	defer cancel()

	resp, err := c.WasmClient().SmartContractState(ctx, &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: wasmtypes.RawContractMessage(query),
	})

	return resp, err
}
