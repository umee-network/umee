package gmp

import sdk "github.com/cosmos/cosmos-sdk/types"

type GeneralMessageHandler interface {
	HandleGeneralMessage(ctx sdk.Context, srcChain, srcAddress string, destAddress string, payload []byte) error
	HandleGeneralMessageWithToken(ctx sdk.Context, srcChain, srcAddress string, destAddress string,
		payload []byte, coin sdk.Coin) error
}

const (
	DefaultGMPAddress = "axelar1dv4u5k73pzqrxlzujxg3qp8kvc3pje7jtdvu72npnt5zhq05ejcsn5qme5"
)

const (
	// TypeUnrecognized means coin type is unrecognized
	TypeUnrecognized = iota
	// TypeGeneralMessage is a pure message
	TypeGeneralMessage
	// TypeGeneralMessageWithToken is a general message with token
	TypeGeneralMessageWithToken
	// TypeSendToken is a direct token transfer
	TypeSendToken
)

// Message is attached in ICS20 packet memo field
type Message struct {
	SourceChain   string `json:"source_chain"`
	SourceAddress string `json:"source_address"`
	Payload       []byte `json:"payload"`
	Type          int64  `json:"type"`
}
