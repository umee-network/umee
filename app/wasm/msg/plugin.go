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
func (plugin *Plugin) DispatchCustomMsg(ctx sdk.Context, rawMsg json.RawMessage) error {
	var smartcontractMessage UmeeMsg
	if err := json.Unmarshal(rawMsg, &smartcontractMessage); err != nil {
		return sdkerrors.Wrap(err, "umee custom msg")
	}

	var err error
	sdkCtx := sdk.WrapSDKContext(ctx)

	if smartcontractMessage.Supply != nil {
		_, err = smartcontractMessage.HandleSupply(sdkCtx, plugin.lvMsgServer)
	} else if smartcontractMessage.Withdraw != nil {
		_, err = smartcontractMessage.HandleWithdraw(sdkCtx, plugin.lvMsgServer)
	} else if smartcontractMessage.MaxWithdraw != nil {
		_, err = smartcontractMessage.HandleMaxWithdraw(sdkCtx, plugin.lvMsgServer)
	} else if smartcontractMessage.Collateralize != nil {
		_, err = smartcontractMessage.HandleCollateralize(sdkCtx, plugin.lvMsgServer)
	} else if smartcontractMessage.Decollateralize != nil {
		_, err = smartcontractMessage.HandleDecollateralize(sdkCtx, plugin.lvMsgServer)
	} else if smartcontractMessage.Borrow != nil {
		_, err = smartcontractMessage.HandleBorrow(sdkCtx, plugin.lvMsgServer)
	} else if smartcontractMessage.MaxBorrow != nil {
		_, err = smartcontractMessage.HandleMaxBorrow(sdkCtx, plugin.lvMsgServer)
	} else if smartcontractMessage.Repay != nil {
		_, err = smartcontractMessage.HandleRepay(sdkCtx, plugin.lvMsgServer)
	} else if smartcontractMessage.Liquidate != nil {
		_, err = smartcontractMessage.HandleLiquidate(sdkCtx, plugin.lvMsgServer)
	} else if smartcontractMessage.SupplyCollateral != nil {
		_, err = smartcontractMessage.HandleSupplyCollateral(sdkCtx, plugin.lvMsgServer)
	} else {
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
