package e2e

import (
	"fmt"
)

func (s *IntegrationTestSuite) TestTokenTransfers() {
	// deploy ERC20 umee token contract
	var umeeERC20Addr string
	s.Run("deploy_umee_erc20", func() {
		umeeERC20Addr = s.deployERC20Token("uumee", "umee", "umee", 6)
	})

	fmt.Println("UMEE ERC20 CONTRACT ADDR:", umeeERC20Addr)

	// deploy ERC20 photon token contact
	var photonERC20Addr string
	s.Run("deploy_umee_erc20", func() {
		photonERC20Addr = s.deployERC20Token("photon", "photon", "photon", 0)
	})

	fmt.Println("PHOTON ERC20 CONTRACT ADDR:", photonERC20Addr)

	// 3. Send photon Umee -> Ethereum
	// 4. Send photon Ethereum -> Umee
	// 5. Send umee Umee -> Ethereum
	// 6. Send umee Ethereum -> Umee
}
