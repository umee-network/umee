package gmp

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h Handler) OnRecvPacket(ctx sdk.Context, coinReceived sdk.Coin, memo string, receiver sdk.AccAddress,
) error {
	logger := ctx.Logger().With("handler", "gmp_handler")
	var msg Message
	var err error

	if err = json.Unmarshal([]byte(memo), &msg); err != nil {
		logger.Error("cannot unmarshal memo", "err", err)
		return err
	}

	switch msg.Type {
	case TypeGeneralMessage:
		err := h.HandleGeneralMessage(ctx, msg.SourceAddress, msg.SourceAddress, receiver, msg.Payload)
		if err != nil {
			logger.Error("err at HandleGeneralMessage", err)
		}
	case TypeGeneralMessageWithToken:
		err := h.HandleGeneralMessageWithToken(
			ctx, msg.SourceAddress, msg.SourceAddress, receiver, msg.Payload, coinReceived)
		if err != nil {
			logger.Error("err at HandleGeneralMessageWithToken", err)
		}
	default:
		logger.Error("unrecognized gmp message type: %d", msg.Type)
	}

	return err
}

func (h Handler) HandleGeneralMessage(ctx sdk.Context, srcChain, srcAddress string, receiver sdk.AccAddress,
	payload []byte) error {
	ctx.Logger().Info("HandleGeneralMessage called",
		"srcChain", srcChain,
		"srcAddress", srcAddress,
		"receiver", receiver,
		"payload", payload,
		"handler", "gmp-handler",
	)
	return nil
}

func (h Handler) HandleGeneralMessageWithToken(ctx sdk.Context, srcChain, srcAddress string,
	receiver sdk.AccAddress, payload []byte, coin sdk.Coin) error {

	ctx.Logger().Info("HandleGeneralMessageWithToken called",
		"srcChain", srcChain,
		"srcAddress", srcAddress,
		"receiver", receiver,
		"payload", payload,
		"coin", coin,
		"handler", "gmp-token-handler",
	)
	return nil
}
