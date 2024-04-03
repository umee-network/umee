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

	var msg Message
	if err := json.Unmarshal([]byte(memoStr), &msg); err != nil {
		return msg, err
	}

	switch msg.Type {
	case TypeGeneralMessage:
		return msg, fmt.Errorf("only msg.type=%d (TypeGeneralMessageWithToken) is supported",
			TypeGeneralMessageWithToken)
	case TypeGeneralMessageWithToken:
		return msg, nil
	default:
		return msg, fmt.Errorf("unrecognized gmp message type: %d", msg.Type)
	}
}
