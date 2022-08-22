package e2e

import (
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	umeeapp "github.com/umee-network/umee/v2/app"
)

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	var ibcStakeDenom string

	valAddr, err := s.chain.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)

	s.Run("send_stake_to_umee", func() {
		recipient := valAddr.String()
		token := sdk.NewInt64Coin("stake", 3300000000) // 3300stake
		s.sendIBC(gaiaChainID, s.chain.id, recipient, token)

		umeeAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))

		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			balances sdk.Coins
			err      error
		)
		s.Require().Eventually(
			func() bool {
				balances, err = queryUmeeAllBalances(umeeAPIEndpoint, recipient)
				s.Require().NoError(err)

				return balances.Len() == 3
			},
			time.Minute,
			5*time.Second,
		)

		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				ibcStakeDenom = c.Denom
				s.Require().Equal(token.Amount.Int64(), c.Amount.Int64())
				break
			}
		}

		s.Require().NotEmpty(ibcStakeDenom)
	})

	var ibcStakeERC20Addr string
	s.Run("deploy_stake_erc20", func() {
		ibcStakeERC20Addr = s.deployERC20Token(ibcStakeDenom)
	})

	// send 300 stake tokens from Umee to Ethereum
	s.Run("send_stake_tokens_to_eth", func() {
		umeeValIdxSender := 0
		orchestratorIdxReceiver := 1
		amount := sdk.NewCoin(ibcStakeDenom, math.NewInt(300))
		umeeFee := sdk.NewCoin(photonDenom, math.NewInt(10))
		gravityFee := sdk.NewCoin(ibcStakeDenom, math.NewInt(7))

		s.sendFromUmeeToEthCheck(umeeValIdxSender, orchestratorIdxReceiver, ibcStakeERC20Addr, amount, umeeFee, gravityFee)
	})

	// send 300 stake tokens from Ethereum back to Umee
	s.Run("send_stake_tokens_from_eth", func() {
		umeeValIdxReceiver := 0
		orchestratorIdxSender := 1
		amount := uint64(300)

		s.sendFromEthToUmeeCheck(orchestratorIdxSender, umeeValIdxReceiver, ibcStakeERC20Addr, ibcStakeDenom, amount)
	})
}

func (s *IntegrationTestSuite) TestPhotonTokenTransfers() {
	// deploy photon ERC20 token contact
	var photonERC20Addr string
	s.Run("deploy_photon_erc20", func() {
		photonERC20Addr = s.deployERC20Token("photon")
	})

	// send 100 photon tokens from Umee to Ethereum
	s.Run("send_photon_tokens_to_eth", func() {
		umeeValIdxSender := 0
		orchestratorIdxReceiver := 1
		amount := sdk.NewCoin(photonDenom, math.NewInt(100))
		umeeFee := sdk.NewCoin(photonDenom, math.NewInt(10))
		gravityFee := sdk.NewCoin(photonDenom, math.NewInt(3))

		s.sendFromUmeeToEthCheck(umeeValIdxSender, orchestratorIdxReceiver, photonERC20Addr, amount, umeeFee, gravityFee)
	})

	// send 100 photon tokens from Ethereum back to Umee
	s.Run("send_photon_tokens_from_eth", func() {
		umeeValIdxReceiver := 0
		orchestratorIdxSender := 1
		amount := uint64(100)

		s.sendFromEthToUmeeCheck(orchestratorIdxSender, umeeValIdxReceiver, photonERC20Addr, photonDenom, amount)
	})
}

func (s *IntegrationTestSuite) TestUmeeTokenTransfers() {
	// deploy umee ERC20 token contract
	var umeeERC20Addr string
	s.Run("deploy_umee_erc20", func() {
		umeeERC20Addr = s.deployERC20Token("uumee")
	})

	// send 300 umee tokens from Umee to Ethereum
	s.Run("send_uumee_tokens_to_eth", func() {
		umeeValIdxSender := 0
		orchestratorIdxReceiver := 1
		amount := sdk.NewCoin(umeeapp.BondDenom, math.NewInt(300))
		umeeFee := sdk.NewCoin(photonDenom, math.NewInt(10))
		gravityFee := sdk.NewCoin(umeeapp.BondDenom, math.NewInt(7))

		s.sendFromUmeeToEthCheck(umeeValIdxSender, orchestratorIdxReceiver, umeeERC20Addr, amount, umeeFee, gravityFee)
	})

	// send 300 umee tokens from Ethereum back to Umee
	s.Run("send_uumee_tokens_from_eth", func() {
		umeeValIdxReceiver := 0
		orchestratorIdxSender := 1
		amount := uint64(300)

		s.sendFromEthToUmeeCheck(orchestratorIdxSender, umeeValIdxReceiver, umeeERC20Addr, umeeapp.BondDenom, amount)
	})
}
