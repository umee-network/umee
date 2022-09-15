package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/umee-network/umee/v3/x/leverage/types"
)

// ParseUpdateRegistryProposal attempts to parse a UpdateRegistryProposal from
// a JSON file.
func ParseUpdateRegistryProposal(cdc codec.JSONCodec, proposalFile string) (types.UpdateRegistryProposal, error) {
	content := types.UpdateRegistryProposal{}

	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return content, err
	}

	if err = cdc.UnmarshalJSON(contents, &content); err != nil {
		return content, err
	}

	return content, nil
}
