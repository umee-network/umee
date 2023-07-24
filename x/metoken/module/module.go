package module

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/umee-network/umee/v5/x/metoken/client/cli"

	"github.com/umee-network/umee/v5/util"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"github.com/umee-network/umee/v5/x/metoken"
	"github.com/umee-network/umee/v5/x/metoken/keeper"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic implements the AppModuleBasic interface for the x/metoken module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name implements module.AppModuleBasic
func (AppModuleBasic) Name() string {
	return metoken.ModuleName
}

// RegisterLegacyAminoCodec implements module.AppModuleBasic
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	metoken.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces implements module.AppModuleBasic
func (AppModuleBasic) RegisterInterfaces(registry types.InterfaceRegistry) {
	metoken.RegisterInterfaces(registry)
}

// DefaultGenesis implements module.AppModuleBasic
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(metoken.DefaultGenesisState())
}

// ValidateGenesis implements module.AppModuleBasic
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs metoken.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", metoken.ModuleName, err)
	}

	return gs.Validate()
}

// RegisterGRPCGatewayRoutes implements module.AppModuleBasic
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	err := metoken.RegisterQueryHandlerClient(context.Background(), mux, metoken.NewQueryClient(clientCtx))
	util.Panic(err)
}

// GetTxCmd implements module.AppModuleBasic
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd implements module.AppModuleBasic
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

func NewAppModuleBasic(cdc codec.Codec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

// AppModule represents the AppModule for this module
type AppModule struct {
	AppModuleBasic
	kb keeper.Builder
}

// InitGenesis implements module.AppModule
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genState metoken.GenesisState
	cdc.MustUnmarshalJSON(data, &genState)
	am.kb.Keeper(&ctx).InitGenesis(genState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis implements module.AppModule
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.kb.Keeper(&ctx).ExportGenesis()
	return cdc.MustMarshalJSON(genState)
}

// RegisterInvariants implements module.AppModule
func (am AppModule) RegisterInvariants(sdk.InvariantRegistry) {}

// RegisterServices implements module.AppModule
func (am AppModule) RegisterServices(cfg module.Configurator) {
	metoken.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.kb))
	metoken.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerier(am.kb))
}

// ConsensusVersion implements module.AppModule
func (am AppModule) ConsensusVersion() uint64 {
	return 1
}

// BeginBlock executes all ABCI BeginBlock logic respective to the x/metoken module.
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock executes all ABCI EndBlock logic respective to the x/metoken module.
// It returns no validator updates.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return EndBlocker(am.kb.Keeper(&ctx))
}

func NewAppModule(cdc codec.Codec, kb keeper.Builder) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		kb:             kb,
	}
}

// DEPRECATED

func (AppModule) LegacyQuerierHandler(*codec.LegacyAmino) sdk.Querier { return nil }
func (AppModule) QuerierRoute() string                                { return "" }
func (AppModule) Route() sdk.Route                                    { return sdk.Route{} }
