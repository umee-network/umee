package oracle

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"github.com/umee-network/umee/price-feeder/config"
	"github.com/umee-network/umee/price-feeder/oracle/client"
)

type OracleTestSuite struct {
	suite.Suite

	oracle *Oracle
}

// SetupSuite executes once before the suite's tests are executed.
func (ots *OracleTestSuite) SetupSuite() {
	ots.oracle = New(
		zerolog.Nop(),
		client.OracleClient{},
		[]config.CurrencyPair{
			{
				Base:      "UMEE",
				Quote:     "USD",
				Providers: []string{config.ProviderBinance, config.ProviderKraken},
			},
		},
	)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OracleTestSuite))
}

func (ots *OracleTestSuite) TestStop() {
	ots.Eventually(
		func() bool {
			ots.oracle.Stop()
			return true
		},
		5*time.Second,
		time.Second,
	)
}

func (ots *OracleTestSuite) TestGetLastPriceSyncTimestamp() {
	// when no tick() has been invoked, assume zero value
	ots.Require().Equal(time.Time{}, ots.oracle.GetLastPriceSyncTimestamp())
}

func (ots *OracleTestSuite) TestPrices() {
	// initial prices should be empty (not set)
	ots.Require().Empty(ots.oracle.GetPrices())
}
