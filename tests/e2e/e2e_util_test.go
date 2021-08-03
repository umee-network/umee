package e2e

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ory/dockertest/v3/docker"
)

func (s *IntegrationTestSuite) deployERC20Token(baseDenom, name, symbol string, decimals int) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

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
			fmt.Sprintf("--erc20-decimals=%d", decimals),
			fmt.Sprintf("--erc20-symbol=%s", symbol),
			fmt.Sprintf("--erc20-name=%s", name),
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

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(err, "stdout: %s, stderr: %s", outBuf, errBuf)

	re := regexp.MustCompile(`We have deployed ERC20 contract (0[xX][0-9a-fA-F]+)`)
	matches := re.FindStringSubmatch(outBuf.String())
	s.Require().GreaterOrEqualf(len(matches), 2, "stdout: %s, stderr: %s", outBuf, errBuf)

	erc20Addr := matches[1]
	_, err = hexutil.Decode(erc20Addr)
	s.Require().NoError(err)

	return erc20Addr
}
