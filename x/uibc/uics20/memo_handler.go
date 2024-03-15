package uics20

import (
	"errors"
	"fmt"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/uibc"
	"github.com/umee-network/umee/v6/x/uibc/gmp"
)

// error messages
var (
	errWrongSigner   = errors.New("msg signer doesn't match the ICS20 receiver")
	errNoSubCoins    = errors.New("message must use only coins sent from the transfer")
	errHooksDisabled = errors.New("ics20-hooks: execute is disabled but valid hook memo is set")
	errMsg0Type      = errors.New("only MsgSupply, MsgSupplyCollateral and MsgLiquidate are supported as messages[0]")
	// errMsg1Type = errors.New("only MsgBorrow is supported as messages[1]")
)

type MemoHandler struct {
	cdc      codec.JSONCodec
	leverage ltypes.MsgServer

	executeEnabled bool
	isGMP          bool
	msgs           []sdk.Msg
	memo           string
	received       sdk.Coin

	receiver         sdk.AccAddress
	fallbackReceiver sdk.AccAddress
}

// onRecvPacketPre parses transfer Memo field and prepares the MemoHandler.
// See ICS20Module.OnRecvPacket for the flow
func (mh *MemoHandler) onRecvPacketPrepare(
	ctx *sdk.Context, packet ibcexported.PacketI, ftData ics20types.FungibleTokenPacketData,
) ([]string, error) {
	var events []string
	var err error
	mh.memo = ftData.Memo
	amount, ok := sdk.NewIntFromString(ftData.Amount)
	if !ok { // must not happen
		return nil, fmt.Errorf("can't parse transfer amount: %s", ftData.Amount)
	}
	ibcDenom := uibc.ExtractDenomFromPacketOnRecv(packet, ftData.Denom)
	mh.received = sdk.NewCoin(ibcDenom, amount)
	mh.receiver, err = sdk.AccAddressFromBech32(ftData.Receiver)
	if err != nil { // must not happen
		return nil, sdkerrors.Wrap(err, "can't parse ftData.Receiver bech32 address")
	}

	if strings.EqualFold(ftData.Sender, gmp.DefaultGMPAddress) {
		events = append(events, "axelar GMP transaction")
		mh.isGMP = true
		return events, nil
	}

	memo, err := deserializeMemo(mh.cdc, []byte(ftData.Memo))
	if err != nil {
		recvPacketLogger(ctx).Debug("Not recognized ICS20 memo, ignoring hook execution", "err", err)
		return nil, nil
	}
	if memo.FallbackAddr != "" {
		if mh.fallbackReceiver, err = sdk.AccAddressFromBech32(memo.FallbackAddr); err != nil {
			return nil,
				sdkerrors.Wrap(err, "ICS20 memo fallback_addr defined, but not formatted correctly")
		}
		if mh.fallbackReceiver.Equals(mh.receiver) {
			mh.fallbackReceiver = nil
		}
	}

	mh.msgs, err = memo.GetMsgs()
	if err != nil {
		e := "ICS20 memo recognized, but can't unpack memo.messages: " + err.Error()
		events = append(events, e)
		return events, nil
	}

	if err := mh.validateMemoMsg(); err != nil {
		events = append(events, "memo.messages are not valid, err: "+err.Error())
		return events, errMemoValidation{err}
	}

	return events, nil
}

// runs messages encoded in the ICS20 memo.
// Returns list of
// NOTE: we fork the store and only commit if all messages pass. Otherwise the fork store
// is discarded.
func (mh MemoHandler) execute(ctx *sdk.Context) error {
	if !mh.executeEnabled && (mh.isGMP || len(mh.msgs) == 0) {
		return errHooksDisabled
	}

	logger := recvPacketLogger(ctx)
	if mh.isGMP {
		gh := gmp.NewHandler()
		return gh.OnRecvPacket(*ctx, mh.received, mh.memo, mh.receiver)
	}

	if len(mh.msgs) == 0 {
		return nil // quick return - we have nothing to handle
	}

	for _, m := range mh.msgs {
		if err := mh.handleMemoMsg(ctx, m); err != nil {
			return sdkerrors.Wrapf(err, "error dispatching msg: %v", m)
		}
		logger.Debug("dispatching", "msg", m)
	}
	return nil
}

// We only support the following message combinations:
// - [MsgSupply]
// - [MsgSupplyCollateral]
// - [MsgLiquidate]
// Signer of each message (account under charged with coins), must be the receiver of the ICS20
// transfer.
// Must be called after onRecvPacketPrepare.
func (mh MemoHandler) validateMemoMsg() error {
	msgLen := len(mh.msgs)
	if msgLen == 0 {
		return nil
	}

	// In this release we only support 1msg, and only messages that don't create or change
	// a borrow position
	if msgLen > 1 {
		return errors.New("ics20 memo with more than 1 message is not supported")
	}

	for _, msg := range mh.msgs {
		if signers := msg.GetSigners(); len(signers) != 1 || !signers[0].Equals(mh.receiver) {
			return errWrongSigner
		}
	}

	var (
		asset *sdk.Coin
		// collateral sdk.Coin
	)
	switch msg := mh.msgs[0].(type) {
	case *ltypes.MsgSupplyCollateral:
		asset = &msg.Asset
		// collateral = asset
	case *ltypes.MsgSupply:
		asset = &msg.Asset
	case *ltypes.MsgLiquidate:
		asset = &msg.Repayment
	default:
		return errMsg0Type
	}

	return adjustOperatedCoin(mh.received, asset)

	/**
	   TODO: handlers v2


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
	case *ltypes.MsgLiquidate:
		_, err = mh.leverage.Liquidate(*ctx, msg)
	default:
		err = fmt.Errorf("unsupported type in the ICS20 memo: %T", msg)
	}
	return err
}

// adjustOperatedCoin assures that received and operated are of the same denom. Returns error if
// not. Moreover it updates operated amount if it is bigger than the received amount.
func adjustOperatedCoin(received sdk.Coin, operated *sdk.Coin) error {
	if received.Denom != operated.Denom {
		return errNoSubCoins
	}
	if received.Amount.LT(operated.Amount) {
		operated.Amount = received.Amount
	}
	return nil
}

func deserializeMemo(cdc codec.JSONCodec, data []byte) (m uibc.ICS20Memo, err error) {
	return m, cdc.UnmarshalJSON(data, &m)
}
