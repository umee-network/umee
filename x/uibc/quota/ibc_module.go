package quota

import (
	"encoding/json"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"

	"github.com/umee-network/umee/v4/util/sdkutil"
	"github.com/umee-network/umee/v4/x/uibc"
	"github.com/umee-network/umee/v4/x/uibc/quota/keeper"
)

var _ porttypes.Middleware = ICS20Middleware{}

// ICS20Middleware overwrites OnAcknowledgementPacket and OnTimeoutPacket to revert
// quota update on acknowledgement error or timeout.
type ICS20Middleware struct {
	porttypes.IBCModule
	kb  keeper.KeeperBuilder
	cdc codec.JSONCodec
}

// NewICS20Middleware is an IBCMiddlware constructor.
// `app` must be an ICS20 app.
func NewICS20Middleware(app porttypes.IBCModule, k keeper.KeeperBuilder, cdc codec.JSONCodec) ICS20Middleware {
	return ICS20Middleware{
		IBCModule: app,
		kb:        k,
		cdc:       cdc,
	}
}

// OnAcknowledgementPacket implements types.Middleware
func (im ICS20Middleware) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := im.cdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return errors.Wrap(err, "cannot unmarshal ICS-20 transfer packet acknowledgement")
	}
	if _, ok := ack.Response.(*channeltypes.Acknowledgement_Error); ok {
		params := im.kb.Keeper(&ctx).GetParams()
		if params.IbcStatus == uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_ENABLED {
			err := im.RevertQuotaUpdate(ctx, packet.Data)
			emitOnRevertQuota(&ctx, "acknowledgement", packet.Data, err)
		}
	}

	return im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements types.Middleware
func (im ICS20Middleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	err := im.RevertQuotaUpdate(ctx, packet.Data)
	emitOnRevertQuota(&ctx, "timeout", packet.Data, err)

	return im.IBCModule.OnTimeoutPacket(ctx, packet, relayer)
}

// RevertQuotaUpdate Notifies the contract that a sent packet wasn't properly received
func (im ICS20Middleware) RevertQuotaUpdate(
	ctx sdk.Context,
	packetData []byte,
) error {
	var data transfertypes.FungibleTokenPacketData
	if err := im.cdc.UnmarshalJSON(packetData, &data); err != nil {
		return errors.Wrap(err,
			"cannot unmarshal ICS-20 transfer packet data")
	}

	amount, ok := sdkmath.NewIntFromString(data.Amount)
	if !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid transfer amount %s", data.Amount)
	}

	return im.kb.Keeper(&ctx).UndoUpdateQuota(data.Denom, amount)
}

func ValidateReceiverAddress(packet channeltypes.Packet) error {
	var packetData transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &packetData); err != nil {
		return err
	}
	if len(packetData.Receiver) >= 4096 {
		return sdkerrors.ErrInvalidAddress.Wrapf(
			"IBC Receiver address too long. Max supported length is %d", 4096,
		)
	}
	return nil
}

// emitOnRevertQuota emits events related to quota update revert.
// packetData is ICS 20 packet data bytes.
func emitOnRevertQuota(ctx *sdk.Context, failureType string, packetData []byte, err error) {
	if err == nil {
		return
	}
	ctx.Logger().Error("revert quota update error", "err", err)
	sdkutil.Emit(ctx, &uibc.EventBadRevert{
		FailureType: failureType,
		Packet:      string(packetData),
	})
}
