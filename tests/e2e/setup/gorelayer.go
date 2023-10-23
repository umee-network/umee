package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/ory/dockertest/v3"
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

	c := exec.Command("cp", "-r", filepath.Join("./scripts/", "relayer"), rlyCfgPath)
	if err = c.Run(); err == nil {
		s.T().Log("rly config files copied from ", filepath.Join("./scripts/", "relayer"), " to ", rlyCfgPath)
	}

	s.GoRelayerResource, err = s.DkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       "umee-gaia-gorelayer",
			Repository: "ghcr.io/cosmos/relayer",
			Tag:        "v2.4.2",
			NetworkID:  s.DkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/home/relayer", rlyCfgPath),
			},
			User: "relayer",
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
				"/home/relayer/gorelayer.sh",
			},
		},
		noRestart,
	)

	s.Require().NoError(err)
	s.T().Logf("✅ Started gorelayer container: %s", s.GoRelayerResource.Container.ID)

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
		5*time.Minute,
		3*time.Second,
	)
	return
}
