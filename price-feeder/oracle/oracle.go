package oracle

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	rpcclient "github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/umee-network/umee/price-feeder/config"
	"github.com/umee-network/umee/price-feeder/oracle/client"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
	pfsync "github.com/umee-network/umee/price-feeder/pkg/sync"
	oracletypes "github.com/umee-network/umee/x/oracle/types"
)

// We define tickerTimeout as the minimum timeout between each oracle loop. We
// define this value empirically based on enough time to collect exchange rates,
// and broadcast pre-vote and vote transactions such that they're committed in a
// block during each voting period.
const (
	tickerTimeout = 1000 * time.Millisecond
)

// CurrencyPair defines a currency exchange pair consisting of a base and a quote.
// We primarily utilize the base for broadcasting exchange rates and use the
// pair for querying for the ticker prices.
type CurrencyPair struct {
	Base  string
	Quote string
}

// String implements the Stringer interface and defines a ticker symbol for
// querying the exchange rate.
func (cp CurrencyPair) String() string {
	return cp.Base + cp.Quote
}

// PreviousPrevote defines a structure for defining the previous prevote
// submitted on-chain.
type PreviousPrevote struct {
	ExchangeRates     string
	Salt              string
	SubmitBlockHeight int64
}

func NewPreviousPrevote() *PreviousPrevote {
	return &PreviousPrevote{
		Salt:              "",
		ExchangeRates:     "",
		SubmitBlockHeight: 0,
	}
}

// Oracle implements the core component responsible for fetching exchange rates
// for a given set of currency pairs and determining the correct exchange rates
// to submit to the on-chain price oracle adhering the oracle specification.
type Oracle struct {
	logger zerolog.Logger
	closer *pfsync.Closer

	providerPairs      map[string][]CurrencyPair
	previousPrevote    *PreviousPrevote
	previousVotePeriod float64
	priceProviders     map[string]provider.Provider
	oracleClient       client.OracleClient

	mtx             sync.RWMutex
	lastPriceSyncTS time.Time
	prices          map[string]sdk.Dec
}

func New(logger zerolog.Logger, oc client.OracleClient, currencyPairs []config.CurrencyPair) *Oracle {
	providerPairs := make(map[string][]CurrencyPair)

	for _, pair := range currencyPairs {
		for _, provider := range pair.Providers {
			providerPairs[provider] = append(providerPairs[provider], CurrencyPair{
				Base:  pair.Base,
				Quote: pair.Quote,
			})
		}
	}

	return &Oracle{
		logger:          logger.With().Str("module", "oracle").Logger(),
		closer:          pfsync.NewCloser(),
		oracleClient:    oc,
		providerPairs:   providerPairs,
		priceProviders:  make(map[string]provider.Provider),
		previousPrevote: nil,
	}
}

// Start starts the oracle process in a blocking fashion.
func (o *Oracle) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			o.closer.Close()

		default:
			o.logger.Debug().Msg("starting oracle tick")
			if err := o.tick(); err != nil {
				o.logger.Err(err).Msg("oracle tick failed")
			}

			o.lastPriceSyncTS = time.Now()

			time.Sleep(tickerTimeout)
		}
	}
}

// Stop stops the oracle process and waits for it to gracefully exit.
func (o *Oracle) Stop() {
	o.closer.Close()
	<-o.closer.Done()
}

// GetLastPriceSyncTimestamp returns the latest timestamp at which prices where
// fetched from the oracle's set of exchange rate providers.
func (o *Oracle) GetLastPriceSyncTimestamp() time.Time {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.lastPriceSyncTS
}

// GetPrices returns a copy of the current prices fetched from the oracle's
// set of exchange rate providers.
func (o *Oracle) GetPrices() map[string]sdk.Dec {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	// Creates a new array for the prices in the oracle
	prices := make(map[string]sdk.Dec, len(o.prices))
	for k, v := range o.prices {
		// Fills in the prices with each value in the oracle
		prices[k] = v
	}

	return prices
}

// SetPrices retrieve all the prices from our set of providers as determined
// in the config, average them out, and update the oracle's current exchange
// rates.
func (o *Oracle) SetPrices() error {
	g := new(errgroup.Group)
	mtx := new(sync.Mutex)
	providerPrices := make(map[string]map[string]provider.TickerPrice)

	for providerName, currencyPairs := range o.providerPairs {
		providerName := providerName
		currencyPairs := currencyPairs

		var (
			priceProvider provider.Provider
			ok            bool
		)

		priceProvider, ok = o.priceProviders[providerName]
		if !ok {
			switch providerName {
			case config.ProviderBinance:
				priceProvider = provider.NewBinanceProvider()

			case config.ProviderKraken:
				priceProvider = provider.NewKrakenProvider()
			}

			o.priceProviders[providerName] = priceProvider
		}

		g.Go(func() error {
			var ticker []string
			for _, cp := range currencyPairs {
				ticker = append(ticker, cp.String())
			}

			prices, err := priceProvider.GetTickerPrices(ticker...)
			if err != nil {
				return err
			}

			// flatten and collect prices based on the base currency per provider
			//
			// e.g.: {ProviderKraken: {"ATOM": <price, volume>, ...}}
			mtx.Lock()
			for _, cp := range currencyPairs {
				if _, ok := providerPrices[providerName]; !ok {
					providerPrices[providerName] = make(map[string]provider.TickerPrice)
				}
				if tp, ok := prices[cp.String()]; ok {
					providerPrices[providerName][cp.Base] = tp
				}
			}
			mtx.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	o.prices = ComputeVWAP(providerPrices)

	return nil
}

// GetParams returns the current on-chain parameters of the x/oracle module.
func (o *Oracle) GetParams() (oracletypes.Params, error) {
	grpcConn, err := grpc.Dial(
		o.oracleClient.GRPCEndpoint,
		// the Cosmos SDK doesn't support any transport security mechanism
		grpc.WithInsecure(),
	)
	if err != nil {
		return oracletypes.Params{}, err
	}

	defer grpcConn.Close()
	queryClient := oracletypes.NewQueryClient(grpcConn)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	queryResponse, err := queryClient.Params(ctx, &oracletypes.QueryParamsRequest{})
	if err != nil {
		return oracletypes.Params{}, err
	}

	return queryResponse.Params, nil
}

func (o *Oracle) tick() error {
	o.logger.Debug().Msg("executing oracle tick")

	clientCtx, err := o.oracleClient.CreateClientContext()
	if err != nil {
		return err
	}

	if err := o.SetPrices(); err != nil {
		return err
	}

	oracleParams, err := o.GetParams()
	if err != nil {
		return err
	}

	blockHeight, err := rpcclient.GetChainHeight(clientCtx)
	if err != nil {
		return nil
	}
	if blockHeight == 0 {
		return fmt.Errorf("expected non-zero block height")
	}

	// Get oracle vote period, next block height, current vote period, and index
	// in the vote period.
	oracleVotePeriod := int64(oracleParams.VotePeriod)
	nextBlockHeight := blockHeight + 1
	currentVotePeriod := math.Floor(float64(nextBlockHeight) / float64(oracleVotePeriod))
	indexInVotePeriod := nextBlockHeight % oracleVotePeriod

	// Skip until new voting period. Specifically, skip when:
	// index [0, oracleVotePeriod - 1] > oracleVotePeriod - 2 OR index is 0
	if (o.previousVotePeriod != 0 && currentVotePeriod == o.previousVotePeriod) ||
		oracleVotePeriod-indexInVotePeriod < 2 {
		o.logger.Info().
			Int64("vote_period", oracleVotePeriod).
			Float64("previous_vote_period", o.previousVotePeriod).
			Float64("current_vote_period", currentVotePeriod).
			Msg("skipping until next voting period")

		return nil
	}

	// If we're past the voting period we needed to hit, reset and submit another
	// prevote.
	if o.previousVotePeriod != 0 && currentVotePeriod-o.previousVotePeriod != 1 {
		o.logger.Info().
			Int64("vote_period", oracleVotePeriod).
			Float64("previous_vote_period", o.previousVotePeriod).
			Float64("current_vote_period", currentVotePeriod).
			Msg("missing vote during voting period")

		o.previousVotePeriod = 0
		o.previousPrevote = nil
		return nil
	}

	salt, err := GenerateSalt(2)
	if err != nil {
		return err
	}

	valAddr, err := sdk.ValAddressFromBech32(o.oracleClient.ValidatorAddrString)
	if err != nil {
		return err
	}

	exchangeRatesStr := GenerateExchangeRatesString(o.prices)
	hash := oracletypes.GetAggregateVoteHash(salt, exchangeRatesStr, valAddr)
	preVoteMsg := &oracletypes.MsgAggregateExchangeRatePrevote{
		Hash:      hash.String(), // hash of prices from the oracle
		Feeder:    o.oracleClient.OracleAddrString,
		Validator: valAddr.String(),
	}

	isPrevoteOnlyTx := o.previousPrevote == nil
	if isPrevoteOnlyTx {
		// This timeout could be as small as oracleVotePeriod-indexInVotePeriod,
		// but we give it some extra time just in case.
		//
		// Ref : https://github.com/terra-money/oracle-feeder/blob/baef2a4a02f57a2ffeaa207932b2e03d7fb0fb25/feeder/src/vote.ts#L222
		o.logger.Info().
			Str("hash", hash.String()).
			Str("validator", preVoteMsg.Validator).
			Str("feeder", preVoteMsg.Feeder).
			Msg("broadcasting pre-vote")
		if err := o.oracleClient.BroadcastTx(nextBlockHeight, oracleVotePeriod*2, preVoteMsg); err != nil {
			return err
		}

		currentHeight, err := rpcclient.GetChainHeight(clientCtx)
		if err != nil {
			return err
		}

		o.previousVotePeriod = math.Floor(float64(currentHeight) / float64(oracleVotePeriod))
		o.previousPrevote = &PreviousPrevote{
			Salt:              salt,
			ExchangeRates:     exchangeRatesStr,
			SubmitBlockHeight: currentHeight,
		}
	} else {
		// otherwise, we're in the next voting period and thus we vote
		voteMsg := &oracletypes.MsgAggregateExchangeRateVote{
			Salt:          o.previousPrevote.Salt,
			ExchangeRates: o.previousPrevote.ExchangeRates,
			Feeder:        o.oracleClient.OracleAddrString,
			Validator:     valAddr.String(),
		}

		o.logger.Info().
			Str("exchange_rates", voteMsg.ExchangeRates).
			Str("validator", voteMsg.Validator).
			Str("feeder", voteMsg.Feeder).
			Msg("broadcasting vote")
		if err := o.oracleClient.BroadcastTx(
			nextBlockHeight,
			oracleVotePeriod-indexInVotePeriod,
			voteMsg,
		); err != nil {
			return err
		}

		o.previousPrevote = nil
		o.previousVotePeriod = 0
	}

	return nil
}

// GenerateSalt generates a random salt, size length/2,  as a HEX encoded string.
func GenerateSalt(length int) (string, error) {
	if length == 0 {
		return "", fmt.Errorf("failed to generate salt: zero length")
	}

	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// GenerateExchangeRatesString generates a canonical string representation of
// the aggregated exchange rates.
func GenerateExchangeRatesString(prices map[string]sdk.Dec) string {
	exchangeRates := make([]string, len(prices))
	i := 0

	// aggregate exchange rates as "<base>:<price>"
	for base, avgPrice := range prices {
		exchangeRates[i] = fmt.Sprintf("%s:%s", base, avgPrice.String())
		i++
	}

	sort.Strings(exchangeRates)

	return strings.Join(exchangeRates, ",")
}
