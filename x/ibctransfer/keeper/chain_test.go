package keeper_test

import (
	"encoding/json"

	ibctesting "github.com/cosmos/ibc-go/testing"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/umee-network/umee/app"
)

type EmptyAppOptions struct{}

func (_ EmptyAppOptions) Get(o string) interface{} { return nil }

func setupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := dbm.NewMemDB()
	encConfig := app.MakeEncodingConfig()
	umeeApp := app.New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, app.DefaultNodeHome, 5, encConfig, EmptyAppOptions{})
	genesisState := app.NewDefaultGenesisState(encConfig.Marshaler)

	return umeeApp, genesisState
}

func newTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort

	return path
}
