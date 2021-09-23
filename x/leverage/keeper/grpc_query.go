package keeper

import (
	"github.com/umee-network/umee/x/leverage/types"
)

var _ types.QueryServer = Keeper{}
