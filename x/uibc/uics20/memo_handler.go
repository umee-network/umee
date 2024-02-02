package uics20

import (
	"errors"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/uibc"
)

type MemoHandler struct {
	cdc      codec.JSONCodec
	leverage ltypes.MsgServer
}

func (mh MemoHandler) onRecvPacket(ctx *sdk.Context, ftData ics20types.FungibleTokenPacketData) error {
	msgs, err := deserializeMemoMsgs(mh.cdc, []byte(ftData.Memo))
	if err != nil {
		// TODO: need to verify if we want to stop the handle the error or revert the ibc transerf
		//   -> same logic in dispatchMemoMsgs
		return sdkerrors.Wrap(err, "can't JSON deserialize ftData Memo, expecting list of Msg [%w]")
	}
	// TODO: need to handle fees!
	// TODO: verify correctly if receiver and sender are similar

	receiver, err := sdk.AccAddressFromBech32(ftData.Receiver)
	if err != nil {
		return sdkerrors.Wrap(err, "can't parse bech32 address")
	}
	amount, ok := sdk.NewIntFromString(ftData.Amount)
	if !ok {
		return fmt.Errorf("can't parse transfer amount: %s [%w]", ftData.Amount, err)
	}
	return mh.dispatchMemoMsgs(ctx, receiver, sdk.NewCoin(ftData.Denom, amount), msgs)
}

// runs messages encoded in the ICS20 memo.
// NOTE: storage is forked, and only committed (flushed) if all messages pass and if all
// messages are supported. Otherwise the fork storage is discarded.
func (mh MemoHandler) dispatchMemoMsgs(ctx *sdk.Context, receiver sdk.AccAddress, sent sdk.Coin, msgs []sdk.Msg) error {
	if len(msgs) == 0 {
		return nil // nothing to do
	}

	if err := mh.validateMemoMsg(receiver, sent, msgs); err != nil {
		return sdkerrors.Wrap(err, "ics20 memo messages are not valid.")
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

// We only support the following message combinations:
// - [MsgSupply]
// - [MsgSupplyCollateral]
// - [MsgSupplyCollateral, MsgBorrow] -- here, borrow must use
// - [MsgLiquidate]
// Signer of each message (account under charged with coins), must be the receiver of the ICS20
// transfer.
func (mh MemoHandler) validateMemoMsg(receiver sdk.AccAddress, sent sdk.Coin, msgs []sdk.Msg) error {
	msgLen := len(msgs)
	if msgLen > 2 {
		return errors.New("ics20 memo with more than 2 messages are not supported")
	}

	for _, msg := range msgs {
		if signers := msg.GetSigners(); len(signers) != 1 || !signers[0].Equals(receiver) {
			return errors.New(
				"msg signer doesn't match the receiver, expected signer: " + receiver.String())
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
		return errors.New("only MsgSupply, MsgSupplyCollateral and MsgLiquidate are supported as messages[0]")
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
			return errors.New("MsgBorrow must use MsgSupplyCollateral from messages[0]")
		}
	default:
		return errors.New("only MsgBorrow is supported as messages[1]")
	}

	return nil
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
		return errors.New("message must use coins sent from the transfer")
	}
	return nil
}

func deserializeMemoMsgs(cdc codec.JSONCodec, data []byte) ([]sdk.Msg, error) {
	var m uibc.ICS20Memo
	if err := cdc.UnmarshalJSON(data, &m); err != nil {
		return nil, err
	}
	return tx.GetMsgs(m.Messages, "memo messages")
}
