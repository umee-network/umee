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

	"github.com/umee-network/umee/app"
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
		10*time.Second,
		"proposal %d (%s) failed to pass", proposalID, content.Title,
	)
}

// updateCollateralWeight modifies the collateral weight of a registered token identified by baseDenom.
// Also returns its previous collateral weight, which is useful when undoing changes.
func updateCollateralWeight(s *IntegrationTestSuite, baseDenom string, collateralWeight sdk.Dec) sdk.Dec {
	val := s.network.Validators[0]
	clientCtx := s.network.Validators[0].ClientCtx

	queryFlags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	// Query all registered tokens
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryAllRegisteredTokens(), queryFlags)
	s.Require().NoError(err)

	resp := &types.QueryRegisteredTokensResponse{}
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp), out.String())

	// Replace the collateral weight of the selected token with the new value
	newTokens := resp.GetRegistry()
	oldCollateralWeight := sdk.ZeroDec()
	for _, token := range newTokens {
		if token.BaseDenom == baseDenom {
			oldCollateralWeight = token.CollateralWeight
			token.CollateralWeight = collateralWeight
		}
	}

	// Update token registry using the modified token registry - waits for proposal accepted
	s.UpdateRegistry(
		clientCtx,
		types.NewUpdateRegistryProposal(
			"test title",
			"test description",
			newTokens,
		),
		sdk.NewCoins(sdk.NewCoin(app.BondDenom, govtypes.DefaultMinDepositTokens)),
		[]string{
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(app.BondDenom, sdk.NewInt(10))).String()),
		}...,
	)

	// Query registered tokens again
	out, err = clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryAllRegisteredTokens(), queryFlags)
	s.Require().NoError(err)

	// Verify that the RegisteredTokens query is returning an updated collateral weight
	registry := &types.QueryRegisteredTokensResponse{}
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), registry), out.String())
	for _, token := range registry.GetRegistry() {
		if token.BaseDenom == baseDenom {
			s.Require().Equal(collateralWeight, token.CollateralWeight, "Collateral weight not updated")
		}
	}

	return oldCollateralWeight
}
