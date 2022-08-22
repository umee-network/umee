package app

import (
	"encoding/json"
	"fmt"
	"time"

	gravitytypes "github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BankModule defines a custom wrapper around the x/bank module's AppModuleBasic
// implementation to provide custom default genesis state.
type BankModule struct {
	bank.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/bank module genesis state.
func (BankModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	umeeMetadata := banktypes.Metadata{
		Description: "The native staking token of the Umee network.",
		Base:        BondDenom,
		Name:        DisplayDenom,
		Display:     DisplayDenom,
		Symbol:      DisplayDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    BondDenom,
				Exponent: 0,
				Aliases: []string{
					"microumee",
				},
			},
			{
				Denom:    DisplayDenom,
				Exponent: 6,
				Aliases:  []string{},
			},
		},
	}

	genState := banktypes.DefaultGenesisState()
	genState.DenomMetadata = append(genState.DenomMetadata, umeeMetadata)

	return cdc.MustMarshalJSON(genState)
}

// StakingModule defines a custom wrapper around the x/staking module's
// AppModuleBasic implementation to provide custom default genesis state.
type StakingModule struct {
	staking.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/staking module genesis state.
func (StakingModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	params := stakingtypes.DefaultParams()
	params.BondDenom = BondDenom

	return cdc.MustMarshalJSON(&stakingtypes.GenesisState{
		Params: params,
	})
}

// CrisisModule defines a custom wrapper around the x/crisis module's
// AppModuleBasic implementation to provide custom default genesis state.
type CrisisModule struct {
	crisis.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/crisis module genesis state.
func (CrisisModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&crisistypes.GenesisState{
		ConstantFee: sdk.NewCoin(BondDenom, sdk.NewInt(1000)),
	})
}

// MintModule defines a custom wrapper around the x/mint module's
// AppModuleBasic implementation to provide custom default genesis state.
type MintModule struct {
	mint.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/mint module genesis state.
func (MintModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genState := minttypes.DefaultGenesisState()
	genState.Params.MintDenom = BondDenom

	return cdc.MustMarshalJSON(genState)
}

// GovModule defines a custom wrapper around the x/gov module's
// AppModuleBasic implementation to provide custom default genesis state.
type GovModule struct {
	gov.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/gov module genesis state.
func (GovModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	minDeposit := sdk.NewCoins(sdk.NewCoin(BondDenom, govv1.DefaultMinDepositTokens))
	genState := govv1.DefaultGenesisState()
	genState.DepositParams.MinDeposit = minDeposit

	return cdc.MustMarshalJSON(genState)
}

// SlashingModule defines a custom wrapper around the x/slashing module's
// AppModuleBasic implementation to provide custom default genesis state.
type SlashingModule struct {
	slashing.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/slashing module genesis state.
func (SlashingModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genState := slashingtypes.DefaultGenesisState()
	genState.Params.SignedBlocksWindow = 10000
	genState.Params.DowntimeJailDuration = 24 * time.Hour

	return cdc.MustMarshalJSON(genState)
}

func GenTxValidator(msgs []sdk.Msg) error {
	if n := len(msgs); n != 2 {
		return fmt.Errorf(
			"contains invalid number of messages; expected: 2; got: %d", n)
	}

	if _, ok := msgs[0].(*stakingtypes.MsgCreateValidator); !ok {
		return fmt.Errorf(
			"contains invalid message at index 0; expected: %T; got: %T",
			&stakingtypes.MsgCreateValidator{}, msgs[0])
	}

	if _, ok := msgs[1].(*gravitytypes.MsgSetOrchestratorAddress); !ok {
		return fmt.Errorf(
			"contains invalid message at index 1; expected: %T; got: %T",
			&gravitytypes.MsgSetOrchestratorAddress{}, msgs[1])
	}

	for i := range msgs {
		if err := msgs[i].ValidateBasic(); err != nil {
			return fmt.Errorf("invalid GenTx msg[%d] '%s': %s", i, msgs[i], err)
		}
	}
	return nil
}
