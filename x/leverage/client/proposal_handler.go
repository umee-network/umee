package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/umee-network/umee/v2/x/leverage/client/cli"
)

// TODO: need to be updated to new gov message hanlders
//    https://github.com/umee-network/umee/issues/1001

// ProposalHandler defines an x/gov proposal handler for the CLI only.
var ProposalHandler = govclient.NewProposalHandler(
	cli.NewCmdSubmitUpdateRegistryProposal)
