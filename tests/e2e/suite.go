package e2e

import (
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	"github.com/umee-network/umee/v5/client"
)

type IntegrationTestSuite struct {
	suite.Suite

	tmpDirs             []string
	chain               *chain
	gaiaRPC             *rpchttp.HTTP
	dkrPool             *dockertest.Pool
	dkrNet              *dockertest.Network
	gaiaResource        *dockertest.Resource
	hermesResource      *dockertest.Resource
	priceFeederResource *dockertest.Resource
	valResources        []*dockertest.Resource
	umee                client.Client
}
