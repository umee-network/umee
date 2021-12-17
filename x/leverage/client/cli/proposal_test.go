package cli_test

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	umeeappbeta "github.com/umee-network/umee/app/beta"
	"github.com/umee-network/umee/x/leverage/client/cli"
)

func TestParseUpdateRegistryProposal(t *testing.T) {
	encCfg := umeeappbeta.MakeEncodingConfig()
	tmpDir := t.TempDir()

	// create a bogus proposal file and ensure parsing fails
	filePath := path.Join(tmpDir, "bad_proposal.json")
	bz := []byte(`
		foo
	`)
	os.WriteFile(filePath, bz, 0644)

	_, err := cli.ParseUpdateRegistryProposal(encCfg.Marshaler, filePath)
	require.Error(t, err)

	// create a good proposal file and ensure parsing does not fail
	filePath = path.Join(tmpDir, "good_proposal.json")
	bz = []byte(`{
	"title": "Update the Leverage Token Registry",
	"description": "Replace the supported tokens in the leverage registry.",
	"registry": [
		{
			"base_denom": "uumee",
			"reserve_factor": "0.1",
			"collateral_weight": "0.05",
			"base_borrow_rate": "0.02",
			"kink_borrow_rate": "0.2",
			"max_borrow_rate": "1.5",
			"kink_utilization_rate": "0.2",
			"liquidation_incentive": "0.1"
		}
	]
}`)
	os.WriteFile(filePath, bz, 0644)

	_, err = cli.ParseUpdateRegistryProposal(encCfg.Marshaler, filePath)
	require.NoError(t, err)
}
