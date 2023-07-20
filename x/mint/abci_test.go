package mint_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gotest.tools/v3/assert"

	umeeapp "github.com/umee-network/umee/v5/app"
	appparams "github.com/umee-network/umee/v5/app/params"
	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/mint"
)

func TestBeginBlock(t *testing.T) {
	app := umeeapp.Setup(t)
	ctx := app.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
	})

	// sdk context should start with current time to make sure to update the inflation min and max rate
	ctx = ctx.WithBlockTime(time.Now())

	oldMintParams := app.MintKeeper.GetParams(ctx)
	uk := app.UGovKeeperB.Keeper(&ctx)

	inflationParams := uk.InflationParams()
	inflationParams.MaxSupply = coin.New(appparams.BondDenom, 21_000000000000000)
	err := uk.SetInflationParams(inflationParams)
	assert.NilError(t, err)

	// Override the mint module BeginBlock
	mint.BeginBlock(ctx, uk, app.MintKeeper)

	// inflation min and max rate should change by reduce rate
	newMintParams := app.MintKeeper.GetParams(ctx)
	liquidationParams := uk.InflationParams()
	inflationReductionRate := liquidationParams.InflationReductionRate.ToDec().Quo(sdk.NewDec(100))
	assert.DeepEqual(t,
		oldMintParams.InflationMax.Mul(sdk.OneDec().Sub(inflationReductionRate)),
		newMintParams.InflationMax,
	)
	assert.DeepEqual(t,
		oldMintParams.InflationMin.Mul(sdk.OneDec().Sub(inflationReductionRate)),
		newMintParams.InflationMin,
	)
}
