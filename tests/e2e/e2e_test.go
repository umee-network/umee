package e2e

import (
	"fmt"
	"time"
)

func (s *IntegrationTestSuite) TestTokenTransfers() {
	// deploy umee ERC20 token contract
	// var umeeERC20Addr string
	s.Run("deploy_umee_erc20", func() {
		_ = s.deployERC20Token("uumee")
	})

	// deploy photon ERC20 token contact
	// var photonERC20Addr string
	s.Run("deploy_photon_erc20", func() {
		_ = s.deployERC20Token("photon")
	})

	// send 100 photon tokens from Umee to Ethereum
	s.Run("send_photon_tokens_to_eth", func() {
		s.sendFromUmeeToEth(0, s.chain.validators[1].ethereumKey.address, "100photon", "10photon", "3photon")

		endpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))
		fromAddr := s.chain.validators[0].keyInfo.GetAddress()

		// require the sender's (validator) balance decreased
		balance, err := queryUmeeDenomBalance(endpoint, fromAddr.String(), "photon")
		s.Require().NoError(err)
		s.Require().Equal(99999999887, balance)

		// TODO/XXX: Test checking Ethereum account balance. This might require
		// creating go bindings to the gravity contract. For now, we sleep enough
		// time for the orchestrator to relay the event to Ethereum.
		time.Sleep(30 * time.Second)
	})

	// TODO: Re-enable once https://github.com/PeggyJV/gravity-bridge/pull/123 is
	// merged and included in a release.
	//
	// Ref: https://github.com/umee-network/umee/issues/10
	//
	// send 100 photon tokens from Ethereum back to Umee
	// s.Run("send_photon_tokens_from_eth", func() {
	// 	s.sendFromEthToUmee(1, photonERC20Addr, s.chain.validators[0].keyInfo.GetAddress().String(), "100")
	// })

	// send 300 umee tokens from Umee to Ethereum
	s.Run("send_uumee_tokens_to_eth", func() {
		s.sendFromUmeeToEth(0, s.chain.validators[1].ethereumKey.address, "300uumee", "10photon", "7uumee")

		endpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))
		fromAddr := s.chain.validators[0].keyInfo.GetAddress()

		balance, err := queryUmeeDenomBalance(endpoint, fromAddr.String(), "uumee")
		s.Require().NoError(err)
		s.Require().Equal(9999999693, balance)

		// TODO/XXX: Test checking Ethereum account balance. This might require
		// creating go bindings to the gravity contract. For now, we sleep enough
		// time for the orchestrator to relay the event to Ethereum.
		time.Sleep(30 * time.Second)
	})

	// TODO: Re-enable once https://github.com/PeggyJV/gravity-bridge/pull/123 is
	// merged and included in a release.
	//
	// Ref: https://github.com/umee-network/umee/issues/10
	//
	// send 300 umee tokens Ethereum back to Umee
	// s.Run("send_uumee_tokens_from_eth", func() {
	// 	s.sendFromEthToUmee(1, umeeERC20Addr, s.chain.validators[0].keyInfo.GetAddress().String(), "300")
	// })
}
