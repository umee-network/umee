package quota

import (
	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
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

func (k Keeper) IBCOnRecvPacket(ft ics20types.FungibleTokenPacketData, packet channeltypes.Packet,
) exported.Acknowledgement {
	params := k.GetParams()
	if !params.IbcStatus.IBCTransferEnabled() {
		return channeltypes.NewErrorAcknowledgement(ics20types.ErrReceiveDisabled)
	}

	if params.IbcStatus.OutflowQuotaEnabled() {
		ackErr := k.RecordIBCInflow(packet, ft.Denom, ft.Amount)
		if ackErr != nil {
			return ackErr
		}
	}

	return nil
}

// IBCRevertQuotaUpdate must be called on packet acknnowledgemenet error or timeout to revert
// necessary changes.
func (k Keeper) IBCRevertQuotaUpdate(amount, denom string) {
	params := k.GetParams()
	if !params.IbcStatus.OutflowQuotaEnabled() {
		return
	}
	if err := k.revertQuotaUpdateStr(amount, denom); err != nil {
		k.ctx.Logger().Error("revert quota update error", "err", err)
		sdkutil.Emit(k.ctx, &uibc.EventBadRevert{
			FailureType: "ibc-ack",
			Packet:      amount + denom,
		})
	}
}

func (k Keeper) revertQuotaUpdateStr(amount, denom string) error {
	amountInt, ok := sdkmath.NewIntFromString(amount)
	if !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid transfer amount %s", amount)
	}
	return k.UndoUpdateQuota(denom, amountInt)
}
