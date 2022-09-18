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
		for _, token := range types.DefaultRegistry() {
			require.NoError(t, k.SetTokenSettings(ctx, token))
		}

		modifiedUmee := fixtures.Token("uumee", "UMEE")
		modifiedUmee.ReserveFactor = sdk.MustNewDecFromStr("0.69")

		osmo := fixtures.Token("uosmo", "OSMO")
		p := &types.UpdateRegistryProposal{
			Title:       "test",
			Description: "test",
			Registry: []types.Token{
				modifiedUmee,
				osmo,
			},
		}
		require.NoError(t, h(ctx, p))

		// no tokens should have been deleted
		tokens := k.GetAllRegisteredTokens(ctx)
		require.Len(t, tokens, 3)

		token, err := k.GetTokenSettings(ctx, "uumee")
		require.NoError(t, err)
		require.Equal(t, "0.690000000000000000", token.ReserveFactor.String(),
			"reserve factor is correctly set")

		_, err = k.GetTokenSettings(ctx, "uosmo")
		require.NoError(t, err)

		_, err = k.GetTokenSettings(ctx, fixtures.AtomDenom)
		require.NoError(t, err)
	})
}
