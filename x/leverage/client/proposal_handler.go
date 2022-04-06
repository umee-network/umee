package client

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"

	"github.com/umee-network/umee/v2/x/leverage/client/cli"
)

// ProposalHandler defines an x/gov proposal handler for the CLI only.
var ProposalHandler = govclient.NewProposalHandler(
	cli.NewCmdSubmitUpdateRegistryProposal,
	updateRegistryProposalNoOpHandler,
)

func updateRegistryProposalNoOpHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "update_registry",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			rest.WriteErrorResponse(w, http.StatusNotFound, "unsupported route")
		},
	}
}
