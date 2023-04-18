package msg

import (
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	lvkeeper "github.com/umee-network/umee/v4/x/leverage/keeper"
	lvtypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// Plugin wraps the msg plugin with Messengers.
type Plugin struct {
	lvMsgServer lvtypes.MsgServer
	wrapped     wasmkeeper.Messenger
}

func NewMessagePlugin(leverageKeeper lvkeeper.Keeper) *Plugin {
	return &Plugin{
		lvMsgServer: lvkeeper.NewMsgServerImpl(leverageKeeper),
	}
}

// DispatchCustomMsg responsible for handling custom messages (umee native messages).
func (plugin *Plugin) DispatchCustomMsg(ctx sdk.Context, rawMsg json.RawMessage) error {
	var smartcontractMessage UmeeMsg
	if err := json.Unmarshal(rawMsg, &smartcontractMessage); err != nil {
		return sdkerrors.Wrap(err, "umee custom msg")
	}

	var err error
	sdkCtx := sdk.WrapSDKContext(ctx)

	switch smartcontractMessage.AssignedMsg {
	case AssignedMsgSupply:
		_, err = smartcontractMessage.HandleSupply(sdkCtx, plugin.lvMsgServer)
	case AssignedMsgWithdraw:
		_, err = smartcontractMessage.HandleWithdraw(sdkCtx, plugin.lvMsgServer)
	case AssignedMsgCollateralize:
		_, err = smartcontractMessage.HandleCollateralize(sdkCtx, plugin.lvMsgServer)
	case AssignedMsgDecollateralize:
		_, err = smartcontractMessage.HandleDecollateralize(sdkCtx, plugin.lvMsgServer)
	case AssignedMsgBorrow:
		_, err = smartcontractMessage.HandleBorrow(sdkCtx, plugin.lvMsgServer)
	case AssignedMsgRepay:
		_, err = smartcontractMessage.HandleRepay(sdkCtx, plugin.lvMsgServer)
	case AssignedMsgLiquidate:
		_, err = smartcontractMessage.HandleLiquidate(sdkCtx, plugin.lvMsgServer)
	case AssignedMsgSupplyCollateral:
		_, err = smartcontractMessage.HandleSupplyCollateral(sdkCtx, plugin.lvMsgServer)

	default:
		err = wasmvmtypes.UnsupportedRequest{Kind: "invalid assigned umee msg"}
	}

	return err
}

// DispatchMsg encodes the wasmVM message and dispatches it.
func (plugin *Plugin) DispatchMsg(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	contractIBCPortID string,
	msg wasmvmtypes.CosmosMsg,
) (events []sdk.Event, data [][]byte, err error) {
	if msg.Custom != nil {
		return nil, nil, plugin.DispatchCustomMsg(ctx, msg.Custom)
	}
	return plugin.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}
