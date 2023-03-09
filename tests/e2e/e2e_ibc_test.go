package e2e

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v4/app/params"
)

var (
	powerReduction = sdk.MustNewDecFromStr("10").Power(6)
)

func (s *IntegrationTestSuite) checkOutflowByPercentage(endpoint, excDenom string, outflow, amount, perDiff sdk.Dec) {
	// get historic average price for denom (SYMBOL_DENOM)
	histoAvgPrice, err := queryHistAvgPrice(endpoint, excDenom)
	s.Require().NoError(err)
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
			a, err := queryOutflows(umeeAPIEndpoint, denom)
			s.Require().NoError(err)
			if checkWithExcRate {
				s.checkOutflowByPercentage(umeeAPIEndpoint, excDenom, a, amount, sdk.MustNewDecFromStr("0.01"))
			}
			return true
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
		atomSymbol := "ATOM"
		umeeSymbol := "UMEE"
		totalQuota := math.NewInt(120)
		tokenQuota := math.NewInt(100)
		// ibc hash of uumee token
		umeeIBCHash := "ibc/9F53D255F5320A4BE124FF20C29D46406E126CE8A09B00CA8D3CFF7905119728"
		// ibc hash of uatom token
		uatomIBCHash := "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"

		// send uatom from gaia to umee
		// Note : gaia -> umee (ibc_quota will not check token limit)
		histoAvgPriceOfAtom, err := queryHistAvgPrice(umeeAPIEndpoint, atomSymbol)
		s.Require().NoError(err)
		emOfAtom := sdk.NewDecFromInt(totalQuota).Quo(histoAvgPriceOfAtom)
		uatomAmount := sdk.NewInt64Coin("uatom", emOfAtom.Mul(powerReduction).RoundInt64())
		s.sendIBC(gaiaChainID, s.chain.id, "", uatomAmount)
		s.checkSupply(umeeAPIEndpoint, uatomIBCHash, uatomAmount.Amount)

		// sending more tokens than token_quota limit of umee (token_quota is 100$)
		histoAvgPriceOfUmee, err := queryHistAvgPrice(umeeAPIEndpoint, umeeSymbol)
		s.Require().NoError(err)
		exceedAmountOfUmee := sdk.NewDecFromInt(totalQuota).Quo(histoAvgPriceOfUmee)
		s.T().Logf("sending %s amount %s more than %s", umeeSymbol, exceedAmountOfUmee.String(), totalQuota.String())
		exceedAmountCoin := sdk.NewInt64Coin(appparams.BondDenom, exceedAmountOfUmee.Mul(powerReduction).RoundInt64())
		s.sendIBC(s.chain.id, gaiaChainID, "", exceedAmountCoin)
		// check the ibc (umee) quota after ibc txs
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, math.ZeroInt())

		// send 100UMEE from umee to gaia
		// Note: receiver is null means hermes will default send to key_name (from config) of target chain (gaia)
		// umee -> gaia (ibc_quota will check)
		umeeInitialQuota := math.NewInt(90)
		belowTokenQuota := sdk.NewDecFromInt(umeeInitialQuota).Quo(histoAvgPriceOfUmee)
		s.T().Logf("sending %s amount %s less than token quota %s", "UMEE", belowTokenQuota.String(), tokenQuota.String())
		token := sdk.NewInt64Coin(appparams.BondDenom, belowTokenQuota.Mul(powerReduction).RoundInt64())
		s.sendIBC(s.chain.id, gaiaChainID, "", token)
		s.checkOutflows(umeeAPIEndpoint, appparams.BondDenom, true, sdk.NewDecFromInt(token.Amount), appparams.Name)
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, token.Amount)

		// send uatom (ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2) from umee to gaia
		uatomIBCToken := sdk.NewInt64Coin(uatomIBCHash, uatomAmount.Amount.Int64())
		// supply will be not be decreased because sending uatomIBCToken amount is more than token quota so it will fail
		s.sendIBC(s.chain.id, gaiaChainID, "", uatomIBCToken)
		s.checkSupply(umeeAPIEndpoint, uatomIBCHash, uatomIBCToken.Amount)

		// send uatom below the token quota
		/*
			umee -> gaia
			umee token_quot = 90$
			total_quota = 120$
		*/
		belowTokenQuotabutNotBelowTotalQuota := sdk.NewDecFromInt(math.NewInt(90)).Quo(histoAvgPriceOfAtom)
		uatomIBCToken.Amount = math.NewInt(belowTokenQuotabutNotBelowTotalQuota.Mul(powerReduction).RoundInt64())
		s.sendIBC(s.chain.id, gaiaChainID, "", uatomIBCToken)
		// supply will be not be decreased because sending more than total quota from umee to gaia
		s.checkSupply(umeeAPIEndpoint, uatomIBCHash, uatomAmount.Amount)
		// making sure below the total quota
		belowTokenQuotaInUSD := totalQuota.Sub(umeeInitialQuota).Sub(math.NewInt(2))
		belowTokenQuotaforAtom := sdk.NewDecFromInt(belowTokenQuotaInUSD).Quo(histoAvgPriceOfAtom)
		uatomIBCToken.Amount = math.NewInt(belowTokenQuotaforAtom.Mul(powerReduction).RoundInt64())
		s.sendIBC(s.chain.id, gaiaChainID, "", uatomIBCToken)
		// remaing supply still exists for uatom in umee
		s.checkSupply(umeeAPIEndpoint, uatomIBCHash, uatomAmount.Amount.Sub(uatomIBCToken.Amount))
		s.checkOutflows(umeeAPIEndpoint, uatomIBCHash, true, sdk.NewDecFromInt(uatomIBCToken.Amount), atomSymbol)

		// sending more tokens then token_quota limit of umee
		s.sendIBC(s.chain.id, gaiaChainID, "", exceedAmountCoin)
		// check the ibc (umee) supply after ibc txs, it will same as previous because it will fail because to quota limit exceed
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, token.Amount)

		/* sending back some amount from receiver to sender (ibc/XXX)
		gaia -> umee
		*/
		s.sendIBC(gaiaChainID, s.chain.id, "", sdk.NewInt64Coin(umeeIBCHash, 1000))
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, token.Amount.Sub(math.NewInt(1000)))
		// sending back remaining ibc amount from receiver to sender (ibc/XXX)
		s.sendIBC(gaiaChainID, s.chain.id, "", sdk.NewInt64Coin(umeeIBCHash, token.Amount.Sub(math.NewInt(1000)).Int64()))
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, math.ZeroInt())

		// reset the outflows
		s.T().Logf("waiting until quota reset, basically it will take around 300 seconds to do quota reset")
		s.Require().Eventually(
			func() bool {
				amount, err := queryOutflows(umeeAPIEndpoint, appparams.BondDenom)
				s.Require().NoError(err)
				if amount.IsZero() {
					s.T().Logf("quota is reset : %s is 0", appparams.BondDenom)
					return true
				}
				return false
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
