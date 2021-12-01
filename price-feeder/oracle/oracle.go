package oracle

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"

	rpcClient "github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neilotoole/errgroup"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/umee-network/umee/price-feeder/config"
	"github.com/umee-network/umee/price-feeder/oracle/client"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
	pfsync "github.com/umee-network/umee/price-feeder/pkg/sync"
	umeetypes "github.com/umee-network/umee/x/oracle/types"
	"google.golang.org/grpc"
)

type CurrencyPair struct {
	Base  string
	Quote string
}

func (cp CurrencyPair) String() string {
	return cp.Base + cp.Quote
}

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

type Oracle struct {
	logger             zerolog.Logger
	closer             *pfsync.Closer
	mtx                sync.RWMutex
	lastPriceSyncTS    time.Time
	prices             map[string]sdk.Dec
	oracleClient       *client.OracleClient
	providerPairs      map[string][]CurrencyPair
	previousPrevote    *PreviousPrevote
	previousVotePeriod float64
}

func New(oc *client.OracleClient, CurrencyPairs []config.CurrencyPair) *Oracle {
	providerPairs := make(map[string][]CurrencyPair)

	for _, pair := range CurrencyPairs {
		for _, provider := range pair.Providers {
			providerPairs[provider] = append(providerPairs[provider], CurrencyPair{
				Base:  pair.Base,
				Quote: pair.Quote,
			})
		}
	}

	return &Oracle{
		logger:          log.With().Str("module", "oracle").Logger(),
		closer:          pfsync.NewCloser(),
		oracleClient:    oc,
		providerPairs:   providerPairs,
		previousPrevote: nil,
	}
}

func (o *Oracle) Stop() {
	o.closer.Close()
	<-o.closer.Done()
}

// SetPrices retrieve all the prices from our set of providers as determined
// in the config, average them out, and update the oracle object
func (o *Oracle) SetPrices() error {
	g := new(errgroup.Group)
	mtx := new(sync.Mutex)
	providerPrices := make(map[string]map[string]sdk.Dec)

	for providerName, currencyPairs := range o.providerPairs {
		var priceProvider provider.Provider

		switch providerName {
		case config.ProviderBinance:
			priceProvider = provider.NewBinanceProvider()

		case config.ProviderKraken:
			priceProvider = provider.NewKrakenProvider()
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
			// e.g.: {ProviderKraken: {"ATOM": 34.03, "UMEE": 6.3}}
			mtx.Lock()
			for _, cp := range currencyPairs {
				if _, ok := providerPrices[providerName]; !ok {
					providerPrices[providerName] = make(map[string]sdk.Dec)
				}
				if p, ok := prices[cp.String()]; ok {
					providerPrices[providerName][cp.Base] = p
				}
			}
			mtx.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// consolidate the different provider maps into one for each exchange rate
	var (
		priceAverages = make(map[string]sdk.Dec)
		priceCounts   = make(map[string]int)
	)

	// TODO: Consider using Volume Weighted Average Price (VWAP).
	//
	// Ref: https://github.com/umee-network/umee/issues/251
	for _, prices := range providerPrices {
		for base, price := range prices {
			if _, ok := priceAverages[base]; !ok {
				priceAverages[base] = sdk.NewDec(0)
			}
			priceAverages[base] = priceAverages[base].Add(price)
			priceCounts[base]++
		}
	}

	for k, average := range priceAverages {
		average = average.QuoInt64(int64(priceCounts[k]))
	}

	o.prices = priceAverages

	return nil
}

func (o *Oracle) GetParams() (umeetypes.Params, error) {
	grpcConn, err := grpc.Dial(
		o.oracleClient.GRPCEndpoint,
		grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism.
	)
	if err != nil {
		return umeetypes.Params{}, err
	}

	defer grpcConn.Close()
	queryClient := umeetypes.NewQueryClient(grpcConn)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	queryResponse, err := queryClient.Params(ctx, &umeetypes.QueryParamsRequest{})
	if err != nil {
		return umeetypes.Params{}, err
	}

	return queryResponse.Params, nil
}

func (o *Oracle) generateSalt(length int) (string, error) {
	if length == 0 {
		return "", fmt.Errorf("failed to generate salt: zero length")
	}

	n := length / 2
	bytes := make([]byte, n)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func (o *Oracle) tick() error {
	ctx, err := o.oracleClient.CreateContext()
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

	blockHeight, err := rpcClient.GetChainHeight(ctx)
	if err != nil || blockHeight == 0 {
		return err
	}

	// Get oracle vote period, next block height,
	// Current vote period, index in vote period

	oracleVotePeriod := int64(oracleParams.VotePeriod)
	nextBlockHeight := blockHeight + 1
	currentVotePeriod := math.Floor(float64(nextBlockHeight) / float64(oracleVotePeriod))
	indexInVotePeriod := nextBlockHeight % oracleVotePeriod
	// Skip until new voting period
	// Skip when index [0, oracleVotePeriod - 1] is bigger than oracleVotePeriod - 2 or index is 0
	if (o.previousVotePeriod != 0 && currentVotePeriod == o.previousVotePeriod) ||
		oracleVotePeriod-indexInVotePeriod < 2 {
		return nil
	}

	// If we're past the voting period we needed to hit,
	// Reset and submit another pre-vote
	if o.previousVotePeriod != 0 && currentVotePeriod-o.previousVotePeriod != 1 {
		// Reset
		o.previousVotePeriod = 0
		o.previousPrevote = nil
		return nil
	}

	isPrevoteOnlyTx := o.previousPrevote == nil

	exchangeRates := ""

	for k, v := range o.prices {
		if exchangeRates != "" {
			exchangeRates = exchangeRates + "," + v.String() + k
		}
		if exchangeRates == "" {
			exchangeRates = v.String() + k
		}
	}

	salt, err := o.generateSalt(2)
	if err != nil {
		return err
	}

	valAddr, err := sdk.ValAddressFromBech32(o.oracleClient.ValidatorAddrString)
	if err != nil {
		return err
	}

	hash := umeetypes.GetAggregateVoteHash(salt, exchangeRates, o.oracleClient.ValidatorAddr)

	msg := &umeetypes.MsgAggregateExchangeRatePrevote{
		Hash:      hash.String(), // Hash of prices from the oracle
		Feeder:    o.oracleClient.OracleAddrString,
		Validator: valAddr.String(), //Hash accepts the actual addr
	}

	if isPrevoteOnlyTx {
		// Broadcast message

		err := o.oracleClient.BroadcastPrevote(msg)
		if err != nil {
			return err
		}

		currentHeight, err := rpcClient.GetChainHeight(ctx)
		if err != nil {
			return err
		}

		o.previousVotePeriod = math.Floor(float64(currentHeight) / float64(oracleVotePeriod))
		o.previousPrevote = &PreviousPrevote{
			Salt:              salt,
			ExchangeRates:     exchangeRates,
			SubmitBlockHeight: int64(currentHeight),
		}
	}

	// Is next voting period, vote

	if !isPrevoteOnlyTx {
		voteMsg := &umeetypes.MsgAggregateExchangeRateVote{
			Salt:          o.previousPrevote.Salt,
			ExchangeRates: o.previousPrevote.ExchangeRates,
			Feeder:        o.oracleClient.OracleAddrString,
			Validator:     valAddr.String(),
		}

		// Broadcast message

		err := o.oracleClient.BroadcastVote(nextBlockHeight,
			oracleVotePeriod-indexInVotePeriod,
			voteMsg)
		if err != nil {
			// This can happen if the voting is off-timed,
			// Or the voting denoms are not currently on whitelist.
			// We want to just reset and handle this silently :
			o.previousPrevote = nil
			o.previousVotePeriod = 0
			return nil
		}
		o.previousPrevote = nil
	}
	return nil
}

func (o *Oracle) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			o.closer.Close()

		default:
			err := o.tick()
			if err != nil {
				return err
			}
			o.lastPriceSyncTS = time.Now()
			time.Sleep(10 * time.Millisecond)
		}
	}
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
