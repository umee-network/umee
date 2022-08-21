package e2e

import (
	"context"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
		ethRecipient := s.chain.orchestrators[1].ethereumKey.address
		s.sendFromUmeeToEth(0, ethRecipient, fmt.Sprintf("300%s", ibcStakeDenom), "10photon", fmt.Sprintf("7%s", ibcStakeDenom))

		umeeAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))
		fromAddr, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)

		// require the sender's (validator) balance decreased
		balance, err := queryUmeeDenomBalance(umeeAPIEndpoint, fromAddr.String(), ibcStakeDenom)
		s.Require().NoError(err)
		s.Require().Equal(int64(3299999693), balance.Amount.Int64())

		// require the Ethereum recipient balance increased
		var latestBalance int
		s.Require().Eventuallyf(
			func() bool {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()

				b, err := queryEthTokenBalance(ctx, s.ethClient, ibcStakeERC20Addr, ethRecipient)
				if err != nil {
					return false
				}

				latestBalance = int(b)

				// The balance could differ if the receiving address was the orchestrator
				// the sent the batch tx and got the gravity fee.
				return b >= 300 && b <= 307
			},
			5*time.Minute,
			5*time.Second,
			"unexpected balance: %d", latestBalance,
		)
	})

	// send 300 stake tokens from Ethereum back to Umee
	s.Run("send_stake_tokens_from_eth", func() {
		valAddr, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)

		s.sendFromEthToUmee(1, ibcStakeERC20Addr, valAddr.String(), "300")

		umeeAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))
		toAddr := valAddr
		expBalance := int64(3299999993)

		// require the original sender's (validator) balance increased
		var latestBalance int64
		s.Require().Eventuallyf(
			func() bool {
				balance, err := queryUmeeDenomBalance(umeeAPIEndpoint, toAddr.String(), ibcStakeDenom)
				if err != nil {
					return false
				}

				latestBalance = balance.Amount.Int64()
				return latestBalance == expBalance
			},
			2*time.Minute,
			5*time.Second,
			"unexpected balance: %d", latestBalance,
		)
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
		valIndex := 0
		umeeEndpoint := fmt.Sprintf("http://%s", s.valResources[valIndex].GetHostPort("1317/tcp"))
		fromAddr, err := s.chain.validators[valIndex].keyInfo.GetAddress()
		s.Require().NoError(err)

		balanceBeforeSend, err := queryUmeeDenomBalance(umeeEndpoint, fromAddr.String(), photonDenom) // 99999998016
		s.Require().NoError(err)
		s.T().Logf(
			"Umee Balance of tokens validator; index: %d, addr: %s, amount: %s, denom: %s",
			valIndex, fromAddr.String(), balanceBeforeSend.String(), photonDenom,
		)

		amount, umeeFee, gravityFee := uint64(100), uint64(10), uint64(3)
		ethRecipient := s.chain.orchestrators[1].ethereumKey.address
		s.sendFromUmeeToEth(0, ethRecipient, photonAmount(amount), photonAmount(umeeFee), photonAmount(gravityFee))

		// require the sender's (validator) balance decreased
		balance, err := queryUmeeDenomBalance(umeeEndpoint, fromAddr.String(), photonDenom) // 99999997903
		s.Require().NoError(err)
		s.T().Logf(
			"Umee Balance of tokens validator; index: %d, addr: %s, amount: %s, denom: %s",
			valIndex, fromAddr.String(), balance.String(), photonDenom,
		)
		s.Require().Equal(balanceBeforeSend.Amount.SubRaw(int64(amount+umeeFee+gravityFee)).Int64(), balance.Amount.Int64())

		// require the Ethereum recipient balance increased
		var latestBalance int
		s.Require().Eventuallyf(
			func() bool {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()

				b, err := queryEthTokenBalance(ctx, s.ethClient, photonERC20Addr, ethRecipient)
				if err != nil {
					return false
				}

				latestBalance = int(b)

				// The balance could differ if the receiving address was the orchestrator
				// that sent the batch tx and got the gravity fee.
				return b >= 100 && b <= 103
			},
			2*time.Minute,
			5*time.Second,
			"unexpected balance: %d", latestBalance,
		)
	})

	// send 100 photon tokens from Ethereum back to Umee
	s.Run("send_photon_tokens_from_eth", func() {
		umeeValIdxReceiver := 0
		orchestratorIdxSender := 1
		amount := uint64(100)

		s.sendFromEthToUmeeCheck(orchestratorIdxSender, umeeValIdxReceiver, photonDenom, photonERC20Addr, amount)
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
		ethRecipient := s.chain.orchestrators[1].ethereumKey.address
		s.sendFromUmeeToEth(0, ethRecipient, "300uumee", "10photon", "7uumee")

		endpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))
		fromAddr, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)

		balance, err := queryUmeeDenomBalance(endpoint, fromAddr.String(), "uumee")
		s.Require().NoError(err)
		s.Require().Equal(int64(9999999693), balance.Amount.Int64())

		// require the Ethereum recipient balance increased
		var latestBalance int
		s.Require().Eventuallyf(
			func() bool {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()

				b, err := queryEthTokenBalance(ctx, s.ethClient, umeeERC20Addr, ethRecipient)
				if err != nil {
					return false
				}

				latestBalance = int(b)

				// The balance could differ if the receiving address was the orchestrator
				// that sent the batch tx and got the gravity fee.
				return b >= 300 && b <= 307
			},
			2*time.Minute,
			5*time.Second,
			"unexpected balance: %d", latestBalance,
		)
	})

	// send 300 umee tokens from Ethereum back to Umee
	s.Run("send_uumee_tokens_from_eth", func() {
		toAddr, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		s.sendFromEthToUmee(1, umeeERC20Addr, toAddr.String(), "300")

		umeeEndpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))
		expBalance := int64(9999999993)

		// require the original sender's (validator) balance increased
		var latestBalance int64
		s.Require().Eventuallyf(
			func() bool {
				b, err := queryUmeeDenomBalance(umeeEndpoint, toAddr.String(), "uumee")
				if err != nil {
					return false
				}

				latestBalance = b.Amount.Int64()

				return latestBalance == expBalance
			},
			2*time.Minute,
			5*time.Second,
			"unexpected balance: %d", latestBalance,
		)
	})
}
