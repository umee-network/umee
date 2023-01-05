package msg

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	lvkeeper "github.com/umee-network/umee/v3/x/leverage/keeper"
	lvtypes "github.com/umee-network/umee/v3/x/leverage/types"
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

	switch smartcontractMessage.AssignedMsg {
	case AssignedMsgSupply:
		return smartcontractMessage.HandleSupply(ctx, plugin.lvMsgServer)
	case AssignedMsgWithdraw:
		return smartcontractMessage.HandleWithdraw(ctx, plugin.lvMsgServer)
	case AssignedMsgCollateralize:
		return smartcontractMessage.HandleCollateralize(ctx, plugin.lvMsgServer)
	case AssignedMsgDecollateralize:
		return smartcontractMessage.HandleDecollateralize(ctx, plugin.lvMsgServer)
	case AssignedMsgBorrow:
		return smartcontractMessage.HandleBorrow(ctx, plugin.lvMsgServer)
	case AssignedMsgRepay:
		return smartcontractMessage.HandleRepay(ctx, plugin.lvMsgServer)
	case AssignedMsgLiquidate:
		return smartcontractMessage.HandleLiquidate(ctx, plugin.lvMsgServer)
	case AssignedMsgSupplyCollateralize:
		return smartcontractMessage.HandleSupply(ctx, plugin.lvMsgServer)

	default:
		return wasmvmtypes.UnsupportedRequest{Kind: "invalid assigned umee msg"}
	}
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
