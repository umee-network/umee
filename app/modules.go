package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type bankModule struct {
	bank.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/bank module genesis state.
func (bankModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	umeeMetadata := banktypes.Metadata{
		Description: "The native staking token of the Umee network.",
		Base:        BondDenom,
		Name:        BondDenom,
		Display:     "umee",
		Symbol:      "UMEE",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    BondDenom,
				Exponent: 0,
				Aliases: []string{
					"microumee",
				},
			},
			{
				Denom:    "umee",
				Exponent: 6,
				Aliases:  []string{},
			},
		},
	}

	genState := banktypes.DefaultGenesisState()
	genState.DenomMetadata = append(genState.DenomMetadata, umeeMetadata)

	return cdc.MustMarshalJSON(genState)
}

// stakingModule wraps the x/staking module in order to overwrite specific
// ModuleManager APIs.
type stakingModule struct {
	staking.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/staking module genesis state.
func (stakingModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	params := stakingtypes.DefaultParams()
	params.BondDenom = BondDenom

	return cdc.MustMarshalJSON(&stakingtypes.GenesisState{
		Params: params,
	})
}

type crisisModule struct {
	crisis.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/crisis module genesis state.
func (crisisModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&crisistypes.GenesisState{
		ConstantFee: sdk.NewCoin(BondDenom, sdk.NewInt(1000)),
	})
}
