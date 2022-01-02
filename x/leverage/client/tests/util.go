package tests

import (
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/umee-network/umee/x/leverage/client/cli"
	"github.com/umee-network/umee/x/leverage/types"
)

func (s *IntegrationTestSuite) UpdateRegistry(
	clientCtx client.Context,
	content types.UpdateRegistryProposal,
	deposit sdk.Coins,
	extraArgs ...string,
) {
	// create proposal file
	dir := s.T().TempDir()
	path := path.Join(dir, "proposal.json")

	bz, err := clientCtx.Codec.MarshalJSON(&content)
	s.Require().NoError(err)
	s.Require().NoError(ioutil.WriteFile(path, bz, 0644))

	// submit proposal
	_, err = clitestutil.ExecTestCLICmd(
		clientCtx,
		cli.NewCmdSubmitUpdateRegistryProposal(),
		append(
			[]string{
				path,
				deposit.String(),
			},
			extraArgs...,
		),
	)
	s.Require().NoError(err)

	// get proposal ID
	var proposalID uint64
	s.Require().Eventually(
		func() bool {
			out, err := clitestutil.ExecTestCLICmd(
				clientCtx, govcli.GetCmdQueryProposals(),
				[]string{
					fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				},
			)
			if err != nil {
				return false
			}

			var resp govtypes.QueryProposalsResponse
			if err := clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp); err != nil {
				return false
			}

			for _, p := range resp.Proposals {
				if p.GetTitle() == content.Title {
					proposalID = p.ProposalId
					return true
				}
			}

			return false
		},
		time.Minute,
		time.Second,
		"failed to find proposal",
	)

	proposalIDStr := strconv.Itoa(int(proposalID))

	// vote on proposal
	_, err = clitestutil.ExecTestCLICmd(
		clientCtx,
		govcli.NewCmdWeightedVote(),
		append(
			[]string{
				proposalIDStr,
				"yes",
			},
			extraArgs...,
		),
	)
	s.Require().NoError(err)

	// wait till proposal passes and is executed
	s.Require().Eventually(
		func() bool {
			out, err := clitestutil.ExecTestCLICmd(
				clientCtx, govcli.GetCmdQueryProposal(),
				[]string{
					proposalIDStr,
					fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				},
			)
			if err != nil {
				return false
			}

			var resp govtypes.QueryProposalResponse
			if err := clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp); err != nil {
				return false
			}

			return resp.Proposal.Status == govtypes.StatusPassed
		},
		time.Minute,
		time.Second,
		"proposal failed to pass",
	)
}
