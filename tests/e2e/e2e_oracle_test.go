package e2e

import (
	"time"

	"github.com/umee-network/umee/v6/tests/grpc"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
)

// TestAllPrices wairs for all tokens in the leverage registry to have prices.
func (s *E2ETest) TestAllPrices() {
	tokens := []leveragetypes.Token{}
	s.Require().Eventually(
		func() bool {
			var err error
			tokens, err = s.QueryRegisteredTokens(s.UmeeREST())
			if err != nil {
				return false
			}
			return true
		},
		time.Minute, time.Second, "get registered tokens",
	)
	s.Require().Eventually(
		func() bool {
			exchangeRates, err := s.QueryExchangeRate(s.UmeeREST(), "")
			if err != nil {
				return false
			}
			for _, t := range tokens {
				if exchangeRates.AmountOf(t.SymbolDenom).IsZero() {
					return false
				}
			}
			return true
		},
		time.Minute, time.Second, "ensure all tokens have prices",
	)
}

// TestMedians queries for the oracle params, collects historical
// prices based on those params, checks that the stored medians and
// medians deviations are correct, updates the oracle params with
// a gov prop, then checks the medians and median deviations again.
func (s *E2ETest) TestMedians() {
	err := grpc.MedianCheck(s.AccountClient(0))
	s.Require().NoError(err)
}

func (s *E2ETest) TestUpdateOracleParams() {
	params, err := s.AccountClient(0).QueryOracleParams()
	s.Require().NoError(err)

	s.Require().Equal(uint64(5), params.HistoricStampPeriod)
	s.Require().Equal(uint64(4), params.MaximumPriceStamps)
	s.Require().Equal(uint64(20), params.MedianStampPeriod)

	// simple retry loop to submit and pass a proposal
	for i := 0; i < 3; i++ {
		err = grpc.SubmitAndPassProposal(
			s.AccountClient(0),
			grpc.OracleParamChanges(10, 2, 20),
		)
		if err == nil {
			break
		}

		time.Sleep(1 * time.Second)
	}

	s.Require().NoError(err, "submit and pass proposal")

	params, err = s.AccountClient(0).QueryOracleParams()
	s.Require().NoError(err)

	s.Require().Equal(uint64(10), params.HistoricStampPeriod)
	s.Require().Equal(uint64(2), params.MaximumPriceStamps)
	s.Require().Equal(uint64(20), params.MedianStampPeriod)

	s.Require().NoError(err)
}
