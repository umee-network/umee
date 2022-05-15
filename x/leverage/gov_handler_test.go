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
		require.NoError(t, k.SetRegisteredToken(ctx, types.Token{
			BaseDenom:            "uosmo",
			SymbolDenom:          "OSMO",
			Exponent:             6,
			ReserveFactor:        sdk.MustNewDecFromStr("0.20"),
			CollateralWeight:     sdk.MustNewDecFromStr("0.25"),
			LiquidationThreshold: sdk.MustNewDecFromStr("0.25"),
			BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
			KinkBorrowRate:       sdk.MustNewDecFromStr("0.22"),
			MaxBorrowRate:        sdk.MustNewDecFromStr("1.52"),
			KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
			LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
			EnableLend:           true,
			EnableBorrow:         true,
			Blacklist:            false,
		}))
		require.NoError(t, k.SetRegisteredToken(ctx, types.Token{
			BaseDenom:            "uatom",
			SymbolDenom:          "ATOM",
			Exponent:             6,
			ReserveFactor:        sdk.MustNewDecFromStr("0.20"),
			CollateralWeight:     sdk.MustNewDecFromStr("0.25"),
			LiquidationThreshold: sdk.MustNewDecFromStr("0.25"),
			BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
			KinkBorrowRate:       sdk.MustNewDecFromStr("0.22"),
			MaxBorrowRate:        sdk.MustNewDecFromStr("1.52"),
			KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
			LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
			EnableLend:           true,
			EnableBorrow:         true,
			Blacklist:            false,
		}))

		p := &types.UpdateRegistryProposal{
			Title:       "test",
			Description: "test",
			Registry: []types.Token{
				{
					BaseDenom:            umeeapp.BondDenom,
					SymbolDenom:          "UMEE",
					Exponent:             6,
					ReserveFactor:        sdk.MustNewDecFromStr("0.20"),
					CollateralWeight:     sdk.MustNewDecFromStr("0.25"),
					LiquidationThreshold: sdk.MustNewDecFromStr("0.25"),
					BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
					KinkBorrowRate:       sdk.MustNewDecFromStr("0.22"),
					MaxBorrowRate:        sdk.MustNewDecFromStr("1.52"),
					KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
					LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
					EnableLend:           true,
					EnableBorrow:         true,
					Blacklist:            false,
				},
				{
					BaseDenom:            "uosmo",
					SymbolDenom:          "OSMO",
					Exponent:             6,
					ReserveFactor:        sdk.MustNewDecFromStr("0.20"),
					CollateralWeight:     sdk.MustNewDecFromStr("0.25"),
					LiquidationThreshold: sdk.MustNewDecFromStr("0.25"),
					BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
					KinkBorrowRate:       sdk.MustNewDecFromStr("0.22"),
					MaxBorrowRate:        sdk.MustNewDecFromStr("1.52"),
					KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
					LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
					EnableLend:           true,
					EnableBorrow:         true,
					Blacklist:            false,
				},
			},
		}
		require.NoError(t, h(ctx, p))

		tokens := k.GetAllRegisteredTokens(ctx)
		require.Len(t, tokens, 3)

		// ensure that new tokens was added
		_, err := k.GetRegisteredToken(ctx, "uumee")
		require.NoError(t, err)

		// ensure that unaffected token was not deleted by the gov proposal
		token, err := k.GetRegisteredToken(ctx, "uatom")
		require.NoError(t, err)

		// ensure that the existing token was updated by the gov proposal
		token, err = k.GetRegisteredToken(ctx, "uosmo")
		require.NoError(t, err)
		require.Equal(t, "0.020000000000000000", token.BaseBorrowRate.String())
	})
}
