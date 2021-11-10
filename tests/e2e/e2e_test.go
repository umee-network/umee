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

	s.Run("send_stake_to_umee", func() {
		recipient := s.chain.validators[0].keyInfo.GetAddress().String()
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
		ethRecipient := s.chain.validators[1].ethereumKey.address
		s.sendFromUmeeToEth(0, ethRecipient, fmt.Sprintf("300%s", ibcStakeDenom), "10photon", fmt.Sprintf("7%s", ibcStakeDenom))

		umeeAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))
		fromAddr := s.chain.validators[0].keyInfo.GetAddress()

		// NOTE: We have to query for all balances because currently querying by
		// IBC denomination will result in an API error due to the denomination not
		// being URI handled.
		var token sdk.Coin
		balances, err := queryUmeeAllBalances(umeeAPIEndpoint, fromAddr.String())
		s.Require().NoError(err)

		for _, c := range balances {
			if c.Denom == ibcStakeDenom {
				token = c
				break
			}
		}

		s.Require().Equal(ibcStakeDenom, token.Denom)
		s.Require().Equal(int64(3299999693), token.Amount.Int64())

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

				latestBalance = b

				// The balance could differ if the receiving address was the orchestrator
				// the sent the batch tx and got the gravity fee.
				return b >= 300 && b <= 307
			},
			2*time.Minute,
			5*time.Second,
			"unexpected balance: %d", latestBalance,
		)
	})

	// send 300 stake tokens from Ethereum back to Umee
	s.Run("send_stake_tokens_from_eth", func() {
		s.sendFromEthToUmee(1, ibcStakeERC20Addr, s.chain.validators[0].keyInfo.GetAddress().String(), "300")

		umeeAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))
		toAddr := s.chain.validators[0].keyInfo.GetAddress()
		expBalance := int64(3299999993)

		// require the original sender's (validator) balance increased
		var latestBalance int64
		s.Require().Eventuallyf(
			func() bool {
				// NOTE: We have to query for all balances because currently querying by
				// IBC denomination will result in an API error due to the denomination
				// not being URI handled.
				var token sdk.Coin
				balances, err := queryUmeeAllBalances(umeeAPIEndpoint, toAddr.String())
				s.Require().NoError(err)

				for _, c := range balances {
					if c.Denom == ibcStakeDenom {
						token = c
						break
					}
				}

				latestBalance = token.Amount.Int64()

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
		ethRecipient := s.chain.validators[1].ethereumKey.address
		s.sendFromUmeeToEth(0, ethRecipient, "100photon", "10photon", "3photon")

		umeeEndpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))
		fromAddr := s.chain.validators[0].keyInfo.GetAddress()

		// require the sender's (validator) balance decreased
		balance, err := queryUmeeDenomBalance(umeeEndpoint, fromAddr.String(), "photon")
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(balance.Amount.Int64(), int64(99999998403))

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

				latestBalance = b

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
		toAddr := s.chain.validators[0].keyInfo.GetAddress()
		s.sendFromEthToUmee(1, photonERC20Addr, toAddr.String(), "100")

		umeeEndpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))
		expBalance := int64(99999998500)

		// require the original sender's (validator) balance increased
		var latestBalance int64
		s.Require().Eventuallyf(
			func() bool {
				b, err := queryUmeeDenomBalance(umeeEndpoint, toAddr.String(), "photon")
				if err != nil {
					return false
				}

				latestBalance = b.Amount.Int64()

				return latestBalance >= expBalance
			},
			2*time.Minute,
			5*time.Second,
			"unexpected balance: %d", latestBalance,
		)
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
		ethRecipient := s.chain.validators[1].ethereumKey.address
		s.sendFromUmeeToEth(0, ethRecipient, "300uumee", "10photon", "7uumee")

		endpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))
		fromAddr := s.chain.validators[0].keyInfo.GetAddress()

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

				latestBalance = b

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
		toAddr := s.chain.validators[0].keyInfo.GetAddress()
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
