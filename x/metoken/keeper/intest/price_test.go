package intest

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/x/metoken"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

func TestPrice_SetPricesToOracle(t *testing.T) {
	index := mocks.StableIndex(mocks.MeUSDDenom)

	s := initTestSuite(t, nil, nil)
	msgServer, ctx, app := s.msgServer, s.ctx, s.app

	_, err := msgServer.GovUpdateRegistry(
		ctx, &metoken.MsgGovUpdateRegistry{
			Authority:   checkers.GovModuleAddr,
			AddIndex:    []metoken.Index{index},
			UpdateIndex: nil,
		},
	)
	require := require.New(t)
	require.NoError(err)

	err = app.MetokenKeeperB.Keeper(&ctx).SetPricesToOracle()
	require.NoError(err)
}
