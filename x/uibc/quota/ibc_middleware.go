package quota

import (
	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	ibcutil "github.com/umee-network/umee/v6/util/ibc"
	"github.com/umee-network/umee/v6/util/sdkutil"
	"github.com/umee-network/umee/v6/x/uibc"
)

func (k Keeper) IBCOnSendPacket(packet []byte) error {
	params := k.GetParams()

	if !params.IbcStatus.IBCTransferEnabled() {
		return ics20types.ErrSendDisabled
	}

	funds, denom, err := ibcutil.GetFundsFromPacket(packet)
	if err != nil {
		return errors.Wrap(err, "bad packet in rate limit's SendPacket")
	}

	if params.IbcStatus.OutflowQuotaEnabled() {
		if err := k.CheckAndUpdateQuota(denom, funds); err != nil {
			return errors.Wrap(err, "sendPacket over the IBC Quota")
		}
	}

	return nil
}

func (k Keeper) IBCOnRecvPacket(packet channeltypes.Packet) exported.Acknowledgement {
	params := k.GetParams()
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
		ackErr := k.RecordIBCInflow(packet, data.Denom, data.Amount, isSourceChain)
		if ackErr != nil {
			return ackErr
		}
	}

	return nil
}

func (k Keeper) IBCOnAckPacket(ack channeltypes.Acknowledgement, packet channeltypes.Packet) error {
	if _, isErr := ack.Response.(*channeltypes.Acknowledgement_Error); isErr {
		params := k.GetParams()
		if !params.IbcStatus.OutflowQuotaEnabled() {
			return nil
		}
		err := k.revertQuotaUpdate(packet.Data)
		emitOnRevertQuota(k.ctx, "acknowledgement", packet.Data, err)
	}
	return nil
}

func (k Keeper) IBCOnTimeout(packet channeltypes.Packet) error {
	err := k.revertQuotaUpdate(packet.Data)
	emitOnRevertQuota(k.ctx, "timeout", packet.Data, err)
	return nil
}

// revertQuotaUpdate must be called on packet acknnowledgemenet error to revert necessary changes.
func (k Keeper) revertQuotaUpdate(packetData []byte) error {
	var data transfertypes.FungibleTokenPacketData
	if err := k.cdc.UnmarshalJSON(packetData, &data); err != nil {
		return errors.Wrap(err,
			"cannot unmarshal ICS-20 transfer packet data")
	}

	amount, ok := sdkmath.NewIntFromString(data.Amount)
	if !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid transfer amount %s", data.Amount)
	}

	return k.UndoUpdateQuota(data.Denom, amount)
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
