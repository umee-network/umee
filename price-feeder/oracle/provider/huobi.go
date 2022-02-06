package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/price-feeder/oracle/types"
)

const (
	huobiBaseURL        = "https://api.huobi.pro"
	huobiTradeEndpoint  = "/market/trade"
	huobiMarketEndpoint = "/market/detail"
)

var _ Provider = (*HuobiProvider)(nil)

type (
	// HuobiProvider defines an Oracle provider implemented by the Huobi public
	// API.
	//
	// REF: https://huobiapi.github.io/docs/spot/v1/en/#market-data
	HuobiProvider struct {
		baseURL string
		client  *http.Client
	}

	// HuobiTraceResponse defines the response type for the last executed trade
	// for a given ticker/symbol.
	HuobiTraceResponse struct {
		Status string `json:"status"`
		ErrMsg string `json:"err-msg"`
		Tick   struct {
			Data []struct {
				Price float64 `json:"price"`
			} `json:"data"`
		} `json:"tick"`
	}

	// HuobiMarketResponse defines the response type for the last 24h market summary
	// for a given ticker/symbol.
	HuobiMarketResponse struct {
		Status string `json:"status"`
		ErrMsg string `json:"err-msg"`
		Tick   struct {
			Vol float64 `json:"vol"`
		} `json:"tick"`
	}
)

func NewHuobiProvider() *HuobiProvider {
	return &HuobiProvider{
		baseURL: huobiBaseURL,
		client:  newDefaultHTTPClient(),
	}
}

func NewHuobiProviderWithTimeout(timeout time.Duration) *HuobiProvider {
	return &HuobiProvider{
		baseURL: huobiBaseURL,
		client:  newHTTPClientWithTimeout(timeout),
	}
}

func (p HuobiProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]TickerPrice, error) {
	tickerPrices := make(map[string]TickerPrice, len(pairs))
	for _, cp := range pairs {
		price, err := p.getTickerPrice(cp.String())
		if err != nil {
			return nil, err
		}

		tickerPrices[cp.String()] = price
	}

	return tickerPrices, nil
}

func (p HuobiProvider) getTickerPrice(ticker string) (TickerPrice, error) {
	path := fmt.Sprintf("%s%s?symbol=%s", p.baseURL, huobiTradeEndpoint, strings.ToLower(ticker))

	resp, err := p.client.Get(path)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to make Huobi request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to read Huobi response body: %w", err)
	}

	var tradeResp HuobiTraceResponse
	if err := json.Unmarshal(bz, &tradeResp); err != nil {
		return TickerPrice{}, fmt.Errorf("failed to unmarshal Huobi response body: %w", err)
	}

	if !strings.EqualFold(tradeResp.Status, "ok") {
		return TickerPrice{}, fmt.Errorf(
			"received unexpected error from Huobi response: %s",
			tradeResp.ErrMsg,
		)
	}

	if len(tradeResp.Tick.Data) == 0 {
		return TickerPrice{}, fmt.Errorf("latest Huobi trade empty for %s", ticker)
	}

	rawPrice := tradeResp.Tick.Data[0].Price
	price, err := sdk.NewDecFromStr(strconv.FormatFloat(rawPrice, 'f', -1, 64))
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to parse Huobi price (%f) for %s", rawPrice, ticker)
	}

	// We have to fetch the volume separately since the last trade response does
	// not include the last 24h volume.
	volume, err := p.getTickerVolume(ticker)
	if err != nil {
		return TickerPrice{}, err
	}

	return TickerPrice{Price: price, Volume: volume}, nil
}

func (p HuobiProvider) getTickerVolume(ticker string) (sdk.Dec, error) {
	path := fmt.Sprintf("%s%s?symbol=%s", p.baseURL, huobiMarketEndpoint, strings.ToLower(ticker))

	resp, err := p.client.Get(path)
	if err != nil {
		return sdk.ZeroDec(), fmt.Errorf("failed to make Huobi request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return sdk.ZeroDec(), fmt.Errorf("failed to read Huobi response body: %w", err)
	}

	var marketResp HuobiMarketResponse
	if err := json.Unmarshal(bz, &marketResp); err != nil {
		return sdk.ZeroDec(), fmt.Errorf("failed to unmarshal Huobi response body: %w", err)
	}

	if !strings.EqualFold(marketResp.Status, "ok") {
		return sdk.ZeroDec(), fmt.Errorf(
			"received unexpected error from Huobi response: %s",
			marketResp.ErrMsg,
		)
	}

	rawVolume := marketResp.Tick.Vol
	volume, err := sdk.NewDecFromStr(strconv.FormatFloat(rawVolume, 'f', -1, 64))
	if err != nil {
		return sdk.ZeroDec(), fmt.Errorf("failed to parse Huobi volume (%f) for %s", rawVolume, ticker)
	}

	return volume, nil
}
