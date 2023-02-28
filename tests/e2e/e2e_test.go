package e2e

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/tests/grpc"
)

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	var ibcStakeDenom string

	s.Run("send_stake_to_umee", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			balances sdk.Coins
			err      error
		)

		valAddr, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		recipient := valAddr.String()
		token := sdk.NewInt64Coin("stake", 3300000000) // 3300stake
		s.sendIBC(gaiaChainID, s.chain.id, recipient, token)

		umeeAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))

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
			// C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878 = denom-hash(transfer/channe-0/stake)
			if c.Denom == "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878" {
				ibcStakeDenom = c.Denom
				s.Require().Equal(token.Amount.Int64(), c.Amount.Int64())
				break
			}
		}

		s.Require().NotEmpty(ibcStakeDenom)
	})

	s.Run("send_umee_to_gaia", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			supply sdk.Coins
			err    error
		)

		s.Require().NoError(err)
		token := sdk.NewInt64Coin(appparams.BondDenom, 100000000) // 100UMEE
		// send 100UMEE to umee to gaia
		// Note: receiver is null means hermes will default send to key_name (from config)
		s.sendIBC(s.chain.id, gaiaChainID, "", token)

		gaiaAPIEndpoint := fmt.Sprintf("http://%s", s.gaiaResource.GetHostPort("1317/tcp"))

		s.Require().Eventually(
			func() bool {
				supply, err = queryTotalSupply(gaiaAPIEndpoint)
				s.Require().NoError(err)
				return supply.Len() == 2
			},
			time.Minute,
			5*time.Second,
		)

		for _, c := range supply {
			// 9F53D255F5320A4BE124FF20C29D46406E126CE8A09B00CA8D3CFF7905119728 = denom-hash(transfer/channe-0/uumee)
			if c.Denom == "ibc/9F53D255F5320A4BE124FF20C29D46406E126CE8A09B00CA8D3CFF7905119728" {
				s.Require().Equal(token.Amount.Int64(), c.Amount.Int64())
				break
			}
		}
	})

	var ibcStakeERC20Addr string
	s.Run("deploy_stake_erc20 ibcStakeERC20Addr", func() {
		s.T().Skip("paused due to Ethereum PoS migration and PoW fork")
		s.Require().NotEmpty(ibcStakeDenom)
		ibcStakeERC20Addr = s.deployERC20Token(ibcStakeDenom)
	})

	// send 300 stake tokens from Umee to Ethereum
	s.Run("send_stake_tokens_to_eth", func() {
		s.T().Skip("paused due to Ethereum PoS migration and PoW fork")
		umeeValIdxSender := 0
		orchestratorIdxReceiver := 1
		amount := sdk.NewCoin(ibcStakeDenom, math.NewInt(300))
		umeeFee := sdk.NewCoin(appparams.BondDenom, math.NewInt(10000))
		gravityFee := sdk.NewCoin(ibcStakeDenom, math.NewInt(7))

		s.sendFromUmeeToEthCheck(umeeValIdxSender, orchestratorIdxReceiver, ibcStakeERC20Addr, amount, umeeFee, gravityFee)
	})

	// send 300 stake tokens from Ethereum back to Umee
	s.Run("send_stake_tokens_from_eth", func() {
		s.T().Skip("paused due to Ethereum PoS migration and PoW fork")
		umeeValIdxReceiver := 0
		orchestratorIdxSender := 1
		amount := uint64(300)

		s.sendFromEthToUmeeCheck(orchestratorIdxSender, umeeValIdxReceiver, ibcStakeERC20Addr, ibcStakeDenom, amount)
	})
}

func (s *IntegrationTestSuite) TestPhotonTokenTransfers() {
	s.T().Skip("paused due to Ethereum PoS migration and PoW fork")
	// deploy photon ERC20 token contact
	var photonERC20Addr string
	s.Run("deploy_photon_erc20", func() {
		photonERC20Addr = s.deployERC20Token(photonDenom)
	})

	// send 100 photon tokens from Umee to Ethereum
	s.Run("send_photon_tokens_to_eth", func() {
		umeeValIdxSender := 0
		orchestratorIdxReceiver := 1
		amount := sdk.NewCoin(photonDenom, math.NewInt(100))
		umeeFee := sdk.NewCoin(appparams.BondDenom, math.NewInt(10000))
		gravityFee := sdk.NewCoin(photonDenom, math.NewInt(3))

		s.sendFromUmeeToEthCheck(umeeValIdxSender, orchestratorIdxReceiver, photonERC20Addr, amount, umeeFee, gravityFee)
	})

	// send 100 photon tokens from Ethereum back to Umee
	s.Run("send_photon_tokens_from_eth", func() {
		s.T().Skip("paused due to Ethereum PoS migration and PoW fork")
		umeeValIdxReceiver := 0
		orchestratorIdxSender := 1
		amount := uint64(100)

		s.sendFromEthToUmeeCheck(orchestratorIdxSender, umeeValIdxReceiver, photonERC20Addr, photonDenom, amount)
	})
}

func (s *IntegrationTestSuite) TestUmeeTokenTransfers() {
	s.T().Skip("paused due to Ethereum PoS migration and PoW fork")
	// deploy umee ERC20 token contract
	var umeeERC20Addr string
	s.Run("deploy_umee_erc20", func() {
		umeeERC20Addr = s.deployERC20Token(appparams.BondDenom)
	})

	// send 300 umee tokens from Umee to Ethereum
	s.Run("send_uumee_tokens_to_eth", func() {
		umeeValIdxSender := 0
		orchestratorIdxReceiver := 1
		amount := sdk.NewCoin(appparams.BondDenom, math.NewInt(300))
		umeeFee := sdk.NewCoin(appparams.BondDenom, math.NewInt(10000))
		gravityFee := sdk.NewCoin(appparams.BondDenom, math.NewInt(7))

		s.sendFromUmeeToEthCheck(umeeValIdxSender, orchestratorIdxReceiver, umeeERC20Addr, amount, umeeFee, gravityFee)
	})

	// send 300 umee tokens from Ethereum back to Umee
	s.Run("send_uumee_tokens_from_eth", func() {
		umeeValIdxReceiver := 0
		orchestratorIdxSender := 1
		amount := uint64(300)

		s.sendFromEthToUmeeCheck(orchestratorIdxSender, umeeValIdxReceiver, umeeERC20Addr, appparams.BondDenom, amount)
	})
}

// TestMedians queries for the oracle params, collects historical
// prices based on those params, checks that the stored medians and
// medians deviations are correct, updates the oracle params with
// a gov prop, then checks the medians and median deviations again.
func (s *IntegrationTestSuite) TestMedians() {
	err := grpc.MedianCheck(s.umeeClient)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestUpdateOracleParams() {
	s.T().Skip("paused due to validator power threshold enforcing")
	params, err := s.umeeClient.QueryClient.QueryParams()
	s.Require().NoError(err)

	s.Require().Equal(uint64(5), params.HistoricStampPeriod)
	s.Require().Equal(uint64(4), params.MaximumPriceStamps)
	s.Require().Equal(uint64(20), params.MedianStampPeriod)

	err = grpc.SubmitAndPassProposal(
		s.umeeClient,
		grpc.OracleParamChanges(10, 2, 20),
	)
	s.Require().NoError(err)

	params, err = s.umeeClient.QueryClient.QueryParams()
	s.Require().NoError(err)

	s.Require().Equal(uint64(10), params.HistoricStampPeriod)
	s.Require().Equal(uint64(2), params.MaximumPriceStamps)
	s.Require().Equal(uint64(20), params.MedianStampPeriod)

	s.Require().NoError(err)
}
