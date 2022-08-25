package leverage_test

import (
	"fmt"
	"testing"

	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	umeeapp "github.com/umee-network/umee/v3/app"
	"github.com/umee-network/umee/v3/x/leverage"
	"github.com/umee-network/umee/v3/x/leverage/fixtures"
	"github.com/umee-network/umee/v3/x/leverage/types"
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

	t.Run("invalid token", func(t *testing.T) {
		p := &types.UpdateRegistryProposal{
			Title:       "test",
			Description: "test",
			Registry: []types.Token{
				fixtures.Token("uosmo", ""), // empty denom is invalid
			},
		}
		require.Error(t, h(ctx, p))
	})

	t.Run("valid proposal", func(t *testing.T) {
		require.NoError(t, k.SetTokenSettings(ctx,
			fixtures.Token("uosmo", "OSMO"),
		))
		require.NoError(t, k.SetTokenSettings(ctx,
			fixtures.Token("uatom", "ATOM"),
		))

		osmo := fixtures.Token("uosmo", "OSMO")
		osmo.ReserveFactor = sdk.MustNewDecFromStr("0.3")
		p := &types.UpdateRegistryProposal{
			Title:       "test",
			Description: "test",
			Registry: []types.Token{
				fixtures.Token("uumee", "UMEE"),
				osmo,
			},
		}
		require.NoError(t, h(ctx, p))

		// no tokens should have been deleted
		tokens := k.GetAllRegisteredTokens(ctx)
		require.Len(t, tokens, 3)

		_, err := k.GetTokenSettings(ctx, "uumee")
		require.NoError(t, err)

		token, err := k.GetTokenSettings(ctx, "uosmo")
		require.NoError(t, err)
		require.Equal(t, "0.300000000000000000", token.ReserveFactor.String(),
			"reserve factor is correctly set")
	})
}
