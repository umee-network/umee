package module

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/umee-network/umee/v6/x/auction"
	// "github.com/umee-network/umee/v6/x/auction/client/cli"
	"github.com/umee-network/umee/v6/x/auction/keeper"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic implements the AppModuleBasic interface for the x/leverage
// module.
type AppModuleBasic struct{}

// Name returns the x/auction module's name.
func (AppModuleBasic) Name() string {
	return auction.ModuleName
}

// RegisterLegacyAminoCodec registers the x/auction module's types with a legacy
// Amino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	auction.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types.
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	auction.RegisterInterfaces(reg)
}

// DefaultGenesis returns the x/auction module's default genesis state.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(auction.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the x/auction module.
func (AppModuleBasic) ValidateGenesis(
	cdc codec.JSONCodec,
	_ client.TxEncodingConfig,
	bz json.RawMessage,
) error {
	var genState auction.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", auction.ModuleName, err)
	}

	return genState.Validate()
}

// Deprecated: RegisterRESTRoutes performs a no-op. Querying is delegated to the
// gRPC service.
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the x/leverage
// module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	err := auction.RegisterQueryHandlerClient(context.Background(), mux, auction.NewQueryClient(clientCtx))
	if err != nil {
		panic(err)
	}
}

// GetTxCmd returns the x/auction module's root tx command.
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	// TODO return cli.GetTxCmd()
	return nil
}

// GetQueryCmd returns the x/auction module's root query command.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	// TODO return cli.GetQueryCmd()
	return nil
}

// AppModule implements the AppModule interface for the x/auction module.
type AppModule struct {
	AppModuleBasic

	kb         keeper.Builder
	bankKeeper auction.BankKeeper
}

func NewAppModule(
	cdc codec.Codec, keeper keeper.Builder, bk auction.BankKeeper,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		kb:             keeper,
		bankKeeper:     bk,
	}
}

// Name returns the x/auction module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

func (AppModule) ConsensusVersion() uint64 {
	return 1
}

// RegisterServices registers gRPC services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	auction.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.kb))
	auction.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerier(am.kb))
}

// RegisterInvariants registers the x/auction module's invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// InitGenesis performs the x/auction module's genesis initialization. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genState auction.GenesisState
	cdc.MustUnmarshalJSON(data, &genState)
	am.kb.Keeper(&ctx).InitGenesis(&genState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the x/auction module's exported genesis state as raw
// JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.kb.Keeper(&ctx).ExportGenesis()
	return cdc.MustMarshalJSON(genState)
}

// BeginBlock executes all ABCI BeginBlock logic respective to the x/auction module.
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock executes all ABCI EndBlock logic respective to the x/auction module.
// It returns no validator updates.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	k := am.kb.Keeper(&ctx)
	if err := k.FinalizeRewardsAuction(); err != nil {
		ctx.Logger().With("module", "x/auction").
			Error("can't finalize rewards auction", "error", err)
	}

	return []abci.ValidatorUpdate{}
}

// sub-module accounts
var (
	subaccRewards = []byte{0x01}
)

// SubAccounts for auction Keeper
func SubAccounts() keeper.SubAccounts {
	n := AppModuleBasic{}.Name()
	return keeper.SubAccounts{
		RewardsCollect: address.Module(n, subaccRewards),
	}
}
