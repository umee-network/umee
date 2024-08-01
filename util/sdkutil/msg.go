package sdkutil

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

// StartMsg unpacks sdk.Context and validates msg.
func StartMsg(ctx context.Context, msg sdk.Msg) (sdk.Context, error) {
	if m, ok := msg.(validateBasic); ok {
		if err := m.ValidateBasic(); err != nil {
			return sdk.Context{}, err
		}
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var err error
	if fvmsg, ok := msg.(fullValidate); ok {
		err = fvmsg.Validate(&sdkCtx)
	}
	return sdkCtx, err
}

type fullValidate interface {
	Validate(*sdk.Context) error
}

type validateBasic interface {
	ValidateBasic() error
}

// ValidateProtoMsg tries to run msg.ValidateBasic
func ValidateProtoMsg(msg proto.Message) error {
	if vm, ok := msg.(validateBasic); ok {
		if err := vm.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}
