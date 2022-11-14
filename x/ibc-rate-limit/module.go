package ibc_rate_limit

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
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/client/cli"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/keeper"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/types"
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
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// GetQueryCmd implements module.AppModuleBasic
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd(types.StoreKey)
}

// GetTxCmd implements module.AppModuleBasic
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// Name implements module.AppModuleBasic
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterGRPCGatewayRoutes implements module.AppModuleBasic
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces implements module.AppModuleBasic
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterLegacyAminoCodec implements module.AppModuleBasic
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// ValidateGenesis implements module.AppModuleBasic
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return gs.Validate()
}

// AppModule represents the AppModule for this module
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
	}
}

// ExportGenesis implements module.AppModule
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(genState)
}

// InitGenesis implements module.AppModule
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genState)
	InitGenesis(ctx, am.keeper, genState)

	return []abci.ValidatorUpdate{}
}

// ConsensusVersion implements module.AppModule
func (AppModule) ConsensusVersion() uint64 {
	return 1
}

// LegacyQuerierHandler implements module.AppModule
func (AppModule) LegacyQuerierHandler(*codec.LegacyAmino) func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
	return func(sdk.Context, []string, abci.RequestQuery) ([]byte, error) {
		return nil, fmt.Errorf("legacy querier not supported for the x/%s module", types.ModuleName)
	}
}

// QuerierRoute implements module.AppModule
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// RegisterInvariants implements module.AppModule
func (AppModule) RegisterInvariants(sdk.InvariantRegistry) {

}

// RegisterServices implements module.AppModule
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerier(am.keeper))
}

// Route implements module.AppModule
func (AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// BeginBlock executes all ABCI BeginBlock logic respective to the x/ibc-rate-limit module.
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	BeginBlock(ctx, am.keeper)
}

// EndBlock executes all ABCI EndBlock logic respective to the x/ibc-rate-limit module.
// It returns no validator updates.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return EndBlocker(ctx, am.keeper)
}
