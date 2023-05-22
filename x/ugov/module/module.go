package uibc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/umee-network/umee/v4/util"
	// "github.com/umee-network/umee/v4/x/ugov/client/cli"
	"github.com/umee-network/umee/v4/x/ugov"
	"github.com/umee-network/umee/v4/x/ugov/client/cli"
	"github.com/umee-network/umee/v4/x/ugov/keeper"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic is the 29-fee AppModuleBasic
type AppModuleBasic struct {
	cdc codec.Codec
}

func NewAppModuleBasic(cdc codec.Codec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

// DefaultGenesis implements module.AppModuleBasic
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(ugov.DefaultGenesis())
}

// GetQueryCmd implements module.AppModuleBasic
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// GetTxCmd implements module.AppModuleBasic
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil // TODO cli.GetTxCmd()
}

// Name implements module.AppModuleBasic
func (AppModuleBasic) Name() string {
	return ugov.ModuleName
}

// RegisterGRPCGatewayRoutes implements module.AppModuleBasic
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	err := ugov.RegisterQueryHandlerClient(
		context.Background(), mux, ugov.NewQueryClient(clientCtx))
	util.Panic(err)
}

// RegisterInterfaces implements module.AppModuleBasic
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	ugov.RegisterInterfaces(registry)
}

// RegisterLegacyAminoCodec implements module.AppModuleBasic
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	ugov.RegisterLegacyAminoCodec(cdc)
}

// ValidateGenesis implements module.AppModuleBasic
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs ugov.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", ugov.ModuleName, err)
	}

	return gs.Validate()
}

// AppModule represents the AppModule for this module
type AppModule struct {
	AppModuleBasic
	kb keeper.Builder
}

func NewAppModule(cdc codec.Codec, kb keeper.Builder) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		kb:             kb,
	}
}

// ExportGenesis implements module.AppModule
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.kb.Keeper(&ctx).ExportGenesis()
	return cdc.MustMarshalJSON(genState)
}

// InitGenesis implements module.AppModule
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genState ugov.GenesisState
	cdc.MustUnmarshalJSON(data, &genState)
	util.Panic(
		am.kb.Keeper(&ctx).InitGenesis(&genState))

	return []abci.ValidatorUpdate{}
}

// ConsensusVersion implements module.AppModule
func (AppModule) ConsensusVersion() uint64 {
	return 1
}

// RegisterInvariants implements module.AppModule
func (AppModule) RegisterInvariants(sdk.InvariantRegistry) {}

// RegisterServices implements module.AppModule
func (am AppModule) RegisterServices(cfg module.Configurator) {
	ugov.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.kb))
	ugov.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerier(am.kb))
}

// BeginBlock executes all ABCI BeginBlock logic respective to the x/uibc module.
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock executes all ABCI EndBlock logic respective to the x/uibc module.
// It returns no validator updates.
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return nil
}

// DEPRECATED

func (AppModule) LegacyQuerierHandler(*codec.LegacyAmino) sdk.Querier { return nil }
func (AppModule) QuerierRoute() string                                { return "" }
func (AppModule) Route() sdk.Route                                    { return sdk.Route{} }
