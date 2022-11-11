package types

// used by appOpts?
const FlagEnableLiquidatorQuery = "enable-liquidator-query"

// import "flag"

// EnableLiquidator must be set to true to enable QueryLiquidationTargets
var EnableLiquidator *bool

/*
func init() {
	// TODO: Is there a better way to do this?
	// Ideally I'd want an app.toml or config.toml field that can be read from functions
	// in x/leverage/keeper/grpc_query.go
	EnableLiquidator = *flag.Bool("liquidator", false, "enable inefficient liquidator queries")
}
*/
