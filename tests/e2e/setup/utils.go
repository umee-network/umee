package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/ory/dockertest/v3/docker"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/unknownproto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	oracletypes "github.com/umee-network/umee/v5/x/oracle/types"
	"github.com/umee-network/umee/v5/x/uibc"
)

func (s *E2ETestSuite) UmeeREST() string {
	return fmt.Sprintf("http://%s", s.ValResources[0].GetHostPort("1317/tcp"))
}

func (s *E2ETestSuite) GaiaREST() string {
	return fmt.Sprintf("http://%s", s.GaiaResource.GetHostPort("1317/tcp"))
}

func (s *E2ETestSuite) SendIBC(srcChainID, dstChainID, recipient string, token sdk.Coin) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("sending %s from %s to %s (%s)", token, srcChainID, dstChainID, recipient)
	cmd := []string{
		"hermes",
		"tx",
		"raw",
		"ft-transfer",
		dstChainID,
		srcChainID,
		"transfer",  // source chain port ID
		"channel-0", // since only one connection/channel exists, assume 0
		token.Amount.String(),
		fmt.Sprintf("--denom=%s", token.Denom),
		"--timeout-height-offset=1000",
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

	s.Require().NoErrorf(
		err,
		"failed to send IBC tokens; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)
	s.T().Log("successfully sent IBC tokens")
	s.T().Log("Waiting for 12 seconds to make sure trasaction is processed or include in the block")
	time.Sleep(time.Second * 12)
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

func decodeTx(cdc codec.Codec, txBytes []byte) (*sdktx.Tx, error) {
	var raw sdktx.TxRaw

	// reject all unknown proto fields in the root TxRaw
	err := unknownproto.RejectUnknownFieldsStrict(txBytes, &raw, encodingConfig.InterfaceRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := cdc.Unmarshal(txBytes, &raw); err != nil {
		return nil, err
	}

	var body sdktx.TxBody
	if err := cdc.Unmarshal(raw.BodyBytes, &body); err != nil {
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	var authInfo sdktx.AuthInfo

	// reject all unknown proto fields in AuthInfo
	err = unknownproto.RejectUnknownFieldsStrict(raw.AuthInfoBytes, &authInfo, encodingConfig.InterfaceRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := cdc.Unmarshal(raw.AuthInfoBytes, &authInfo); err != nil {
		return nil, fmt.Errorf("failed to decode auth info: %w", err)
	}

	return &sdktx.Tx{
		Body:       &body,
		AuthInfo:   &authInfo,
		Signatures: raw.Signatures,
	}, nil
}
