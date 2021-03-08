package keeper

import (
	"github.com/umee-network/umee/x/umee/types"
)

var _ types.QueryServer = Keeper{}
