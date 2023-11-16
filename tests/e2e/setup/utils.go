package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/unknownproto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/ory/dockertest/v3/docker"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/client"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v6/x/oracle/types"
	"github.com/umee-network/umee/v6/x/uibc"
)

func (s *E2ETestSuite) UmeeREST() string {
	return fmt.Sprintf("http://%s", s.ValResources[0].GetHostPort("1317/tcp"))
}

func (s *E2ETestSuite) GaiaREST() string {
	return fmt.Sprintf("http://%s", s.GaiaResource.GetHostPort("1317/tcp"))
}

// Delegates an amount of uumee from the test account at a given index to a specified validator.
func (s *E2ETestSuite) Delegate(testAccount, valIndex int, amount uint64) error {
	addr := s.AccountAddr(testAccount)

	if len(s.Chain.Validators) <= valIndex {
		return fmt.Errorf("validator %d not found", valIndex)
	}
	valAddr, err := s.Chain.Validators[valIndex].KeyInfo.GetAddress()
	if err != nil {
		return err
	}
	valOperAddr := sdk.ValAddress(valAddr)

	asset := sdk.NewCoin(appparams.BaseDenom, sdk.NewIntFromUint64(amount))
	msg := stakingtypes.NewMsgDelegate(addr, valOperAddr, asset)
	return s.BroadcastTxWithRetry(msg, s.AccountClient(testAccount))
}

func (s *E2ETestSuite) SendIBC(srcChainID, dstChainID, recipient string, token sdk.Coin, failDueToQuota bool, desc string) {
	s.T().Logf("sending %s from %s to %s (exceed quota: %t) %s", token, srcChainID, dstChainID, failDueToQuota, desc)
	// ibctransfertypes.NewMsgTransfer()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// retry up to 5 times
	for i := 0; i < 5; i++ {
		if i > 0 {
			s.T().Logf("...try %d", i+1)
		}

		cmd := []string{
			"hermes",
			"tx",
			"ft-transfer",
			"--dst-chain",
			dstChainID,
			"--src-chain",
			srcChainID,
			"--src-port",
			"transfer", // source chain port ID
			"--src-channel",
			"channel-0", // since only one connection/channel exists, assume 0
			"--amount",
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
			time.Sleep(1 * time.Second)
			continue
		}

		s.Require().NoErrorf(
			err,
			"failed to send IBC tokens; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
		)

		// Note: we are cchecking only one side of ibc , we don't know whethever ibc transfer is succeed on one side
		// some times relayer can't send the packets to another chain

		// // don't check for the tx hash if we expect this to fail due to quota
		if strings.Contains(errBuf.String(), "quota transfer exceeded") {
			s.Require().True(failDueToQuota)
			return
		}

		// retry if we didn't succeed
		if !strings.Contains(outBuf.String(), "SUCCESS") {
			if i < 4 {
				continue
			}
			s.Require().Failf("failed to find transaction hash in output outBuf: %s  errBuf: %s", outBuf.String(), errBuf.String())
		}

		// s.Require().NotEmptyf(txHash, "failed to find transaction hash in output outBuf: %s  errBuf: %s", outBuf.String(), errBuf.String())
		// endpoint := s.UmeeREST()
		// if strings.Contains(srcChainID, "gaia") {
		// 	endpoint = s.GaiaREST()
		// }

		// s.Require().Eventually(func() bool {
		// 	err := s.QueryUmeeTx(endpoint, txHash)
		// 	if err != nil {
		// 		s.T().Log("Tx Query Error", err)
		// 	}
		// 	return err == nil
		// }, 5*time.Second, 200*time.Millisecond, "require tx to be included in block")
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

	return decodeRespBody(s.cdc, endpoint, resp.Body, valPtr)
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

func (s *E2ETestSuite) QueryRegisteredTokens(endpoint string) ([]leveragetypes.Token, error) {
	endpoint = fmt.Sprintf("%s/umee/leverage/v1/registered_tokens", endpoint)
	var resp leveragetypes.QueryRegisteredTokensResponse
	if err := s.QueryREST(endpoint, &resp); err != nil {
		return nil, err
	}

	return resp.Registry, nil
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

func (s *E2ETestSuite) QueryIBCChannels(endpoint string) (bool, error) {
	ibcChannelsEndPoint := fmt.Sprintf("%s/ibc/core/channel/v1/channels", endpoint)
	var resp channeltypes.QueryChannelsResponse
	if err := s.QueryREST(ibcChannelsEndPoint, &resp); err != nil {
		return false, err
	}
	if len(resp.Channels) > 0 {
		s.T().Log("✅ Channels state is  :", resp.Channels[0].State)
		if resp.Channels[0].State == channeltypes.OPEN {
			s.T().Log("✅ Channels are created among the chains :", resp.Channels[0].ChannelId)
			return true, nil
		}
	}
	return false, nil
}

func (s *E2ETestSuite) BroadcastTxWithRetry(msg sdk.Msg, cli client.Client) error {
	var err error
	// TODO: decrease it when possible
	for retry := 0; retry < 10; retry++ {
		// retry if txs fails, because sometimes account sequence mismatch occurs due to txs pending
		_, err = cli.Tx.BroadcastTx(0, msg)
		if err == nil {
			return nil
		}

		if err != nil && !strings.Contains(err.Error(), "incorrect account sequence") {
			return err
		}

		// if we were told an expected account sequence, we should use it next time
		re := regexp.MustCompile(`expected [\d]+`)
		n, err := strconv.Atoi(strings.TrimPrefix(re.FindString(err.Error()), "expected "))
		if err != nil {
			return err
		}
		cli.WithAccSeq(uint64(n))

		time.Sleep(time.Millisecond * 300)
	}

	return err
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

func decodeRespBody(cdc codec.Codec, endpoint string, body io.ReadCloser, valPtr interface{}) error {
	if valProto, ok := valPtr.(proto.Message); ok {
		bz, err := io.ReadAll(body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w, endpoint: %s", err, endpoint)
		}
		if err = cdc.UnmarshalJSON(bz, valProto); err != nil {
			return fmt.Errorf("failed to protoJSON.decode response body: %w, endpoint: %s", err, endpoint)
		}
	} else {
		if err := json.NewDecoder(body).Decode(valPtr); err != nil {
			return fmt.Errorf("failed to json.decode response body: %w, endpoint: %s", err, endpoint)
		}
	}

	return nil
}
