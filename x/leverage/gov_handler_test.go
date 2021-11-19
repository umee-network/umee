package leverage_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	umeeapp "github.com/umee-network/umee/app"
	"github.com/umee-network/umee/x/leverage"
	"github.com/umee-network/umee/x/leverage/types"
)

func TestUpdateRegistryProposalHandler(t *testing.T) {
	app := umeeapp.Setup(t, false, 1)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})

	k := app.LeverageKeeper
	h := leverage.NewUpdateRegistryProposalHandler(k)

	t.Run("invalid proposal type", func(t *testing.T) {
		require.Error(t, h(ctx, &paramsproposal.ParameterChangeProposal{}))
	})

	t.Run("valid proposal", func(t *testing.T) {
		k.SetRegisteredToken(ctx, types.Token{BaseDenom: "uosmo"})
		k.SetRegisteredToken(ctx, types.Token{BaseDenom: "uatom", BaseBorrowRate: sdk.MustNewDecFromStr("5.0")})

		p := &types.UpdateRegistryProposal{
			Title:       "test",
			Description: "test",
			Registry: []types.Token{
				{BaseDenom: "uumee"},
				{BaseDenom: "uatom", BaseBorrowRate: sdk.MustNewDecFromStr("2.0")},
			},
		}
		require.NoError(t, h(ctx, p))

		tokens, err := k.GetAllRegisteredTokens(ctx)
		require.NoError(t, err)
		require.Len(t, tokens, 2)

		_, err = k.GetRegisteredToken(ctx, "uosmo")
		require.Error(t, err)

		_, err = k.GetRegisteredToken(ctx, "uumee")
		require.NoError(t, err)

		token, err := k.GetRegisteredToken(ctx, "uatom")
		require.NoError(t, err)
		require.Equal(t, "2.000000000000000000", token.BaseBorrowRate.String())
	})
}
