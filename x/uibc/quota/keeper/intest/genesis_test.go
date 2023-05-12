package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/v4/x/uibc"
)

func TestGenesis(t *testing.T) {
	require := require.New(t)
	s := initTestSuite(t)
	kb := s.app.UIbcQuotaKeeperB

	genesis := uibc.DefaultGenesisState()
	genesis.Outflows = append(genesis.Outflows, sampleOutflow)
	require.Equal(uibc.IBCTransferQuotaStatus_IBC_TRANSFER_QUOTA_STATUS_ENABLED, genesis.Params.QuotaStatus)
	kb.InitGenesis(s.ctx, *genesis)

	// verify params
	k := kb.Keeper(&s.ctx)
	defaultParams := uibc.DefaultParams()
	params := k.GetParams()
	require.Equal(genesis.Params, params)
	require.Equal(defaultParams, params)

	// verify export
	exported := kb.ExportGenesis(s.ctx)
	require.Equal(exported, genesis)

	// update params in genesis state
	genesis.Params.QuotaStatus = uibc.IBCTransferQuotaStatus_IBC_TRANSFER_QUOTA_STATUS_DISABLED
	kb.InitGenesis(s.ctx, *genesis)
	exported = kb.ExportGenesis(s.ctx)
	require.Equal(genesis.Params.QuotaStatus, exported.Params.QuotaStatus)
}
