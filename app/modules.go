package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// stakingModule wraps the x/staking module in order to overwrite specific
// ModuleManager APIs.
type stakingModule struct {
	staking.AppModuleBasic
}

// DefaultGenesis returns custom Umee staking module genesis state.
func (stakingModule) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	params := stakingtypes.DefaultParams()
	params.BondDenom = "uumee"

	return cdc.MustMarshalJSON(&stakingtypes.GenesisState{
		Params: params,
	})
}
