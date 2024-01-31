package uics20

import (
	stderrors "errors"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/uibc"
	"github.com/umee-network/umee/v6/x/uibc/quota"
)

var _ porttypes.IBCModule = ICS20Module{}

// ICS20Module implements ibcporttypes.IBCModule for ICS20 transfer middleware.
// It overwrites OnAcknowledgementPacket and OnTimeoutPacket to revert
// quota update on acknowledgement error or timeout.
type ICS20Module struct {
	porttypes.IBCModule
	kb       quota.KeeperBuilder
	leverage ltypes.MsgServer
	cdc      codec.JSONCodec
}

// NewICS20Module is an IBCMiddlware constructor.
// `app` must be an ICS20 app.
func NewICS20Module(app porttypes.IBCModule, cdc codec.JSONCodec, k quota.KeeperBuilder, l ltypes.MsgServer,
) ICS20Module {
	return ICS20Module{
		IBCModule: app,
		kb:        k,
		leverage:  l,
		cdc:       cdc,
	}
}

// OnRecvPacket is called when a receiver chain receives a packet from SendPacket.
func (im ICS20Module) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress,
) exported.Acknowledgement {
	ftData, err := deserializeFTData(im.cdc, packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}
	qk := im.kb.Keeper(&ctx)
	if ackResp := qk.IBCOnRecvPacket(ftData, packet); ackResp != nil && !ackResp.Success() {
		return ackResp
	}

	ack := im.IBCModule.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}
	if ftData.Memo != "" {
		logger := ctx.Logger()
		msgs, err := DeserializeMemoMsgs(im.cdc, []byte(ftData.Memo))
		if err != nil {
			// TODO: need to verify if we want to stop the handle the error or revert the ibc transerf
			//   -> same logic in dispatchMemoMsgs
			logger.Error("can't JSON deserialize ftData Memo, expecting list of Msg", "err", err)
		} else {
			// TODO: need to handle fees!
			logger.Info("handling IBC transfer with memo", "sender", ftData.Sender,
				"receiver", ftData.Receiver)

			// TODO: we need to rework this if this is not a case, and check receiver!
			if ftData.Sender != ftData.Receiver {
				logger.Error("sender and receiver are not the same")
			}

			receiver, err := sdk.AccAddressFromBech32(ftData.Receiver)
			if err != nil {
				logger.Error("can't parse bech32 address", "err", err)
				return ack
			}
			amount, ok := sdk.NewIntFromString(ftData.Amount)
			if !ok {
				logger.Error("can't parse transfer amount", "amount", ftData.Amount)
				return ack
			}
			im.dispatchMemoMsgs(&ctx, receiver, sdk.NewCoin(ftData.Denom, amount), msgs)
		}
	}

	return ack
}

// OnAcknowledgementPacket is called on the packet sender chain, once the receiver acknowledged
// the packet reception.
func (im ICS20Module) OnAcknowledgementPacket(
	ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := im.cdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return errors.Wrap(err, "cannot unmarshal ICS-20 transfer packet acknowledgement")
	}
	if _, isErr := ack.Response.(*channeltypes.Acknowledgement_Error); isErr {
		// we don't return to propagate the ack error to the other layers
		im.onAckErr(&ctx, packet)
	}

	return im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements types.Middleware
func (im ICS20Module) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	im.onAckErr(&ctx, packet)
	return im.IBCModule.OnTimeoutPacket(ctx, packet, relayer)
}

func (im ICS20Module) onAckErr(ctx *sdk.Context, packet channeltypes.Packet) {
	ftData, err := deserializeFTData(im.cdc, packet)
	if err != nil {
		// we only log error, because we want to propagate the ack to other layers.
		ctx.Logger().Error("can't revert quota update", "err", err)
	}
	qk := im.kb.Keeper(ctx)
	qk.IBCRevertQuotaUpdate(ftData.Amount, ftData.Denom)
}

// runs messages encoded in the ICS20 memo.
// NOTE: storage is forked, and only committed (flushed) if all messages pass and if all
// messages are supported. Otherwise the fork storage is discarded.
func (im ICS20Module) dispatchMemoMsgs(ctx *sdk.Context, receiver sdk.AccAddress, sent sdk.Coin, msgs []sdk.Msg) {
	logger := ctx.Logger().With("scope", "ics20-OnRecvPacket")
	if len(msgs) == 0 {
		return // nothing to do
	}

	if err := im.validateMemoMsg(receiver, sent, msgs); err != nil {
		logger.Error("ics20 memo messages are not valid.", "err", err)
		return
	}

	// Caching context so that we don't update the store in case of failure.
	cacheCtx, flush := ctx.CacheContext()
	for _, m := range msgs {
		if err := im.handleMemoMsg(&cacheCtx, m); err != nil {
			// ignore changes in cacheCtx and return
			logger.Error("error dispatching", "msg: %v\t\t err: %v", m, err)
			return
		}
		logger.Debug("dispatching", "msg", m)
	}
	flush()
}

// We only support the following message combinations:
// - [MsgSupply]
// - [MsgSupplyCollateral]
// - [MsgSupplyCollateral, MsgBorrow] -- here, borrow must use
// - [MsgLiquidate]
// Signer of each message (account under charged with coins), must be the receiver of the ICS20
// transfer.
func (im ICS20Module) validateMemoMsg(receiver sdk.AccAddress, sent sdk.Coin, msgs []sdk.Msg) error {
	msgLen := len(msgs)
	if msgLen > 2 {
		return stderrors.New("ics20 memo with more than 2 messages are not supported")
	}

	for _, msg := range msgs {
		if signers := msg.GetSigners(); len(signers) != 1 || !signers[0].Equals(receiver) {
			return sdkerrors.ErrInvalidRequest.Wrapf(
				"msg signer doesn't match the receiver, expected signer: %s", receiver)
		}
	}

	var asset sdk.Coin
	switch msg := msgs[0].(type) {
	case *ltypes.MsgSupplyCollateral:
		asset = msg.Asset
	case *ltypes.MsgSupply:
		asset = msg.Asset
	case *ltypes.MsgLiquidate:
		asset = msg.Repayment
		// TODO more asserts, will be handled in other PR
	default:
		return stderrors.New("only MsgSupply, MsgSupplyCollateral and MsgLiquidate are supported as messages[0]")
	}

	if err := assertSubCoins(sent, asset); err != nil {
		return err
	}

	if msgLen == 1 {
		// early return - we don't need to do more checks
		return nil
	}

	switch msg := msgs[1].(type) {
	case *ltypes.MsgBorrow:
		if assertSubCoins(asset, msg.Asset) != nil {
			return stderrors.New("MsgBorrow must use MsgSupplyCollateral from messages[0]")
		}
	default:
		return stderrors.New("only MsgBorrow is supported as messages[1]")
	}

	return nil
}

func assertSubCoins(sent, operated sdk.Coin) error {
	if sent.Denom != operated.Denom || sent.Amount.LT(operated.Amount) {
		return stderrors.New("message must use coins sent from the transfer")
	}
	return nil
}

func (im ICS20Module) handleMemoMsg(ctx *sdk.Context, msg sdk.Msg) (err error) {
	switch msg := msg.(type) {
	case *ltypes.MsgSupply:
		_, err = im.leverage.Supply(*ctx, msg)
	case *ltypes.MsgSupplyCollateral:
		_, err = im.leverage.SupplyCollateral(*ctx, msg)
	case *ltypes.MsgBorrow:
		_, err = im.leverage.Borrow(*ctx, msg)
	case *ltypes.MsgLiquidate:
		_, err = im.leverage.Liquidate(*ctx, msg)
	default:
		err = sdkerrors.ErrInvalidRequest.Wrapf("unsupported type in the ICS20 memo: %T", msg)
	}
	return err
}

func deserializeFTData(cdc codec.JSONCodec, packet channeltypes.Packet,
) (d ics20types.FungibleTokenPacketData, err error) {
	if err = cdc.UnmarshalJSON(packet.GetData(), &d); err != nil {
		err = errors.Wrap(err, "cannot unmarshal ICS-20 transfer packet data")
	}
	return
}

func DeserializeMemoMsgs(cdc codec.JSONCodec, data []byte) ([]sdk.Msg, error) {
	var m uibc.ICS20Memo
	if err := cdc.UnmarshalJSON(data, &m); err != nil {
		return nil, err
	}
	return tx.GetMsgs(m.Messages, "memo messages")
}
