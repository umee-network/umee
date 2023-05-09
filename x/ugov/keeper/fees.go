package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/gogoproto"

	"github.com/umee-network/umee/v4/util/store"
	"github.com/umee-network/umee/v4/x/ugov"
)

func (k Keeper) SetMinGasPrice(p *sdk.DecCoin) error {
	return store.SetValue(k.store, keyMinGasPrice, p, "gas_price")
}
