package uics20

import (
	"encoding/json"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v6/util/sdkutil"
	"github.com/umee-network/umee/v6/x/uibc"
	"github.com/umee-network/umee/v6/x/uibc/quota"
)

var _ porttypes.IBCModule = ICS20Module{}

// ICS20Module implements ibcporttypes.IBCModule for ICS20 transfer middleware.
// It overwrites OnAcknowledgementPacket and OnTimeoutPacket to revert
// quota update on acknowledgement error or timeout.
type ICS20Module struct {
	porttypes.IBCModule
	kb  quota.Builder
	cdc codec.JSONCodec
}

// NewICS20Module is an IBCMiddlware constructor.
// `app` must be an ICS20 app.
func NewICS20Module(app porttypes.IBCModule, k quota.Builder, cdc codec.JSONCodec) ICS20Module {
	return ICS20Module{
		IBCModule: app,
		kb:        k,
		cdc:       cdc,
	}
}

// OnRecvPacket implements types.Middleware
func (im ICS20Module) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress,
) exported.Acknowledgement {
	params := im.kb.Keeper(&ctx).GetParams()
	if !params.IbcStatus.IBCTransferEnabled() {
		return channeltypes.NewErrorAcknowledgement(transfertypes.ErrReceiveDisabled)
	}

	if params.IbcStatus.OutflowQuotaEnabled() {
		var data transfertypes.FungibleTokenPacketData
		if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
			ackErr := sdkerrors.ErrInvalidType.Wrap("cannot unmarshal ICS-20 transfer packet data")
			return channeltypes.NewErrorAcknowledgement(ackErr)
		}

		isSourceChain := transfertypes.SenderChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom)
		ackErr := im.kb.Keeper(&ctx).RecordIBCInflow(ctx, packet, data.Denom, data.Amount, isSourceChain)
		if ackErr != nil {
			return ackErr
		}
	}

	return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket implements types.Middleware
func (im ICS20Module) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := im.cdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return errors.Wrap(err, "cannot unmarshal ICS-20 transfer packet acknowledgement")
	}
	if _, ok := ack.Response.(*channeltypes.Acknowledgement_Error); ok {
		params := im.kb.Keeper(&ctx).GetParams()
		if params.IbcStatus.OutflowQuotaEnabled() {
			err := im.revertQuotaUpdate(ctx, packet.Data)
			emitOnRevertQuota(&ctx, "acknowledgement", packet.Data, err)
		}
	}

	return im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements types.Middleware
func (im ICS20Module) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	err := im.revertQuotaUpdate(ctx, packet.Data)
	emitOnRevertQuota(&ctx, "timeout", packet.Data, err)

	return im.IBCModule.OnTimeoutPacket(ctx, packet, relayer)
}

// revertQuotaUpdate must be called on packet acknnowledgemenet error to revert necessary changes.
func (im ICS20Module) revertQuotaUpdate(
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
