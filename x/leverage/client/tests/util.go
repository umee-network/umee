package tests

import (
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/umee-network/umee/x/leverage/client/cli"
	"github.com/umee-network/umee/x/leverage/types"
)

// UpdateRegistry submits an UpdateRegistryProposal governance proposal with a
// deposit and automatically votes yes on it. It will wait until the proposal
// passes prior to returning. Note, the extraArgs are passed into the proposal
// creation command along with the vote command.
func (s *IntegrationTestSuite) UpdateRegistry(
	clientCtx client.Context,
	content *types.UpdateRegistryProposal,
	deposit sdk.Coins,
	extraArgs ...string,
) {
	// create proposal file
	dir := s.T().TempDir()
	path := path.Join(dir, "proposal.json")

	bz, err := clientCtx.Codec.MarshalJSON(content)
	s.Require().NoError(err)
	s.Require().NoError(ioutil.WriteFile(path, bz, 0600))

	cmd := cli.NewCmdSubmitUpdateRegistryProposal()
	flags.AddTxFlagsToCmd(cmd) // add flags manually since the gov workflow adds them automatically

	// submit proposal
	_, err = clitestutil.ExecTestCLICmd(
		clientCtx,
		cmd,
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
				clientCtx,
				govcli.GetCmdQueryProposals(),
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
				var c govtypes.Content
				if err := clientCtx.Codec.UnpackAny(p.Content, &c); err != nil {
					return false
				}

				if c.GetTitle() == content.Title {
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
	s.Require().Eventuallyf(
		func() bool {
			out, err := clitestutil.ExecTestCLICmd(
				clientCtx,
				govcli.GetCmdQueryProposal(),
				[]string{
					proposalIDStr,
					fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				},
			)
			if err != nil {
				return false
			}

			var prop govtypes.Proposal
			if err := clientCtx.Codec.UnmarshalJSON(out.Bytes(), &prop); err != nil {
				return false
			}

			return prop.Status == govtypes.StatusPassed
		},
		2*time.Minute,
		time.Second,
		"proposal %d (%s) failed to pass", proposalID, content.Title,
	)
}
