package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/umee-network/umee/x/leverage/client/cli"
)

// ProposalHandler defines an x/gov proposal handler for the CLI only.
var ProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitUpdateRegistryProposal, nil)
