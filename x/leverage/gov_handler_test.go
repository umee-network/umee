package leverage_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/x/leverage"
	"github.com/umee-network/umee/v2/x/leverage/types"
)

func newTestToken(base, symbol, reserveFactor string) types.Token {
	return types.Token{
		BaseDenom:              base,
		SymbolDenom:            symbol,
		Exponent:               6,
		ReserveFactor:          sdk.MustNewDecFromStr(reserveFactor),
		CollateralWeight:       sdk.MustNewDecFromStr("0.25"),
		LiquidationThreshold:   sdk.MustNewDecFromStr("0.25"),
		BaseBorrowRate:         sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:         sdk.MustNewDecFromStr("0.22"),
		MaxBorrowRate:          sdk.MustNewDecFromStr("1.52"),
		KinkUtilization:        sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive:   sdk.MustNewDecFromStr("0.1"),
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdk.MustNewDecFromStr("1.0"),
		MaxSupplyUtilization:   sdk.MustNewDecFromStr("0.9"),
		MinCollateralLiquidity: sdk.MustNewDecFromStr("0.0"),
	}
}

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
				newTestToken("uosmo", "", "0.2"), // empty denom is invalid
			},
		}
		require.Error(t, h(ctx, p))
	})

	t.Run("valid proposal", func(t *testing.T) {
		require.NoError(t, k.SetTokenSettings(ctx,
			newTestToken("uosmo", "OSMO", "0.2"),
		))
		require.NoError(t, k.SetTokenSettings(ctx,
			newTestToken("uatom", "ATOM", "0.2"),
		))

		p := &types.UpdateRegistryProposal{
			Title:       "test",
			Description: "test",
			Registry: []types.Token{
				newTestToken("uumee", "UMEE", "0.2"),
				newTestToken("uosmo", "OSMO", "0.3"),
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
		require.Equal(t, "0.300000000000000000", token.ReserveFactor.String())
	})
}
