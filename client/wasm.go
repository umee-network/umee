package client

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

func (c Client) WasmClient() wasmtypes.QueryClient {
	return wasmtypes.NewQueryClient(c.GrpcConn)
}

func (c Client) QueryContract(contractAddr string, query []byte) (*wasmtypes.QuerySmartContractStateResponse, error) {
	ctx, cancel := c.NewCtxWithTimeout()
	defer cancel()

	return c.WasmClient().SmartContractState(ctx, &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: query,
	})
}
