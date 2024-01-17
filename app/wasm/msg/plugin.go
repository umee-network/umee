package msg

import (
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	lvkeeper "github.com/umee-network/umee/v6/x/leverage/keeper"
	lvtypes "github.com/umee-network/umee/v6/x/leverage/types"
)

// Plugin wraps the msg plugin with Messengers.
type Plugin struct {
	lvMsgServer lvtypes.MsgServer
	wrapped     wasmkeeper.Messenger
}

var _ wasmkeeper.Messenger = (*Plugin)(nil)

// CustomMessageDecorator returns decorator for custom CosmWasm bindings messages
func NewMessagePlugin(leverageKeeper lvkeeper.Keeper) func(wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &Plugin{
			wrapped:     old,
			lvMsgServer: lvkeeper.NewMsgServerImpl(leverageKeeper),
		}
	}
}

// DispatchCustomMsg responsible for handling custom messages (umee native messages).
func (plugin *Plugin) DispatchCustomMsg(ctx sdk.Context, contractAddr sdk.AccAddress, rawMsg json.RawMessage) error {
	var smartcontractMessage UmeeMsg
	if err := json.Unmarshal(rawMsg, &smartcontractMessage); err != nil {
		return sdkerrors.Wrap(err, "invalid umee custom msg")
	}

	sender := contractAddr.String()
	var err error
	switch {
	case smartcontractMessage.Supply != nil:
		_, err = smartcontractMessage.HandleSupply(ctx, sender, plugin.lvMsgServer)
	case smartcontractMessage.Withdraw != nil:
		_, err = smartcontractMessage.HandleWithdraw(ctx, sender, plugin.lvMsgServer)
	case smartcontractMessage.MaxWithdraw != nil:
		_, err = smartcontractMessage.HandleMaxWithdraw(ctx, sender, plugin.lvMsgServer)
	case smartcontractMessage.Collateralize != nil:
		_, err = smartcontractMessage.HandleCollateralize(ctx, sender, plugin.lvMsgServer)
	case smartcontractMessage.Decollateralize != nil:
		_, err = smartcontractMessage.HandleDecollateralize(ctx, sender, plugin.lvMsgServer)
	case smartcontractMessage.Borrow != nil:
		_, err = smartcontractMessage.HandleBorrow(ctx, sender, plugin.lvMsgServer)
	case smartcontractMessage.MaxBorrow != nil:
		_, err = smartcontractMessage.HandleMaxBorrow(ctx, sender, plugin.lvMsgServer)
	case smartcontractMessage.Repay != nil:
		_, err = smartcontractMessage.HandleRepay(ctx, sender, plugin.lvMsgServer)
	case smartcontractMessage.Liquidate != nil:
		_, err = smartcontractMessage.HandleLiquidate(ctx, sender, plugin.lvMsgServer)
	case smartcontractMessage.SupplyCollateral != nil:
		_, err = smartcontractMessage.HandleSupplyCollateral(ctx, sender, plugin.lvMsgServer)
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
		return nil, nil, plugin.DispatchCustomMsg(ctx, contractAddr, msg.Custom)
	}
	return plugin.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}
