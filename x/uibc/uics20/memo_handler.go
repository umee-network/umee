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
		recvPacketLogger(ctx).Debug("Can't deserialize ICS20 memo for hook execution", "err", err)
		return nil
	}

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
// NOTE: we fork the store and only commit if all messages pass. Otherwise the fork store
// is discarded.
func (mh MemoHandler) dispatchMemoMsgs(ctx *sdk.Context, receiver sdk.AccAddress, sent sdk.Coin, msgs []sdk.Msg) error {
	if len(msgs) == 0 {
		return nil // quick return - we have nothing to handle
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

func deserializeMemoMsgs(cdc codec.JSONCodec, data []byte) ([]sdk.Msg, error) {
	var m uibc.ICS20Memo
	if err := cdc.UnmarshalJSON(data, &m); err != nil {
		return nil, err
	}
	return tx.GetMsgs(m.Messages, "memo messages")
}
