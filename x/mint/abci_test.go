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
	mint.BeginBlock(ctx, uk, app.MintKeeper)

	// inflation min and max rate should change by reduce rate
	newMintParams := app.MintKeeper.GetParams(ctx)
	liquidationParams := uk.InflationParams()
	assert.DeepEqual(t,
		oldMintParams.InflationMax.Mul(sdk.OneDec().Sub(liquidationParams.InflationReductionRate)),
		newMintParams.InflationMax,
	)
	assert.DeepEqual(t,
		oldMintParams.InflationMin.Mul(sdk.OneDec().Sub(liquidationParams.InflationReductionRate)),
		newMintParams.InflationMin,
	)
}
