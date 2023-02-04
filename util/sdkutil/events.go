package sdkutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
)

// Emit proto event and log on error
func Emit(ctx *sdk.Context, event proto.Message) {
	err := ctx.EventManager().EmitTypedEvent(event)
	if err != nil {
		ctx.Logger().Error("emit event error", "err", err)
	}
}
