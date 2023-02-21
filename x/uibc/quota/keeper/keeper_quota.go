package keeper

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"

	"github.com/umee-network/umee/v4/util/coin"
	ltypes "github.com/umee-network/umee/v4/x/leverage/types"
	"github.com/umee-network/umee/v4/x/uibc"
)

var (
	denomQuotaKey = []byte{1}
)

type outflowKeeper struct {
	oracle uibc.OracleKeeper
	store  sdk.KVStore
}

func (k outflowKeeper) checkAndUpdateQuota(ctx sdk.Context, denom string, newOutflow sdkmath.Int, params uibc.Params) error {
	quota, err := k.GetQuota(ctx, denom)
	if err != nil {
		return err
	}

	price, err := k.oracle.HistoricAvgPrice(ctx, denom)
	if err != nil {
		if ltypes.ErrNotRegisteredToken.Is(err) {
			return nil
		} else if err != nil {
			return err
		}
	}

	// checking ibc-transfer token quota
	quota.Amount = quota.Amount.Add(price)
	if quota.Amount.GT(params.TokenQuota) {
		return uibc.ErrQuotaExceeded
	}

	// update the per token outflow sum
	return k.SetDenomQuota(ctx, quota)
}

func (k outflowKeeper) GetQuota(ctx sdk.Context, denom string) (sdk.DecCoin, error) {
	bz := k.store.Get(denomQuotaKey)
	var q sdk.DecCoin
	err := q.Unmarshal(bz)
	return q, err
}

func (k outflowKeeper) SetDenomQuota(ctx sdk.Context, d sdk.DecCoin) error {
	bz, err := d.Marshal()
	if err != nil {
		return err
	}
	k.store.Set(denomQuotaKey, bz)
	return nil
}
