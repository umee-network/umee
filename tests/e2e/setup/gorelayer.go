package setup

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func (s *E2ETestSuite) runIBCGoRelayer() {
	s.T().Log("starting gorelayer container...")

	tmpDir, err := os.MkdirTemp("", "umee-e2e-testnet-go-relayer-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	gaiaVal := s.Chain.GaiaValidators[0]
	// umeeVal for the relayer needs to be a different account
	// than what we use for runPriceFeeder.
	umeeVal := s.Chain.Validators[0]
	rlyCfgPath := path.Join(tmpDir, "relayer")

	s.Require().NoError(os.MkdirAll(rlyCfgPath, 0o755))
	_, err = copyFile(
		filepath.Join("./scripts/", "gorelayer.sh"),
		filepath.Join(rlyCfgPath, "gorelayer.sh"),
	)
	s.Require().NoError(err)

	s.HermesResource, err = s.DkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name: "umee-gaia-go-relayer",
			// Note: we are using this image for testing purpose
			Repository: "ghcr.io/umee-network/gorelayer-e2e",
			Tag:        "latest",
			NetworkID:  s.DkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/home/relayer", rlyCfgPath),
			},
			User:         "root",
			ExposedPorts: []string{"3000"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"3000/tcp": {{HostIP: "", HostPort: "3000"}},
			},
			Env: []string{
				fmt.Sprintf("UMEE_E2E_GAIA_CHAIN_ID=%s", GaiaChainID),
				fmt.Sprintf("UMEE_E2E_UMEE_CHAIN_ID=%s", s.Chain.ID),
				fmt.Sprintf("UMEE_E2E_GAIA_VAL_MNEMONIC=%s", gaiaVal.mnemonic),
				fmt.Sprintf("UMEE_E2E_UMEE_VAL_MNEMONIC=%s", umeeVal.mnemonic),
				fmt.Sprintf("UMEE_E2E_GAIA_VAL_HOST=%s", s.GaiaResource.Container.Name[1:]),
				fmt.Sprintf("UMEE_E2E_UMEE_VAL_HOST=%s", s.ValResources[0].Container.Name[1:]),
			},
			Entrypoint: []string{
				"sh",
				"-c",
				"chmod +x /home/relayer/gorelayer.sh && /home/relayer/gorelayer.sh",
			},
		},
		noRestart,
	)
	s.Require().NoError(err)
	s.T().Logf("✅ Started gorelayer container: %s", s.HermesResource.Container.ID)

	s.T().Logf("ℹ️ Waiting for ibc channel creation...")
	s.Require().Eventually(
		func() bool {
			s.T().Log("We are waiting for channel creation...")
			channels, err := s.QueryIBCChannels(s.UmeeREST())
			if channels {
				s.T().Log("✅ IBC Channel is created among the the chains")
			}
			if err != nil {
				return false
			}
			return channels
		},
		10*time.Minute,
		3*time.Second,
	)
}
