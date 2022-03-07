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
	"github.com/umee-network/umee/price-feeder/oracle/types"
	pfsync "github.com/umee-network/umee/price-feeder/pkg/sync"
	"github.com/umee-network/umee/price-feeder/telemetry"
	oracletypes "github.com/umee-network/umee/x/oracle/types"
)

// We define tickerTimeout as the minimum timeout between each oracle loop. We
// define this value empirically based on enough time to collect exchange rates,
// and broadcast pre-vote and vote transactions such that they're committed in a
// block during each voting period.
const (
	tickerTimeout = 1000 * time.Millisecond
)

// deviationThreshold defines how many 𝜎 a provider can be away from the mean
// without being considered faulty.
var deviationThreshold = sdk.MustNewDecFromStr("2")

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

	providerPairs      map[string][]types.CurrencyPair
	previousPrevote    *PreviousPrevote
	previousVotePeriod float64
	priceProviders     map[string]provider.Provider
	oracleClient       client.OracleClient

	mtx             sync.RWMutex
	lastPriceSyncTS time.Time
	prices          map[string]sdk.Dec
}

func New(logger zerolog.Logger, oc client.OracleClient, currencyPairs []config.CurrencyPair) *Oracle {
	providerPairs := make(map[string][]types.CurrencyPair)

	for _, pair := range currencyPairs {
		for _, provider := range pair.Providers {
			providerPairs[provider] = append(providerPairs[provider], types.CurrencyPair{
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

			startTime := time.Now()

			if err := o.tick(ctx); err != nil {
				telemetry.IncrCounter(1, "failure", "tick")
				o.logger.Err(err).Msg("oracle tick failed")
			}

			o.lastPriceSyncTS = time.Now()

			telemetry.MeasureSince(startTime, "runtime", "tick")
			telemetry.IncrCounter(1, "new", "tick")

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

// SetPrices retrieves all the prices and candles from our set of providers as
// determined in the config. If candles are available, uses TVWAP in order
// to determine prices. If candles are not available, uses the most recent prices
// with VWAP. Warns the the user of any missing prices, and filters out any faulty
// providers which do not report prices or candles within 2𝜎 of the others.
func (o *Oracle) SetPrices(ctx context.Context, acceptList oracletypes.DenomList) error {
	g := new(errgroup.Group)
	mtx := new(sync.Mutex)
	providerPrices := make(provider.AggregatedProviderPrices)
	providerCandles := make(provider.AggregatedProviderCandles)
	requiredRates := make(map[string]struct{})

	for providerName, currencyPairs := range o.providerPairs {
		providerName := providerName
		currencyPairs := currencyPairs

		priceProvider, err := o.getOrSetProvider(ctx, providerName)
		if err != nil {
			return err
		}

		var acceptedPairs []types.CurrencyPair
		for _, pair := range currencyPairs {
			if acceptList.Contains(pair.Base) {
				acceptedPairs = append(acceptedPairs, pair)
				if _, ok := requiredRates[pair.Base]; !ok {
					requiredRates[pair.Base] = struct{}{}
				}
			} else {
				o.logger.Warn().Str("denom", pair.Base).Msg("attempting to vote on unaccepted denom")
			}
		}

		g.Go(func() error {
			prices, err := priceProvider.GetTickerPrices(acceptedPairs...)
			if err != nil {
				telemetry.IncrCounter(1, "failure", "provider")
				return err
			}

			candles, err := priceProvider.GetCandlePrices(acceptedPairs...)
			if err != nil {
				telemetry.IncrCounter(1, "failure", "provider")
				return err
			}

			// flatten and collect prices based on the base currency per provider
			//
			// e.g.: {ProviderKraken: {"ATOM": <price, volume>, ...}}
			mtx.Lock()
			for _, pair := range acceptedPairs {
				if _, ok := providerPrices[providerName]; !ok {
					providerPrices[providerName] = make(map[string]provider.TickerPrice)
				}
				if _, ok := providerCandles[providerName]; !ok {
					providerCandles[providerName] = make(map[string][]provider.CandlePrice)
				}

				tp, pricesOk := prices[pair.String()]
				cp, candlesOk := candles[pair.String()]
				if pricesOk {
					providerPrices[providerName][pair.Base] = tp
				}
				if candlesOk {
					providerCandles[providerName][pair.Base] = cp
				}

				if !pricesOk && !candlesOk {
					mtx.Unlock()
					return fmt.Errorf("failed to find any exchange rates in provider response")
				}
			}
			mtx.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		o.logger.Debug().Err(err).Msg("failed to get ticker prices from provider")
	}

	reportedRates := make(map[string]struct{})
	for _, providers := range providerPrices {
		for base := range providers {
			if _, ok := reportedRates[base]; !ok {
				reportedRates[base] = struct{}{}
			}
		}
	}

	// warn the user of any missing prices
	if len(reportedRates) != len(requiredRates) {
		return fmt.Errorf("unable to get prices for all exchange rates")
	}
	for base := range requiredRates {
		if _, ok := reportedRates[base]; !ok {
			return fmt.Errorf("reported rates were not equal to required rates")
		}
	}

	filteredCandles, err := o.filterCandleDeviations(providerCandles)
	if err != nil {
		return err
	}

	// attempt to use candles for tvwap calculations
	tvwapPrices, err := ComputeTVWAP(filteredCandles)
	if err != nil {
		return err
	}

	// If TVWAP candles are not available or were filtered out due to staleness,
	// use most recent prices & VWAP instead.
	if len(tvwapPrices) == 0 {
		filteredProviderPrices, err := o.filterTickerDeviations(providerPrices)
		if err != nil {
			return err
		}

		vwapPrices, err := ComputeVWAP(filteredProviderPrices)
		if err != nil {
			return err
		}

		o.prices = vwapPrices
	} else {
		o.prices = tvwapPrices
	}

	return nil
}

// GetParams returns the current on-chain parameters of the x/oracle module.
func (o *Oracle) GetParams() (oracletypes.Params, error) {
	grpcConn, err := grpc.Dial(
		o.oracleClient.GRPCEndpoint,
		// the Cosmos SDK doesn't support any transport security mechanism
		grpc.WithInsecure(),
		grpc.WithContextDialer(dialerFunc),
	)
	if err != nil {
		return oracletypes.Params{}, fmt.Errorf("failed to dial Cosmos gRPC service: %w", err)
	}

	defer grpcConn.Close()
	queryClient := oracletypes.NewQueryClient(grpcConn)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	queryResponse, err := queryClient.Params(ctx, &oracletypes.QueryParamsRequest{})
	if err != nil {
		return oracletypes.Params{}, fmt.Errorf("failed to get x/oracle params: %w", err)
	}

	return queryResponse.Params, nil
}

func (o *Oracle) getOrSetProvider(ctx context.Context, providerName string) (provider.Provider, error) {
	var (
		priceProvider provider.Provider
		ok            bool
	)

	priceProvider, ok = o.priceProviders[providerName]
	if !ok {
		switch providerName {
		case config.ProviderBinance:
			binanceProvider, err := provider.NewBinanceProvider(ctx, o.logger, o.providerPairs[config.ProviderBinance]...)
			if err != nil {
				return nil, err
			}
			priceProvider = binanceProvider

		case config.ProviderKraken:
			krakenProvider, err := provider.NewKrakenProvider(ctx, o.logger, o.providerPairs[config.ProviderKraken]...)
			if err != nil {
				return nil, err
			}
			priceProvider = krakenProvider

		case config.ProviderOsmosis:
			priceProvider = provider.NewOsmosisProvider()

		case config.ProviderHuobi:
			huobiProvider, err := provider.NewHuobiProvider(ctx, o.logger, o.providerPairs[config.ProviderHuobi]...)
			if err != nil {
				return nil, err
			}
			priceProvider = huobiProvider

		case config.ProviderGate:
			gateProvider, err := provider.NewGateProvider(ctx, o.logger, o.providerPairs[config.ProviderGate]...)
			if err != nil {
				return nil, err
			}
			priceProvider = gateProvider

		case config.ProviderOkx:
			okxProvider, err := provider.NewOkxProvider(ctx, o.logger, o.providerPairs[config.ProviderOkx]...)
			if err != nil {
				return nil, err
			}
			priceProvider = okxProvider

		case config.ProviderMock:
			priceProvider = provider.NewMockProvider()
		}

		o.priceProviders[providerName] = priceProvider
	}

	return priceProvider, nil
}

// filterTickerDeviations finds the standard deviations of the prices of
// all assets, and filters out any providers that are not within 2𝜎 of the mean.
func (o *Oracle) filterTickerDeviations(
	prices provider.AggregatedProviderPrices,
) (provider.AggregatedProviderPrices, error) {
	var (
		filteredPrices = make(provider.AggregatedProviderPrices)
		priceMap       = make(map[string]map[string]sdk.Dec)
	)

	for providerName, providerPrices := range prices {
		if _, ok := priceMap[providerName]; !ok {
			priceMap[providerName] = make(map[string]sdk.Dec)
		}
		for base, price := range providerPrices {
			priceMap[providerName][base] = price.Price
		}
	}

	deviations, means, err := StandardDeviation(priceMap)
	if err != nil {
		return nil, err
	}

	// accept any prices that are within 2𝜎, or for which we couldn't get 𝜎
	for providerName, priceTickers := range prices {
		for base, ticker := range priceTickers {
			if _, ok := deviations[base]; !ok ||
				(ticker.Price.GTE(means[base].Sub(deviations[base].Mul(deviationThreshold))) &&
					ticker.Price.LTE(means[base].Add(deviations[base].Mul(deviationThreshold)))) {
				if _, ok := filteredPrices[providerName]; !ok {
					filteredPrices[providerName] = make(map[string]provider.TickerPrice)
				}
				filteredPrices[providerName][base] = ticker
			} else {
				telemetry.IncrCounter(1, "failure", "provider")
				o.logger.Warn().Str("base", base).Str("provider", providerName).Str(
					"price", ticker.Price.String()).Msg("provider deviating from other prices")
			}
		}
	}

	return filteredPrices, nil
}

// filterCandleDeviations finds the standard deviations of the tvwaps of
// all assets, and filters out any providers that are not within 2𝜎 of the mean.
func (o *Oracle) filterCandleDeviations(
	candles provider.AggregatedProviderCandles,
) (provider.AggregatedProviderCandles, error) {
	var (
		filteredCandles = make(provider.AggregatedProviderCandles)
		tvwaps          = make(map[string]map[string]sdk.Dec)
	)

	for providerName, c := range candles {
		candlePrices := make(provider.AggregatedProviderCandles)

		for assetName, asset := range c {
			if _, ok := candlePrices[providerName]; !ok {
				candlePrices[providerName] = make(map[string][]provider.CandlePrice)
			}
			candlePrices[providerName][assetName] = asset
		}

		tvwap, err := ComputeTVWAP(candlePrices)
		if err != nil {
			return nil, err
		}

		for assetName, asset := range tvwap {
			if _, ok := tvwaps[providerName]; !ok {
				tvwaps[providerName] = make(map[string]sdk.Dec)
			}
			tvwaps[providerName][assetName] = asset
		}
	}

	deviations, means, err := StandardDeviation(tvwaps)
	if err != nil {
		return nil, err
	}

	// accept any tvwaps that are within 2𝜎, or for which we couldn't get 𝜎
	for providerName, priceMap := range tvwaps {
		for base, price := range priceMap {
			if _, ok := deviations[base]; !ok ||
				(price.GTE(means[base].Sub(deviations[base].Mul(deviationThreshold))) &&
					price.LTE(means[base].Add(deviations[base].Mul(deviationThreshold)))) {
				if _, ok := filteredCandles[providerName]; !ok {
					filteredCandles[providerName] = make(map[string][]provider.CandlePrice)
				}
				filteredCandles[providerName][base] = candles[providerName][base]
			} else {
				telemetry.IncrCounter(1, "failure", "provider")
				o.logger.Warn().Str("base", base).Str("provider", providerName).Str(
					"price", price.String()).Msg("provider deviating from other candles")
			}
		}
	}

	return filteredCandles, nil
}

func (o *Oracle) checkAcceptList(params oracletypes.Params) {
	for _, denom := range params.AcceptList {
		symbol := strings.ToUpper(denom.SymbolDenom)
		if _, ok := o.prices[symbol]; !ok {
			o.logger.Warn().Str("denom", symbol).Msg("price missing for required denom")
		}
	}
}

func (o *Oracle) tick(ctx context.Context) error {
	o.logger.Debug().Msg("executing oracle tick")

	clientCtx, err := o.oracleClient.CreateClientContext()
	if err != nil {
		return err
	}
	oracleParams, err := o.GetParams()
	if err != nil {
		return err
	}
	if err := o.SetPrices(ctx, oracleParams.AcceptList); err != nil {
		return err
	}

	o.checkAcceptList(oracleParams)

	blockHeight, err := rpcclient.GetChainHeight(clientCtx)
	if err != nil {
		return err
	}
	if blockHeight < 1 {
		return fmt.Errorf("expected positive block height")
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

	salt, err := GenerateSalt(32)
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
