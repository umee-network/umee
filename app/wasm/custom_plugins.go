package wasm

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/umee-network/umee/v4/app/wasm/msg"
	"github.com/umee-network/umee/v4/app/wasm/query"
	leveragekeeper "github.com/umee-network/umee/v4/x/leverage/keeper"
	oraclekeeper "github.com/umee-network/umee/v4/x/oracle/keeper"
)

// RegisterCustomPlugins expose the queries and msgs of native modules to wasm.
func RegisterCustomPlugins(
	leverageKeeper leveragekeeper.Keeper,
	oracleKeeper oraclekeeper.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := query.NewQueryPlugin(leverageKeeper, oracleKeeper)
	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: wasmQueryPlugin.CustomQuerier(),
	})

	messagePluginOpt := wasmkeeper.WithMessageHandlerDecorator(func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return msg.NewMessagePlugin(leverageKeeper)
	})

	return []wasm.Option{
		queryPluginOpt,
		messagePluginOpt,
	}
}
