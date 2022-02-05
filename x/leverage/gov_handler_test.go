package leverage_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	umeeappbeta "github.com/umee-network/umee/app/beta"
	"github.com/umee-network/umee/x/leverage"
	"github.com/umee-network/umee/x/leverage/types"
)

func TestUpdateRegistryProposalHandler(t *testing.T) {
	app := umeeappbeta.Setup(t, false, 1)
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
		k.SetRegisteredToken(ctx, types.Token{BaseDenom: "uatom", BaseBorrowRate: sdk.MustNewDecFromStr("0.05")})

		p := &types.UpdateRegistryProposal{
			Title:       "test",
			Description: "test",
			Registry: []types.Token{
				{BaseDenom: "uumee"},
				{BaseDenom: "uatom", BaseBorrowRate: sdk.MustNewDecFromStr("0.02")},
			},
		}
		require.NoError(t, h(ctx, p))

		tokens := k.GetAllRegisteredTokens(ctx)
		require.Len(t, tokens, 2)

		_, err := k.GetRegisteredToken(ctx, "uosmo")
		require.Error(t, err)

		_, err = k.GetRegisteredToken(ctx, "uumee")
		require.NoError(t, err)

		token, err := k.GetRegisteredToken(ctx, "uatom")
		require.NoError(t, err)
		require.Equal(t, "0.020000000000000000", token.BaseBorrowRate.String())
	})
}
