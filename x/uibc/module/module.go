package uibc

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/x/uibc"
	"github.com/umee-network/umee/v6/x/uibc/client/cli"
	"github.com/umee-network/umee/v6/x/uibc/quota"
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
	return cdc.MustMarshalJSON(uibc.DefaultGenesisState())
}

// GetQueryCmd implements module.AppModuleBasic
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// GetTxCmd implements module.AppModuleBasic
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// Name implements module.AppModuleBasic
func (AppModuleBasic) Name() string {
	return uibc.ModuleName
}

// RegisterGRPCGatewayRoutes implements module.AppModuleBasic
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	err := uibc.RegisterQueryHandlerClient(
		context.Background(), mux, uibc.NewQueryClient(clientCtx))
	util.Panic(err)
}

// RegisterInterfaces implements module.AppModuleBasic
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	uibc.RegisterInterfaces(registry)
}

// RegisterLegacyAminoCodec implements module.AppModuleBasic
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	uibc.RegisterLegacyAminoCodec(cdc)
}

// ValidateGenesis implements module.AppModuleBasic
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs uibc.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", uibc.ModuleName, err)
	}

	return gs.Validate()
}

// AppModule represents the AppModule for this module
type AppModule struct {
	AppModuleBasic
	kb quota.Builder
}

// IsAppModule implements module.AppModule.
func (AppModule) IsAppModule() {}

// IsOnePerModuleType implements module.AppModule.
func (AppModule) IsOnePerModuleType() {}

func NewAppModule(cdc codec.Codec, kb quota.Builder) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		kb:             kb,
	}
}

// ExportGenesis implements module.AppModule
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.kb.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// InitGenesis implements module.AppModule
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genState uibc.GenesisState
	cdc.MustUnmarshalJSON(data, &genState)
	am.kb.InitGenesis(ctx, genState)

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
	uibc.RegisterMsgServer(cfg.MsgServer(), quota.NewMsgServerImpl(am.kb))
	uibc.RegisterQueryServer(cfg.QueryServer(), quota.NewQuerier(am.kb))
}

// BeginBlock executes all ABCI BeginBlock logic respective to the x/uibc module.
func (am AppModule) BeginBlock(ctx sdk.Context) {
	BeginBlock(ctx, am.kb.Keeper(&ctx))
}

// EndBlock executes all ABCI EndBlock logic respective to the x/uibc module.
// It returns no validator updates.
func (am AppModule) EndBlock(_ sdk.Context) []abci.ValidatorUpdate {
	return EndBlocker()
}
