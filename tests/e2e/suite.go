package e2e

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	"github.com/umee-network/umee/v4/client"
)

type IntegrationTestSuite struct {
	suite.Suite

	tmpDirs             []string
	chain               *chain
	ethClient           *ethclient.Client
	gaiaRPC             *rpchttp.HTTP
	dkrPool             *dockertest.Pool
	dkrNet              *dockertest.Network
	ethResource         *dockertest.Resource
	gaiaResource        *dockertest.Resource
	hermesResource      *dockertest.Resource
	priceFeederResource *dockertest.Resource
	valResources        []*dockertest.Resource
	orchResources       []*dockertest.Resource
	gravityContractAddr string
	umee                client.Client
}
