package mocks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	ugov "github.com/umee-network/umee/v5/x/ugov"
)

func NewUgovParamsBuilder(ctrl *gomock.Controller) ugov.ParamsKeeperBuilder {
	return func(_ *sdk.Context) ugov.ParamsKeeper {
		return NewMockParamsKeeper(ctrl)
	}
}
