package e2e

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

	gravitytypes "github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ory/dockertest/v3/docker"
)

func (s *IntegrationTestSuite) deployERC20Token(baseDenom string) string {
	s.T().Logf("deploying ERC20 token contract: %s", baseDenom)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.orchResources[0].Container.ID,
		User:         "root",
		Env:          []string{"PEGGO_ETH_PK=" + ethMinerPK},
		Cmd: []string{
			"peggo",
			"bridge",
			"deploy-erc20",
			s.gravityContractAddr,
			baseDenom,
			"--eth-rpc",
			fmt.Sprintf("http://%s:8545", s.ethResource.Container.Name[1:]),
			"--cosmos-chain-id",
			s.chain.id,
			"--cosmos-grpc",
			fmt.Sprintf("tcp://%s:9090", s.valResources[0].Container.Name[1:]),
			"--tendermint-rpc",
			fmt.Sprintf("http://%s:26657", s.valResources[0].Container.Name[1:]),
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(
		err,
		"failed to get ERC20 deployment logs; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	re := regexp.MustCompile(`Transaction: (0x.+)`)
	tokens := re.FindStringSubmatch(errBuf.String())
	s.Require().Lenf(tokens, 2, "stderr: %s", errBuf.String())

	txHash := tokens[1]
	s.Require().NotEmpty(txHash)

	s.Require().Eventually(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := queryEthTx(ctx, s.ethClient, txHash); err != nil {
				return false
			}

			return true
		},
		6*time.Minute,
		time.Second,
		"failed to confirm ERC20 deployment transaction",
	)

	umeeAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[0].GetHostPort("1317/tcp"))

	var erc20Addr string
	s.Require().Eventually(
		func() bool {
			addr, cosmosNative, err := queryDenomToERC20(umeeAPIEndpoint, baseDenom)
			if err != nil {
				return false
			}

			if cosmosNative && len(addr) > 0 {
				erc20Addr = addr
				return true
			}

			return false
		},
		2*time.Minute,
		time.Second,
		"failed to query ERC20 contract address",
	)

	s.T().Logf("deployed %s contract: %s", baseDenom, erc20Addr)

	return erc20Addr
}

func (s *IntegrationTestSuite) sendFromUmeeToEth(valIdx int, ethDest, amount, umeeFee, gravityFee string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	valAddr, err := s.chain.validators[valIdx].keyInfo.GetAddress()
	s.Require().NoError(err)

	s.T().Logf(
		"sending tokens from Umee to Ethereum; from: %s, to: %s, amount: %s, umeeFee: %s, gravityFee: %s",
		valAddr, ethDest, amount, umeeFee, gravityFee,
	)

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.valResources[valIdx].Container.ID,
		User:         "root",
		Cmd: []string{
			"umeed",
			"tx",
			"gravity",
			"send-to-eth",
			ethDest,
			amount,
			gravityFee,
			fmt.Sprintf("--%s=%s", flags.FlagFrom, s.chain.validators[valIdx].keyInfo.Name),
			fmt.Sprintf("--%s=%s", flags.FlagChainID, s.chain.id),
			fmt.Sprintf("--%s=%s", flags.FlagFees, umeeFee),
			"--keyring-backend=test",
			"--broadcast-mode=sync",
			"--output=json",
			"-y",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(err, "stdout: %s, stderr: %s", outBuf.String(), errBuf.String())

	var broadcastResp map[string]interface{}
	s.Require().NoError(json.Unmarshal(outBuf.Bytes(), &broadcastResp), outBuf.String())

	endpoint := fmt.Sprintf("http://%s", s.valResources[valIdx].GetHostPort("1317/tcp"))
	txHash := broadcastResp["txhash"].(string)

	s.Require().Eventuallyf(
		func() bool {
			return queryUmeeTx(endpoint, txHash) == nil
		},
		2*time.Minute,
		5*time.Second,
		"stdout: %s, stderr: %s",
		outBuf.String(), errBuf.String(),
	)
}

func (s *IntegrationTestSuite) sendFromUmeeToEthCheck(
	umeeValIdxSender,
	orchestratorIdxReceiver int,
	ethTokenAddr string,
	amount, umeeFee, gravityFee sdk.Coin,
) {
	if !strings.EqualFold(amount.Denom, gravityFee.Denom) {
		s.T().Error("Amount and gravityFee should be the same denom", amount, gravityFee)
	}

	// if all the coins are on the same denom
	allSameDenom := strings.EqualFold(amount.Denom, umeeFee.Denom) && strings.EqualFold(amount.Denom, gravityFee.Denom)
	var umeeFeeBalanceBeforeSend sdk.Coin
	if !allSameDenom {
		umeeFeeBalanceBeforeSend, _ = s.queryUmeeBalance(umeeValIdxSender, umeeFee.Denom)
	}

	umeeAmountBalanceBeforeSend, ethBalanceBeforeSend, _, ethAddr := s.queryUmeeEthBalance(umeeValIdxSender, orchestratorIdxReceiver, amount.Denom, ethTokenAddr) // 3300000000

	s.sendFromUmeeToEth(umeeValIdxSender, ethAddr, amount.String(), umeeFee.String(), gravityFee.String())
	umeeAmountBalanceAfterSend, ethBalanceAfterSend, _, _ := s.queryUmeeEthBalance(umeeValIdxSender, orchestratorIdxReceiver, amount.Denom, ethTokenAddr) // 3299999693

	if allSameDenom {
		s.Require().Equal(umeeAmountBalanceBeforeSend.Sub(amount).Sub(umeeFee).Sub(gravityFee).Amount.Int64(), umeeAmountBalanceAfterSend.Amount.Int64())
	} else { // the umeeFee and amount have different denom
		s.Require().Equal(umeeAmountBalanceBeforeSend.Sub(amount).Sub(gravityFee).Amount.Int64(), umeeAmountBalanceAfterSend.Amount.Int64())
		umeeFeeBalanceAfterSend, _ := s.queryUmeeBalance(umeeValIdxSender, umeeFee.Denom)
		s.Require().Equal(umeeFeeBalanceBeforeSend.Sub(umeeFee).Amount.Int64(), umeeFeeBalanceAfterSend.Amount.Int64())
	}

	// require the Ethereum recipient balance increased
	// peggo needs time to read the event and cross the tx
	ethLatestBalance := ethBalanceAfterSend
	expectedAmount := (ethBalanceBeforeSend + int64(amount.Amount.Int64()))
	s.Require().Eventuallyf(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			b, err := queryEthTokenBalance(ctx, s.ethClient, ethTokenAddr, ethAddr)
			if err != nil {
				return false
			}

			ethLatestBalance = b

			// The balance could differ if the receiving address was the orchestrator
			// that sent the batch tx and got the gravity fee.
			return b >= expectedAmount && b <= expectedAmount+gravityFee.Amount.Int64()
		},
		2*time.Minute,
		5*time.Second,
		"unexpected balance: %d", ethLatestBalance,
	)
}

func (s *IntegrationTestSuite) sendFromEthToUmeeCheck(
	orchestratorIdxSender,
	umeeValIdxReceiver int,
	ethTokenAddr,
	umeeTokenDenom string,
	amount uint64,
) {
	umeeBalanceBeforeSend, ethBalanceBeforeSend, umeeAddr, _ := s.queryUmeeEthBalance(umeeValIdxReceiver, orchestratorIdxSender, umeeTokenDenom, ethTokenAddr)
	s.sendFromEthToUmee(orchestratorIdxSender, ethTokenAddr, umeeAddr, fmt.Sprintf("%d", amount))
	umeeBalanceAfterSend, ethBalanceAfterSend, _, _ := s.queryUmeeEthBalance(umeeValIdxReceiver, orchestratorIdxSender, umeeTokenDenom, ethTokenAddr)

	s.Require().Equal(ethBalanceBeforeSend-int64(amount), ethBalanceAfterSend)

	umeeEndpoint := fmt.Sprintf("http://%s", s.valResources[umeeValIdxReceiver].GetHostPort("1317/tcp"))
	// require the original sender's (validator) balance increased
	// peggo needs time to read the event and cross the tx
	umeeLatestBalance := umeeBalanceAfterSend.Amount
	s.Require().Eventuallyf(
		func() bool {
			b, err := queryUmeeDenomBalance(umeeEndpoint, umeeAddr, umeeTokenDenom)
			if err != nil {
				s.T().Logf("Error at sendFromEthToUmeeCheck.queryUmeeDenomBalance %+v", err)
				return false
			}

			umeeLatestBalance = b.Amount

			return umeeBalanceBeforeSend.Amount.AddRaw(int64(amount)).Equal(umeeLatestBalance)
		},
		2*time.Minute,
		5*time.Second,
		"unexpected balance: %d", umeeLatestBalance.Int64(),
	)
}

func (s *IntegrationTestSuite) sendFromEthToUmee(valIdx int, tokenAddr, toUmeeAddr, amount string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf(
		"sending tokens from Ethereum to Umee; from: %s, to: %s, amount: %s, contract: %s",
		s.chain.orchestrators[valIdx].ethereumKey.address, toUmeeAddr, amount, tokenAddr,
	)

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.orchResources[valIdx].Container.ID,
		User:         "root",
		Env:          []string{"PEGGO_ETH_PK=" + s.chain.orchestrators[valIdx].ethereumKey.privateKey},
		Cmd: []string{
			"peggo",
			"bridge",
			"send-to-cosmos",
			s.gravityContractAddr,
			tokenAddr,
			toUmeeAddr,
			amount,
			"--eth-rpc",
			fmt.Sprintf("http://%s:8545", s.ethResource.Container.Name[1:]),
			"--cosmos-chain-id",
			s.chain.id,
			"--cosmos-grpc",
			fmt.Sprintf("tcp://%s:9090", s.valResources[valIdx].Container.Name[1:]),
			"--tendermint-rpc",
			fmt.Sprintf("http://%s:26657", s.valResources[valIdx].Container.Name[1:]),
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(err, "stdout: %s, stderr: %s", outBuf.String(), errBuf.String())

	re := regexp.MustCompile(`Transaction: (0x.+)`)
	tokens := re.FindStringSubmatch(errBuf.String())
	s.Require().Len(tokens, 2)

	txHash := tokens[1]
	s.Require().NotEmpty(txHash)

	s.Require().Eventuallyf(
		func() bool {
			return queryEthTx(ctx, s.ethClient, txHash) == nil
		},
		5*time.Minute,
		5*time.Second,
		"stdout: %s, stderr: %s",
		outBuf.String(), errBuf.String(),
	)
}

func (s *IntegrationTestSuite) connectIBCChains() {
	s.T().Logf("connecting %s and %s chains via IBC", s.chain.id, gaiaChainID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.hermesResource.Container.ID,
		User:         "root",
		Cmd: []string{
			"hermes",
			"create",
			"channel",
			s.chain.id,
			gaiaChainID,
			"--port-a=transfer",
			"--port-b=transfer",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(
		err,
		"failed connect chains; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.Require().Containsf(
		errBuf.String(),
		"successfully opened init channel",
		"failed to connect chains via IBC: %s", errBuf.String(),
	)

	s.T().Logf("connected %s and %s chains via IBC", s.chain.id, gaiaChainID)
}

func (s *IntegrationTestSuite) sendIBC(srcChainID, dstChainID, recipient string, token sdk.Coin) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("sending %s from %s to %s (%s)", token, srcChainID, dstChainID, recipient)

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.hermesResource.Container.ID,
		User:         "root",
		Cmd: []string{
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
			fmt.Sprintf("--receiver=%s", recipient),
			"--timeout-height-offset=1000",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
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
}

func queryUmeeTx(endpoint, txHash string) error {
	resp, err := http.Get(fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", endpoint, txHash))
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("tx query returned non-200 status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	txResp := result["tx_response"].(map[string]interface{})
	if v := txResp["code"]; v.(float64) != 0 {
		return fmt.Errorf("tx %s failed with status code %v", txHash, v)
	}

	return nil
}

func queryUmeeAllBalances(endpoint, addr string) (sdk.Coins, error) {
	resp, err := http.Get(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", endpoint, addr))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var balancesResp banktypes.QueryAllBalancesResponse
	if err := cdc.UnmarshalJSON(bz, &balancesResp); err != nil {
		return nil, err
	}

	return balancesResp.Balances, nil
}

func queryUmeeDenomBalance(endpoint, addr, denom string) (sdk.Coin, error) {
	var zeroCoin sdk.Coin

	path := fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s/by_denom?denom=%s",
		endpoint, addr, denom,
	)
	resp, err := http.Get(path)
	if err != nil {
		return zeroCoin, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return zeroCoin, err
	}

	var balanceResp banktypes.QueryBalanceResponse
	if err := cdc.UnmarshalJSON(bz, &balanceResp); err != nil {
		return zeroCoin, err
	}

	return *balanceResp.Balance, nil
}

func queryDenomToERC20(endpoint, denom string) (string, bool, error) {
	resp, err := http.Get(fmt.Sprintf("%s/gravity/v1beta/cosmos_originated/denom_to_erc20?denom=%s", endpoint, denom))
	if err != nil {
		return "", false, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, err
	}

	var denomToERC20Resp gravitytypes.QueryDenomToERC20Response
	if err := cdc.UnmarshalJSON(bz, &denomToERC20Resp); err != nil {
		return "", false, err
	}

	return denomToERC20Resp.Erc20, denomToERC20Resp.CosmosOriginated, nil
}

func queryEthTx(ctx context.Context, c *ethclient.Client, txHash string) error {
	_, pending, err := c.TransactionByHash(ctx, ethcmn.HexToHash(txHash))
	if err != nil {
		return err
	}

	if pending {
		return fmt.Errorf("ethereum tx %s is still pending", txHash)
	}

	return nil
}

func queryEthTokenBalance(ctx context.Context, c *ethclient.Client, contractAddr, recipientAddr string) (int64, error) {
	data, err := ethABI.Pack(abiMethodNameBalanceOf, ethcmn.HexToAddress(recipientAddr))
	if err != nil {
		return 0, fmt.Errorf("failed to pack ABI method call: %w", err)
	}

	token := ethcmn.HexToAddress(contractAddr)
	callMsg := ethereum.CallMsg{
		To:   &token,
		Data: data,
	}

	bz, err := c.CallContract(ctx, callMsg, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call Ethereum contract: %w", err)
	}

	balance, err := strconv.ParseInt(ethcmn.Bytes2Hex(bz), 16, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse balance: %w", err)
	}

	return balance, nil
}

func (s *IntegrationTestSuite) queryUmeeBalance(
	umeeValIdx int,
	umeeTokenDenom string,
) (umeeBalance sdk.Coin, umeeAddr string) {
	umeeEndpoint := fmt.Sprintf("http://%s", s.valResources[umeeValIdx].GetHostPort("1317/tcp"))
	umeeAddress, err := s.chain.validators[umeeValIdx].keyInfo.GetAddress()
	s.Require().NoError(err)
	umeeAddr = umeeAddress.String()

	umeeBalance, err = queryUmeeDenomBalance(umeeEndpoint, umeeAddr, umeeTokenDenom)
	s.Require().NoError(err)
	s.T().Logf(
		"Umee Balance of tokens validator; index: %d, addr: %s, amount: %s, denom: %s",
		umeeValIdx, umeeAddr, umeeBalance.String(), umeeTokenDenom,
	)

	return umeeBalance, umeeAddr
}

func (s *IntegrationTestSuite) queryUmeeEthBalance(
	umeeValIdx,
	orchestratorIdx int,
	umeeTokenDenom,
	ethTokenAddr string,
) (umeeBalance sdk.Coin, ethBalance int64, umeeAddr, ethAddr string) {
	umeeBalance, umeeAddr = s.queryUmeeBalance(umeeValIdx, umeeTokenDenom)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	ethAddr = s.chain.orchestrators[orchestratorIdx].ethereumKey.address

	ethBalance, err := queryEthTokenBalance(ctx, s.ethClient, ethTokenAddr, ethAddr)
	s.Require().NoError(err)
	s.T().Logf(
		"ETh Balance of tokens; index: %d, addr: %s, amount: %d, denom: %s, erc20Addr: %s",
		orchestratorIdx, ethAddr, ethBalance, photonDenom, ethTokenAddr,
	)

	return umeeBalance, ethBalance, umeeAddr, ethAddr
}
