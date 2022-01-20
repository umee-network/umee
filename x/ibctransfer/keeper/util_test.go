package keeper_test

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v2/testing"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/umee-network/umee/app"
)

func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := dbm.NewMemDB()
	encConfig := app.MakeEncodingConfig()
	umeeApp := app.New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, app.DefaultNodeHome, 5, encConfig, app.EmptyAppOptions{})
	genesisState := app.NewDefaultGenesisState(encConfig.Marshaler)

	return umeeApp, genesisState
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
