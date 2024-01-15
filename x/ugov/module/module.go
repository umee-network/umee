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
	"github.com/umee-network/umee/v6/x/ugov"
	"github.com/umee-network/umee/v6/x/ugov/client/cli"
	"github.com/umee-network/umee/v6/x/ugov/keeper"
)

const (
	consensusVersion uint64 = 2
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
	return nil // there are no tx for the moment.
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

// IsAppModule implements module.AppModule.
func (AppModule) IsAppModule() {}

// IsOnePerModuleType implements module.AppModule.
func (AppModule) IsOnePerModuleType() {}

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
	util.Panic(am.kb.Keeper(&ctx).InitGenesis(&genState))

	return []abci.ValidatorUpdate{}
}

// ConsensusVersion implements module.AppModule
func (AppModule) ConsensusVersion() uint64 {
	return consensusVersion
}

// RegisterInvariants implements module.AppModule
func (AppModule) RegisterInvariants(sdk.InvariantRegistry) {}

// RegisterServices implements module.AppModule
func (am AppModule) RegisterServices(cfg module.Configurator) {
	ugov.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.kb))
	ugov.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerier(am.kb))

	m := keeper.NewMigrator(am.kb)
	if err := cfg.RegisterMigration(ugov.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", ugov.ModuleName, err))
	}
}

// BeginBlock executes all ABCI BeginBlock logic respective to the x/uibc module.
func (am AppModule) BeginBlock(_ sdk.Context) {}

// EndBlock executes all ABCI EndBlock logic respective to the x/uibc module.
// It returns no validator updates.
func (am AppModule) EndBlock(_ sdk.Context) []abci.ValidatorUpdate {
	return nil
}
