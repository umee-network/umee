package gmp

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	ibcutil "github.com/umee-network/umee/v6/util/ibc"
)

type Handler struct {
}

var _ GeneralMessageHandler = Handler{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h Handler) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, data ics20types.FungibleTokenPacketData,
) error {
	logger := ctx.Logger().With("handler", "gmp_handler")
	var msg Message
	var err error

	if err = json.Unmarshal([]byte(data.GetMemo()), &msg); err != nil {
		logger.With(err).Error("cannot unmarshal memo")
		return err
	}

	switch msg.Type {
	case TypeGeneralMessage:
		err := h.HandleGeneralMessage(ctx, msg.SourceAddress, msg.SourceAddress, data.Receiver, msg.Payload)
		if err != nil {
			logger.Error("err at HandleGeneralMessage", err)
		}
	case TypeGeneralMessageWithToken:
		// parse the transfer amount
		amt, ok := sdk.NewIntFromString(data.Amount)
		if !ok {
			return errors.Wrapf(
				ics20types.ErrInvalidAmount,
				"unable to parse transfer amount (%s) into sdk.Int",
				data.Amount,
			)
		}
		denom := ibcutil.ParseDenom(packet, data.Denom)
		err := h.HandleGeneralMessageWithToken(ctx, msg.SourceAddress, msg.SourceAddress, data.Receiver,
			msg.Payload, sdk.NewCoin(denom, amt))

		if err != nil {
			logger.Error("err at HandleGeneralMessageWithToken", err)
		}
	default:
		logger.With(fmt.Errorf("unrecognized message type: %d", msg.Type)).Error("unrecognized gmp message")
	}

	return err
}

func (h Handler) HandleGeneralMessage(ctx sdk.Context, srcChain, srcAddress string, destAddress string,
	payload []byte) error {
	ctx.Logger().Info("HandleGeneralMessage called",
		"srcChain", srcChain,
		"srcAddress", srcAddress,
		"destAddress", destAddress,
		"payload", payload,
		"handler", "gmp-handler",
	)
	return nil
}

func (h Handler) HandleGeneralMessageWithToken(ctx sdk.Context, srcChain, srcAddress string, destAddress string,
	payload []byte, coin sdk.Coin) error {
	ctx.Logger().Info("HandleGeneralMessageWithToken called",
		"srcChain", srcChain,
		"srcAddress", srcAddress,
		"destAddress", destAddress,
		"payload", payload,
		"coin", coin,
		"handler", "gmp-handler",
	)
	return nil
}
