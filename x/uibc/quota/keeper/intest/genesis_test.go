package intest

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/v5/x/uibc"
)

func TestGenesis(t *testing.T) {
	require := require.New(t)
	s := initTestSuite(t)
	kb := s.app.UIbcQuotaKeeperB

	genesis := uibc.DefaultGenesisState()
	genesis.Outflows = append(genesis.Outflows, sampleOutflow)
	require.Equal(uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_ENABLED, genesis.Params.IbcStatus)
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
	genesis.Params.IbcStatus = uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_DISABLED
	kb.InitGenesis(s.ctx, *genesis)
	exported = kb.ExportGenesis(s.ctx)
	require.Equal(genesis.Params.IbcStatus, exported.Params.IbcStatus)
}
