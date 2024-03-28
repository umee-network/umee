package gmp

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h Handler) OnRecvPacket(ctx sdk.Context, coinReceived sdk.Coin, memoStr string, receiver sdk.AccAddress,
) (Message, error) {
	if len(memoStr) == 0 {
		return Message{}, nil
	}

	logger := ctx.Logger().With("handler", "gmp_handler")
	var msg Message
	var err error

	if err = json.Unmarshal([]byte(memoStr), &msg); err != nil {
		logger.Error("cannot unmarshal memo", "err", err)
		return Message{}, err
	}

	switch msg.Type {
	case TypeGeneralMessage:
		return msg, fmt.Errorf("we are not supporting general message: %d", msg.Type)
	case TypeGeneralMessageWithToken:
		return msg, nil
	default:
		logger.Error("unrecognized gmp message type: %d", msg.Type)
		return msg, fmt.Errorf("unrecognized gmp message type: %d", msg.Type)
	}
}
