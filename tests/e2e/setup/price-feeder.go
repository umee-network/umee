package setup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	PriceFeederContainerRepo  = "ghcr.io/ojo-network/price-feeder-umee-47"
	PriceFeederServerPort     = "7171/tcp"
	PriceFeederMaxStartupTime = 20 // seconds
)

// runPriceFeeder runs the price feeder using one of the chain's validators to vote on prices.
// The container accesses the validator's key using the test keyring located in the validator's config directory.
func (s *E2ETestSuite) runPriceFeeder(valIndex int) {
	s.T().Log("starting price-feeder container...")

	umeeVal := s.Chain.Validators[valIndex]
	umeeValAddr, err := umeeVal.KeyInfo.GetAddress()
	s.Require().NoError(err)

	grpcEndpoint := fmt.Sprintf("tcp://%s:%s", umeeVal.instanceName(), "9090")
	tmrpcEndpoint := fmt.Sprintf("http://%s:%s", umeeVal.instanceName(), "26657")

	s.priceFeederResource, err = s.DkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       "umee-price-feeder",
			NetworkID:  s.DkrNet.Network.ID,
			Repository: PriceFeederContainerRepo,
			Mounts: []string{
				fmt.Sprintf("%s/:/root/.umee", umeeVal.configDir()),
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				PriceFeederServerPort: {{HostIP: "", HostPort: "7171"}},
			},
			Env: []string{
				fmt.Sprintf("PRICE_FEEDER_PASS=%s", keyringPassphrase),
				fmt.Sprintf("ACCOUNT_ADDRESS=%s", umeeValAddr),
				fmt.Sprintf("ACCOUNT_VALIDATOR=%s", sdk.ValAddress(umeeValAddr)),
				fmt.Sprintf("KEYRING_DIR=%s", "/root/.umee"),
				fmt.Sprintf("ACCOUNT_CHAIN_ID=%s", s.Chain.ID),
				fmt.Sprintf("RPC_GRPC_ENDPOINT=%s", grpcEndpoint),
				fmt.Sprintf("RPC_TMRPC_ENDPOINT=%s", tmrpcEndpoint),
			},
			Cmd: []string{
				"--skip-provider-check",
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	var endpoint string
	switch os.Getenv("DOCKER_HOST") {
	case "":
		endpoint = s.priceFeederResource.GetHostPort(PriceFeederServerPort)
	default:
		endpoint = s.priceFeederResource.Container.NetworkSettings.Networks["bridge"].IPAddress + ":" + s.priceFeederResource.GetPort(PriceFeederServerPort)
	}

	endpoint = fmt.Sprintf("http://%s/api/v1/prices", endpoint)
	s.T().Log("this is the endpoint:", endpoint, PriceFeederContainerRepo)

	checkHealth := func() bool {
		resp, err := http.Get(endpoint)
		if err != nil {
			s.T().Log("Price feeder endpoint not available", err)
			return false
		}

		defer resp.Body.Close()

		bz, err := io.ReadAll(resp.Body)
		if err != nil {
			s.T().Log("Can't get price feeder response", err)
			return false
		}

		var respBody map[string]interface{}
		if err := json.Unmarshal(bz, &respBody); err != nil {
			s.T().Log("Can't unmarshal price feed", err)
			return false
		}

		prices, ok := respBody["prices"].(map[string]interface{})
		if !ok {
			s.T().Log("price feeder: no prices")
			return false
		}

		return len(prices) > 0
	}

	isHealthy := false
	for i := 0; i < PriceFeederMaxStartupTime; i++ {
		isHealthy = checkHealth()
		if isHealthy {
			break
		}
		time.Sleep(time.Second)
	}

	if !isHealthy {
		err := s.DkrPool.Client.Logs(docker.LogsOptions{
			Container:    s.priceFeederResource.Container.ID,
			OutputStream: os.Stdout,
			ErrorStream:  os.Stderr,
			Stdout:       true,
			Stderr:       true,
			Tail:         "false",
		})
		if err != nil {
			s.T().Log("Error retrieving price feeder logs", err)
		}

		s.T().Fatal("price-feeder not healthy")
	}

	s.T().Logf("started price-feeder container: %s", s.priceFeederResource.Container.ID)
}
