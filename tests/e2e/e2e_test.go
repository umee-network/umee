package e2e

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/tests/grpc"
)

func (s *IntegrationTestSuite) checkOutflowByPercentage(endpoint, excDenom string, outflow, amount, perDiff sdk.Dec) {
	// get historic average price for denom (SYMBOL_DENOM)
	histoAvgPrice, err := queryHistroAvgPrice(endpoint, excDenom)
	s.Require().NoError(err)
	powerReduction := sdk.MustNewDecFromStr("10").Power(6)
	totalPrice := amount.Quo(powerReduction).Mul(histoAvgPrice)
	s.T().Log("exchangeRate total price ", totalPrice.String(), "outflow value", outflow.String())
	percentageDiff := totalPrice.Mul(perDiff)
	// Note: checking outflow >= total_price with percentageDiff
	// either total_price >= outflow with percentageDiff
	s.Require().True(outflow.GTE(totalPrice.Sub(percentageDiff)) || totalPrice.GTE(outflow.Sub(percentageDiff)))
}

func (s *IntegrationTestSuite) checkOutflows(umeeAPIEndpoint, denom string, checkWithExcRate bool, amount sdk.Dec, excDenom string) {
	s.Require().Eventually(
		func() bool {
			outflows, err := queryOutflows(umeeAPIEndpoint, denom)
			s.Require().NoError(err)
			if checkWithExcRate {
				outflow := outflows.AmountOf(denom)
				s.checkOutflowByPercentage(umeeAPIEndpoint, excDenom, outflow, amount, sdk.MustNewDecFromStr("0.01"))
			}
			return outflows.Len() == 1 && outflows[0].Denom == denom
		},
		time.Minute,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) checkSupply(endpoint, ibcDenom string, amount math.Int) {
	s.Require().Eventually(
		func() bool {
			supply, err := queryTotalSupply(endpoint)
			s.Require().NoError(err)
			s.Require().Equal(supply.AmountOf(ibcDenom).Int64(), amount.Int64())
			return supply.AmountOf(ibcDenom).Equal(amount)
		},
		time.Minute,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	// s.T().Parallel()
	var ibcStakeDenom string

	s.Run("ibc_transfer_quota", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		gaiaAPIEndpoint := s.gaiaREST()
		umeeAPIEndpoint := s.umeeREST()
		// ibc hash of uumee token
		umeeIBCHash := "ibc/9F53D255F5320A4BE124FF20C29D46406E126CE8A09B00CA8D3CFF7905119728"
		// ibc hash of uatom token
		uatomIBCHash := "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"

		// send uatom from gaia to umee
		uatomAmount := sdk.NewInt64Coin("uatom", 100000000) // 100ATOM
		s.sendIBC(gaiaChainID, s.chain.id, "", uatomAmount)
		s.checkSupply(umeeAPIEndpoint, uatomIBCHash, uatomAmount.Amount)

		// TODO: needs to calculate the quota before ibc-transfer to check quota
		// sending more tokens than token_quota limit of umee
		// 120000 * 0.007863408515960442 => 943 $
		exceedAmount := sdk.NewInt64Coin(appparams.BondDenom, 120000000000) // 120000UMEE
		s.sendIBC(s.chain.id, gaiaChainID, "", exceedAmount)
		// check the ibc (umee) quota after ibc txs
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, math.ZeroInt())

		// send 100UMEE from umee to gaia
		// Note: receiver is null means hermes will default send to key_name (from config) of target chain (gaia)
		token := sdk.NewInt64Coin(appparams.BondDenom, 100000000) // 100UMEE
		s.sendIBC(s.chain.id, gaiaChainID, "", token)
		s.checkOutflows(umeeAPIEndpoint, appparams.BondDenom, true, sdk.NewDecFromInt(token.Amount), appparams.Name)
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, token.Amount)

		// send uatom (ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2) from umee to gaia
		uatomIBCToken := sdk.NewInt64Coin(uatomIBCHash, 100000000) // 100ATOM
		// supply will be not be decreased because sending uatomIBCToken amount is more than token quota so it will fail
		s.sendIBC(s.chain.id, gaiaChainID, "", uatomIBCToken)
		s.checkSupply(umeeAPIEndpoint, uatomIBCHash, uatomIBCToken.Amount)

		uatomIBCToken.Amount = math.NewInt(1000000) // 1 ATOM
		s.sendIBC(s.chain.id, gaiaChainID, "", uatomIBCToken)
		// remaing supply still exists for uatom in umee
		s.checkSupply(umeeAPIEndpoint, uatomIBCHash, uatomAmount.Amount.Sub(uatomIBCToken.Amount))
		s.checkOutflows(umeeAPIEndpoint, uatomIBCHash, true, sdk.NewDecFromInt(uatomIBCToken.Amount), "ATOM")

		// sending more tokens then token_quota limit of umee
		// 120000 * 0.007863408515960442 => 943 $
		s.sendIBC(s.chain.id, gaiaChainID, "", exceedAmount)
		// check the ibc (umee) supply after ibc txs, it will same as previous because it will fail because to quota limit exceed
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, token.Amount)

		// sending back some amount from receiver to sender (ibc/XXX)
		s.sendIBC(gaiaChainID, s.chain.id, "", sdk.NewInt64Coin(umeeIBCHash, 1000))
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, token.Amount.Sub(math.NewInt(1000)))

		// sending back remaining ibc amount from receiver to sender (ibc/XXX)
		s.sendIBC(gaiaChainID, s.chain.id, "", sdk.NewInt64Coin(umeeIBCHash, token.Amount.Sub(math.NewInt(1000)).Int64()))
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, math.ZeroInt())

		// reset the outflows
		s.T().Logf("waiting until quota reset, basically it will take around 300 seconds to do quota reset")
		s.Require().Eventually(
			func() bool {
				outflows, err := queryOutflows(umeeAPIEndpoint, appparams.BondDenom)
				s.Require().NoError(err)
				outflow := outflows.AmountOf(appparams.BondDenom)
				if outflow.Equal(sdk.NewDec(0)) {
					s.T().Logf("quota is reset : %s is %s ", appparams.BondDenom, outflow.String())
				}
				return outflow.Equal(sdk.NewDec(0))
			},
			5*time.Minute,
			5*time.Second,
		)

		// after reset sending again tokens from umee to gaia
		// send 100UMEE from umee to gaia
		// Note: receiver is null means hermes will default send to key_name (from config) of target chain (gaia)
		s.sendIBC(s.chain.id, gaiaChainID, "", token)
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, token.Amount)
	})

	// Non registered tokens (not exists in oracle for quota test)
	s.Run("send_stake_to_umee", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			balances sdk.Coins
			err      error
		)

		stakeIBCHash := "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
		umeeAPIEndpoint := s.umeeREST()

		valAddr, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		recipient := valAddr.String()
		token := sdk.NewInt64Coin("stake", 3300000000) // 3300stake
		s.sendIBC(gaiaChainID, s.chain.id, recipient, token)

		s.Require().Eventually(
			func() bool {
				balances, err = queryUmeeAllBalances(umeeAPIEndpoint, recipient)
				s.Require().NoError(err)
				return token.Amount.Equal(balances.AmountOf(stakeIBCHash))
			},
			time.Minute,
			5*time.Second,
		)
		s.checkSupply(umeeAPIEndpoint, stakeIBCHash, token.Amount)
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
