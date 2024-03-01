package uics20

import (
	"errors"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/uibc"
)

type MemoHandler struct {
	cdc      codec.JSONCodec
	leverage ltypes.MsgServer
}

// See ICS20Module.OnRecvPacket for the flow
func (mh MemoHandler) onRecvPacketPre(
	ctx *sdk.Context, packet ibcexported.PacketI, ftData ics20types.FungibleTokenPacketData,
) ([]sdk.Msg, sdk.AccAddress, []string, error) {
	var events []string
	memo, err := deserializeMemo(mh.cdc, []byte(ftData.Memo))
	if err != nil {
		recvPacketLogger(ctx).Debug("Not recognized ICS20 memo, ignoring hook execution", "err", err)
		return nil, nil, nil, nil
	}
	var msgs []sdk.Msg
	var fallbackReceiver sdk.AccAddress
	if memo.FallbackAddr != "" {
		if fallbackReceiver, err = sdk.AccAddressFromBech32(memo.FallbackAddr); err != nil {
			return nil, nil, nil,
				sdkerrors.Wrap(err, "ICS20 memo fallback_addr defined, but not formatted correctly")
		}
	}

	msgs, err = memo.GetMsgs()
	if err != nil {
		e := "ICS20 memo recognized, but can't unpack memo.messages: " + err.Error()
		events = append(events, e)
		return nil, fallbackReceiver, events, nil
	}

	receiver, err := sdk.AccAddressFromBech32(ftData.Receiver)
	if err != nil { // must not happen
		return nil, nil, nil, sdkerrors.Wrap(err, "can't parse ftData.Receiver bech32 address")
	}
	amount, ok := sdk.NewIntFromString(ftData.Amount)
	if !ok { // must not happen
		return nil, nil, nil, fmt.Errorf("can't parse transfer amount: %s [%w]", ftData.Amount, err)
	}
	ibcDenom := uibc.ExtractDenomFromPacketOnRecv(packet, ftData.Denom)
	sentCoin := sdk.NewCoin(ibcDenom, amount)
	if err := mh.validateMemoMsg(receiver, sentCoin, msgs); err != nil {
		events = append(events, "memo.messages are not valid, err: "+err.Error())
		return nil, fallbackReceiver, events, nil
	}

	return msgs, fallbackReceiver, events, nil
}

// runs messages encoded in the ICS20 memo.
// NOTE: we fork the store and only commit if all messages pass. Otherwise the fork store
// is discarded.
func (mh MemoHandler) dispatchMemoMsgs(ctx *sdk.Context, msgs []sdk.Msg) error {
	if len(msgs) == 0 {
		return nil // quick return - we have nothing to handle
	}

	// Caching context so that we don't update the store in case of failure.
	cacheCtx, flush := ctx.CacheContext()
	logger := recvPacketLogger(ctx)
	for _, m := range msgs {
		if err := mh.handleMemoMsg(&cacheCtx, m); err != nil {
			// ignore changes in cacheCtx and return
			return sdkerrors.Wrapf(err, "error dispatching msg: %v", m)
		}
		logger.Debug("dispatching", "msg", m)
	}
	flush()
	return nil
}

// error messages used in validateMemoMsg
var (
	errNoSubCoins = errors.New("message must use only coins sent from the transfer")
	errMsg0Type   = errors.New("only MsgSupply, MsgSupplyCollateral and MsgLiquidate are supported as messages[0]")
	// errMsg1Type = errors.New("only MsgBorrow is supported as messages[1]")
)

// We only support the following message combinations:
// - [MsgSupply]
// - [MsgSupplyCollateral]
// - [MsgLiquidate]
// Signer of each message (account under charged with coins), must be the receiver of the ICS20
// transfer.
func (mh MemoHandler) validateMemoMsg(_receiver sdk.AccAddress, sent sdk.Coin, msgs []sdk.Msg) error {
	msgLen := len(msgs)
	if msgLen == 0 {
		return nil
	}
	// In this release we only support 1msg, and only messages that don't create or change
	// a borrow position
	if msgLen > 1 {
		return errors.New("ics20 memo with more than 1 message is not supported")
	}

	var (
		asset sdk.Coin
		// collateral sdk.Coin
	)
	switch msg := msgs[0].(type) {
	case *ltypes.MsgSupplyCollateral:
		asset = msg.Asset
		// collateral = asset
	case *ltypes.MsgSupply:
		asset = msg.Asset
	case *ltypes.MsgLiquidate:
		asset = msg.Repayment
	default:
		return errMsg0Type
	}

	return assertSubCoins(sent, asset)

	/**
	   TODO: handlers v2

	for _, msg := range msgs {
		if signers := msg.GetSigners(); len(signers) != 1 || !signers[0].Equals(receiver) {
			return errors.New(
				"msg signer doesn't match the receiver, expected signer: " + receiver.String())
		}
	}

	if msgLen == 1 {
		// early return - we don't need to do more checks
		return nil
	}

	switch msg := msgs[1].(type) {
	case *ltypes.MsgBorrow:
		if assertSubCoins(collateral, msg.Asset) != nil {
			return errors.New("the MsgBorrow must use MsgSupplyCollateral from messages[0]")
		}
	default:
		return errors.New(msg1typeErr)
	}

	return nil
	*/
}

func (mh MemoHandler) handleMemoMsg(ctx *sdk.Context, msg sdk.Msg) (err error) {
	switch msg := msg.(type) {
	case *ltypes.MsgSupply:
		_, err = mh.leverage.Supply(*ctx, msg)
	case *ltypes.MsgSupplyCollateral:
		_, err = mh.leverage.SupplyCollateral(*ctx, msg)
	case *ltypes.MsgBorrow:
		_, err = mh.leverage.Borrow(*ctx, msg)
	case *ltypes.MsgLiquidate:
		_, err = mh.leverage.Liquidate(*ctx, msg)
	default:
		err = fmt.Errorf("unsupported type in the ICS20 memo: %T", msg)
	}
	return err
}

func assertSubCoins(sent, operated sdk.Coin) error {
	if sent.Denom != operated.Denom || sent.Amount.LT(operated.Amount) {
		return errNoSubCoins
	}
	return nil
}

func deserializeMemo(cdc codec.JSONCodec, data []byte) (m uibc.ICS20Memo, err error) {
	return m, cdc.UnmarshalJSON(data, &m)
}
