package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gogo/protobuf/proto"
	"github.com/ory/dockertest/v3/docker"

	oracletypes "github.com/umee-network/umee/v6/x/oracle/types"
	"github.com/umee-network/umee/v6/x/uibc"
)

func (s *E2ETestSuite) UmeeREST() string {
	return fmt.Sprintf("http://%s", s.ValResources[0].GetHostPort("1317/tcp"))
}

func (s *E2ETestSuite) GaiaREST() string {
	return fmt.Sprintf("http://%s", s.GaiaResource.GetHostPort("1317/tcp"))
}

func (s *E2ETestSuite) SendIBC(srcChainID, dstChainID, recipient string, token sdk.Coin, failDueToQuota bool) {
	// ibctransfertypes.NewMsgTransfer()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// retry up to 5 times
	for i := 0; i < 5; i++ {
		s.T().Logf("sending %s from %s to %s (%s). Try %d", token, srcChainID, dstChainID, recipient, i+1)
		cmd := []string{
			"hermes",
			"tx",
			"raw",
			"ft-transfer",
			dstChainID,
			srcChainID,
			"transfer",  // source Chain port ID
			"channel-0", // since only one connection/channel exists, assume 0
			token.Amount.String(),
			fmt.Sprintf("--denom=%s", token.Denom),
			"--timeout-height-offset=3000",
		}

		if len(recipient) != 0 {
			cmd = append(cmd, fmt.Sprintf("--receiver=%s", recipient))
		}

		exec, err := s.DkrPool.Client.CreateExec(docker.CreateExecOptions{
			Context:      ctx,
			AttachStdout: true,
			AttachStderr: true,
			Container:    s.HermesResource.Container.ID,
			Cmd:          cmd,
		})
		s.Require().NoError(err)

		var (
			outBuf bytes.Buffer
			errBuf bytes.Buffer
		)

		err = s.DkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
			Context:      ctx,
			Detach:       false,
			OutputStream: &outBuf,
			ErrorStream:  &errBuf,
		})

		// retry if we got an error
		if err != nil && i < 4 {
			continue
		}

		s.Require().NoErrorf(
			err,
			"failed to send IBC tokens; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
		)

		// don't check for the tx hash if we expect this to fail due to quota
		if strings.Contains(errBuf.String(), "quota transfer exceeded") {
			s.Require().True(failDueToQuota)
			return
		}

		re := regexp.MustCompile(`[0-9A-Fa-f]{64}`)
		txHash := re.FindString(errBuf.String() + outBuf.String())

		// retry if we didn't get a txHash
		if len(txHash) == 0 && i < 4 {
			continue
		}

		s.T().Log("successfully sent IBC tokens")
		s.Require().NotEmptyf(txHash, "failed to find transaction hash in output outBuf: %s  errBuf: %s", outBuf.String(), errBuf.String())
		s.T().Log("Waiting for Tx to be included in a block", txHash, srcChainID)
		endpoint := s.UmeeREST()
		if strings.Contains(srcChainID, "gaia") {
			endpoint = s.GaiaREST()
		}

		s.Require().Eventually(func() bool {
			err := s.QueryUmeeTx(endpoint, txHash)
			if err != nil {
				s.T().Log("Tx Query Error", err)
			}
			return err == nil
		}, 5*time.Second, 200*time.Millisecond)
		return
	}
}

// QueryREST make http query to grpc-web endpoint and tries to decode valPtr using proto-JSON
// decoder if valPtr implements proto.Message. Otherwise standard JSON decoder is used.
// valPtr must be a pointer.
func (s *E2ETestSuite) QueryREST(endpoint string, valPtr interface{}) error {
	resp, err := http.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("tx query returned non-200 status: %d (%s)", resp.StatusCode, endpoint)
	}

	if valProto, ok := valPtr.(proto.Message); ok {
		bz, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w, endpoint: %s", err, endpoint)
		}
		if err = s.cdc.UnmarshalJSON(bz, valProto); err != nil {
			return fmt.Errorf("failed to protoJSON.decode response body: %w, endpoint: %s", err, endpoint)
		}
	} else {
		if err := json.NewDecoder(resp.Body).Decode(valPtr); err != nil {
			return fmt.Errorf("failed to json.decode response body: %w, endpoint: %s", err, endpoint)
		}
	}

	return nil
}

func (s *E2ETestSuite) QueryUmeeTx(endpoint, txHash string) error {
	endpoint = fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", endpoint, txHash)
	var result map[string]interface{}
	if err := s.QueryREST(endpoint, &result); err != nil {
		return err
	}

	txResp := result["tx_response"].(map[string]interface{})
	if v := txResp["code"]; v.(float64) != 0 {
		return fmt.Errorf("tx %s failed with status code %v", txHash, v)
	}
	return nil
}

func (s *E2ETestSuite) QueryUmeeAllBalances(endpoint, addr string) (sdk.Coins, error) {
	endpoint = fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", endpoint, addr)
	var balancesResp banktypes.QueryAllBalancesResponse
	if err := s.QueryREST(endpoint, &balancesResp); err != nil {
		return nil, err
	}

	return balancesResp.Balances, nil
}

func (s *E2ETestSuite) QueryTotalSupply(endpoint string) (sdk.Coins, error) {
	endpoint = fmt.Sprintf("%s/cosmos/bank/v1beta1/supply", endpoint)
	var balancesResp banktypes.QueryTotalSupplyResponse
	if err := s.QueryREST(endpoint, &balancesResp); err != nil {
		return nil, err
	}

	return balancesResp.Supply, nil
}

func (s *E2ETestSuite) QueryExchangeRate(endpoint, denom string) (sdk.DecCoins, error) {
	endpoint = fmt.Sprintf("%s/umee/oracle/v1/denoms/exchange_rates/%s", endpoint, denom)
	var resp oracletypes.QueryExchangeRatesResponse
	if err := s.QueryREST(endpoint, &resp); err != nil {
		return nil, err
	}

	return resp.ExchangeRates, nil
}

func (s *E2ETestSuite) QueryHistAvgPrice(endpoint, denom string) (sdk.Dec, error) {
	endpoint = fmt.Sprintf("%s/umee/historacle/v1/avg_price/%s", endpoint, strings.ToUpper(denom))
	var resp oracletypes.QueryAvgPriceResponse
	if err := s.QueryREST(endpoint, &resp); err != nil {
		return sdk.Dec{}, err
	}

	return resp.Price, nil
}

func (s *E2ETestSuite) QueryOutflows(endpoint, denom string) (sdk.Dec, error) {
	endpoint = fmt.Sprintf("%s/umee/uibc/v1/outflows?denom=%s", endpoint, denom)
	var resp uibc.QueryOutflowsResponse
	if err := s.QueryREST(endpoint, &resp); err != nil {
		return sdk.Dec{}, err
	}

	return resp.Amount, nil
}

func (s *E2ETestSuite) QueryUmeeDenomBalance(endpoint, addr, denom string) (sdk.Coin, error) {
	endpoint = fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s/by_denom?denom=%s", endpoint, addr, denom)
	var resp banktypes.QueryBalanceResponse
	if err := s.QueryREST(endpoint, &resp); err != nil {
		return sdk.Coin{}, err
	}

	return *resp.Balance, nil
}

func (s *E2ETestSuite) QueryUmeeBalance(
	umeeValIdx int,
	umeeTokenDenom string,
) (umeeBalance sdk.Coin, umeeAddr string) {
	umeeEndpoint := fmt.Sprintf("http://%s", s.ValResources[umeeValIdx].GetHostPort("1317/tcp"))
	umeeAddress, err := s.Chain.Validators[umeeValIdx].KeyInfo.GetAddress()
	s.Require().NoError(err)
	umeeAddr = umeeAddress.String()

	umeeBalance, err = s.QueryUmeeDenomBalance(umeeEndpoint, umeeAddr, umeeTokenDenom)
	s.Require().NoError(err)
	s.T().Logf(
		"Umee Balance of tokens validator; index: %d, addr: %s, amount: %s, denom: %s",
		umeeValIdx, umeeAddr, umeeBalance.String(), umeeTokenDenom,
	)

	return umeeBalance, umeeAddr
}
