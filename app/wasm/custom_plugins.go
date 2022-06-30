package wasm

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/umee-network/umee/v2/app/wasm/query"
	leveragekeeper "github.com/umee-network/umee/v2/x/leverage/keeper"
	oraclekeeper "github.com/umee-network/umee/v2/x/oracle/keeper"
)

func registerCustomPlugins(
	leverageKeeper leveragekeeper.Keeper,
	oracleKeeper oraclekeeper.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := query.NewQueryPlugin(leverageKeeper, oracleKeeper)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: wasmQueryPlugin.CustomQuerier(),
	})

	return []wasm.Option{
		queryPluginOpt,
	}
}
