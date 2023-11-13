package setup

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

const GaiaChainID = "test-gaia-chain"

type gaiaValidator struct {
	index    int
	mnemonic string
	keyInfo  keyring.Record
}

func (g *gaiaValidator) instanceName() string {
	return fmt.Sprintf("gaiaval%d", g.index)
}

func (c *chain) createAndInitGaiaValidator(cdc codec.Codec) error {
	// create gaia validator
	gaiaVal := c.createGaiaValidator(0)

	// create keys
	mnemonic, info, err := createMemoryKey(cdc)
	if err != nil {
		return err
	}

	gaiaVal.keyInfo = *info
	gaiaVal.mnemonic = mnemonic

	c.GaiaValidators = append(c.GaiaValidators, gaiaVal)

	return nil
}

func (c *chain) createGaiaValidator(index int) *gaiaValidator {
	return &gaiaValidator{
		index: index,
	}
}

func (s *E2ETestSuite) runGaiaNetwork() {
	s.T().Log("starting Gaia network container...")

	tmpDir, err := os.MkdirTemp("", "umee-e2e-testnet-gaia-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	gaiaVal := s.Chain.GaiaValidators[0]

	gaiaCfgPath := path.Join(tmpDir, "cfg")
	s.Require().NoError(os.MkdirAll(gaiaCfgPath, 0o755))

	_, err = copyFile(
		filepath.Join("./scripts/", "gaia_bootstrap.sh"),
		filepath.Join(gaiaCfgPath, "gaia_bootstrap.sh"),
	)
	s.Require().NoError(err)

	s.GaiaResource, err = s.DkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name: gaiaVal.instanceName(),
			// Note: we are using this image for testing purpose
			Repository: "ghcr.io/umee-network/gaia-e2e",
			Tag:        "latest",
			NetworkID:  s.DkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/root/.gaia", tmpDir),
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"1317/tcp":  {{HostIP: "", HostPort: "1417"}},
				"9090/tcp":  {{HostIP: "", HostPort: "9190"}},
				"26656/tcp": {{HostIP: "", HostPort: "27656"}},
				"26657/tcp": {{HostIP: "", HostPort: "27657"}},
			},
			Env: []string{
				fmt.Sprintf("UMEE_E2E_GAIA_CHAIN_ID=%s", GaiaChainID),
				fmt.Sprintf("UMEE_E2E_GAIA_VAL_MNEMONIC=%s", gaiaVal.mnemonic),
			},
			Entrypoint: []string{
				"sh",
				"-c",
				"chmod +x /root/.gaia/cfg/gaia_bootstrap.sh && /root/.gaia/cfg/gaia_bootstrap.sh",
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	endpoint := fmt.Sprintf("tcp://%s", s.GaiaResource.GetHostPort("26657/tcp"))
	s.gaiaRPC, err = rpchttp.New(endpoint, "/websocket")
	s.Require().NoError(err)

	s.Require().Eventually(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			status, err := s.gaiaRPC.Status(ctx)
			if err != nil {
				return false
			}

			// let the node produce a few blocks
			if status.SyncInfo.CatchingUp || status.SyncInfo.LatestBlockHeight < 1 {
				return false
			}

			return true
		},
		5*time.Minute,
		time.Second,
		"gaia node failed to produce blocks",
	)

	s.T().Logf("✅ Started Gaia network container: %s", s.GaiaResource.Container.ID)
}
