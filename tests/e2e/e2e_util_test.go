package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ory/dockertest/v3/docker"
)

func (s *IntegrationTestSuite) deployERC20Token(baseDenom string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("deploying ERC20 token contract: %s", baseDenom)

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.orchResources[0].Container.ID,
		User:         "root",
		Cmd: []string{
			"gravity-client",
			"deploy-erc20-representation",
			fmt.Sprintf("--ethereum-rpc=http://%s:8545", s.ethResource.Container.Name[1:]),
			fmt.Sprintf("--cosmos-grpc=http://%s:9090", s.valResources[0].Container.Name[1:]),
			fmt.Sprintf("--cosmos-denom=%s", baseDenom),
			fmt.Sprintf("--contract-address=%s", s.gravityContractAddr),
			fmt.Sprintf("--ethereum-key=%s", s.chain.validators[0].ethereumKey.privateKey),
			"--cosmos-prefix=umee",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	// TODO: This sometimes fails with "replacement transaction underpriced". We
	// should:
	//
	// 1. Consider instead sending the raw Ethereum transaction ourselves instead
	// of via the 'deploy-erc20-representation' command so we can control the
	// nonce ourselves if this error happens.
	//
	// 2. Or, wrap this call in an eventually/retry block.
	//
	//
	// Ref: https://github.com/umee-network/umee/issues/12
	// Ref: https://ethereum.stackexchange.com/questions/27256/error-replacement-transaction-underpriced
	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(err, "stdout: %s, stderr: %s", outBuf.String(), errBuf.String())

	re := regexp.MustCompile(`has accepted new ERC20 representation (0[xX][0-9a-fA-F]+)`)
	matches := re.FindStringSubmatch(outBuf.String())
	s.Require().GreaterOrEqualf(len(matches), 2, "stdout: %s, stderr: %s", outBuf.String(), errBuf.String())

	erc20Addr := matches[1]
	_, err = hexutil.Decode(erc20Addr)
	s.Require().NoError(err)

	s.T().Logf("deployed %s contract: %s", baseDenom, erc20Addr)

	return erc20Addr
}

func (s *IntegrationTestSuite) sendFromEthToUmee(valIdx int, tokenAddr, toUmeeAddr, amount string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf(
		"sending tokens from Ethereum to Umee; from: %s, to: %s, amount: %s, contract: %s",
		s.chain.validators[valIdx].ethereumKey.address, toUmeeAddr, amount, tokenAddr,
	)

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.valResources[valIdx].Container.ID,
		User:         "root",
		Cmd: []string{
			"gravity-client",
			"eth-to-cosmos",
			fmt.Sprintf("--ethereum-rpc=http://%s:8545", s.ethResource.Container.Name[1:]),
			fmt.Sprintf("--amount=%s", amount),
			fmt.Sprintf("--ethereum-key=%s", s.chain.validators[valIdx].ethereumKey.privateKey),
			fmt.Sprintf("--contract-address=%s", s.gravityContractAddr),
			fmt.Sprintf("--erc20-address=%s", tokenAddr),
			fmt.Sprintf("--cosmos-destination=%s", toUmeeAddr),
			"--cosmos-prefix=umee",
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

	re := regexp.MustCompile(`Send to Cosmos txid: (0[xX][0-9a-fA-F]+)`)
	matches := re.FindStringSubmatch(outBuf.String())
	s.Require().GreaterOrEqualf(len(matches), 2, "stdout: %s, stderr: %s", outBuf.String(), errBuf.String())

	txHash := matches[1]
	_, err = hexutil.Decode(txHash)
	s.Require().NoError(err)

	s.Require().Eventuallyf(
		func() bool {
			return queryEthTx(ctx, s.ethClient, txHash) == nil
		},
		time.Minute,
		5*time.Second,
		"stdout: %s, stderr: %s",
		outBuf.String(), errBuf.String(),
	)
}

func (s *IntegrationTestSuite) sendFromUmeeToEth(valIdx int, toEthAddr, amount, umeeFee, gravityFee string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf(
		"sending tokens from Umee to Ethereum; from: %s, to: %s, amount: %s, umeeFee: %s, gravityFee: %s",
		s.chain.validators[valIdx].keyInfo.GetAddress(), toEthAddr, amount, umeeFee, gravityFee,
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
			"send-to-etheruem",
			toEthAddr,
			amount,
			gravityFee,
			fmt.Sprintf("--%s=%s", flags.FlagFrom, s.chain.validators[valIdx].keyInfo.GetName()),
			fmt.Sprintf("--%s=%s", flags.FlagChainID, s.chain.id),
			fmt.Sprintf("--%s=%s", flags.FlagFees, umeeFee),
			"--keyring-backend=test",
			"--broadcast-mode=sync",
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
	s.Require().NoError(json.Unmarshal(outBuf.Bytes(), &broadcastResp))

	endpoint := fmt.Sprintf("http://%s", s.valResources[valIdx].GetHostPort("1317/tcp"))
	txHash := broadcastResp["txhash"].(string)

	s.Require().Eventuallyf(
		func() bool {
			return queryUmeeTx(endpoint, txHash) == nil
		},
		time.Minute,
		5*time.Second,
		"stdout: %s, stderr: %s",
		outBuf.String(), errBuf.String(),
	)
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

func queryUmeeDenomBalance(endpoint, addr, denom string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s/%s", endpoint, addr, denom))
	if err != nil {
		return 0, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	balance := result["balance"].(map[string]interface{})
	amount, err := strconv.Atoi(balance["amount"].(string))
	if err != nil {
		return 0, err
	}

	return amount, nil
}

func queryEthTx(ctx context.Context, c *ethclient.Client, txHash string) error {
	_, pending, err := c.TransactionByHash(ctx, common.HexToHash(txHash))
	if err != nil {
		return err
	}

	if pending {
		return fmt.Errorf("ethereum tx %s is still pending", txHash)
	}

	return nil
}

func queryEthTokenBalance(ctx context.Context, c *ethclient.Client, contractAddr, recipientAddr string) (int, error) {
	data, err := ethABI.Pack(abiMethodNameBalanceOf, common.HexToAddress(recipientAddr))
	if err != nil {
		return 0, fmt.Errorf("failed to pack ABI method call: %w", err)
	}

	token := common.HexToAddress(contractAddr)
	callMsg := ethereum.CallMsg{
		To:   &token,
		Data: data,
	}

	bz, err := c.CallContract(ctx, callMsg, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call Ethereum contract: %w", err)
	}

	balance, err := strconv.ParseInt(common.Bytes2Hex(bz), 16, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse balance: %w", err)
	}

	return int(balance), nil
}
