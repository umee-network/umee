package keeper_test

import (
	"encoding/json"

	gravitytypes "github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v2/testing"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/umee-network/umee/app"
	"github.com/umee-network/umee/tests/util"
)

func SetupTestingApp(valSet *tmtypes.ValidatorSet) func() (ibctesting.TestingApp, map[string]json.RawMessage) {
	return func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		db := dbm.NewMemDB()
		encConfig := app.MakeEncodingConfig()
		umeeApp := app.New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, app.DefaultNodeHome, 5, encConfig, app.EmptyAppOptions{})
		cdc := encConfig.Marshaler

		genesisState := app.NewDefaultGenesisState(cdc)

		var gravityGenState gravitytypes.GenesisState
		if err := cdc.UnmarshalJSON(genesisState[gravitytypes.ModuleName], &gravityGenState); err != nil {
			panic(err)
		}

		delegateKeys := make([]gravitytypes.MsgSetOrchestratorAddress, len(valSet.Validators))
		for i, val := range valSet.Validators {
			_, _, ethAddr, err := util.GenerateRandomEthKey()
			if err != nil {
				panic(err)
			}

			gravityEthAddr, err := gravitytypes.NewEthAddress(ethAddr.Hex())
			if err != nil {
				panic(err)
			}

			delegateKeys[i] = *gravitytypes.NewMsgSetOrchestratorAddress(
				sdk.ValAddress(val.Address),
				sdk.AccAddress(val.Address),
				*gravityEthAddr,
			)
		}

		gravityGenState.DelegateKeys = delegateKeys

		bz, err := cdc.MarshalJSON(&gravityGenState)
		if err != nil {
			panic(err)
		}
		genesisState[gravitytypes.ModuleName] = bz

		return umeeApp, genesisState
	}
}

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort

	return path
}

func AddressFromString(address string) string {
	return sdk.AccAddress(crypto.AddressHash([]byte(address))).String()
}
