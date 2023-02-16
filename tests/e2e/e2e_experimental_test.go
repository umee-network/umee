//go:build experimental
// +build experimental

package e2e

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: we need to check once HistoricAvgPrice is stored
func (s *IntegrationTestSuite) TestExperimentalIBCTokenTransfer() {
	valAddr, err := s.chain.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)

	s.Run("ibc_txs_for_umee", func() {
		recipient := valAddr.String()
		token := sdk.NewInt64Coin("uumee", 100000000) // 100UMEE
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
				s.Require().NotEmpty(c.Denom)
				s.Require().Equal(token.Amount.Int64(), c.Amount.Int64())
				break
			}
		}
	})
}
