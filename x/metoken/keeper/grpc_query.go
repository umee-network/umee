package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v6/x/metoken"
)

var _ metoken.QueryServer = Querier{}

// Querier implements a QueryServer for the x/metoken module.
type Querier struct {
	Builder
}

// Params returns params of the x/metoken module.
func (q Querier) Params(goCtx context.Context, _ *metoken.QueryParams) (*metoken.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper(&ctx).GetParams()

	return &metoken.QueryParamsResponse{Params: params}, nil
}

// Indexes returns registered indexes in the x/metoken module. If index denom is not specified,
// returns all the registered indexes.
func (q Querier) Indexes(goCtx context.Context, req *metoken.QueryIndexes) (*metoken.QueryIndexesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k := q.Keeper(&ctx)

	var indexes []metoken.Index
	if len(req.MetokenDenom) == 0 {
		indexes = k.GetAllRegisteredIndexes()
	} else {
		index, err := k.RegisteredIndex(req.MetokenDenom)
		if err != nil {
			return nil, err
		}
		indexes = []metoken.Index{index}
	}

	return &metoken.QueryIndexesResponse{
		Registry: indexes,
	}, nil
}

// SwapFee returns the fee for the swap operation, given a specific amount of tokens and the meToken denom.
func (q Querier) SwapFee(goCtx context.Context, req *metoken.QuerySwapFee) (*metoken.QuerySwapFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k := q.Keeper(&ctx)

	asset, err := sdk.ParseCoinNormalized(req.Asset)
	if err != nil {
		return nil, err
	}

	if err := asset.Validate(); err != nil {
		return nil, err
	}

	index, err := k.RegisteredIndex(req.MetokenDenom)
	if err != nil {
		return nil, err
	}

	// get index prices
	indexPrices, err := k.Prices(index)
	if err != nil {
		return nil, err
	}

	// calculate the fee for the asset amount
	_, feeAmount, err := k.swapFee(index, indexPrices, asset)
	if err != nil {
		return nil, err
	}

	return &metoken.QuerySwapFeeResponse{Asset: feeAmount}, nil
}

// RedeemFee returns the fee for the redeem operation, given a specific amount of meTokens and the asset denom.
func (q Querier) RedeemFee(goCtx context.Context, req *metoken.QueryRedeemFee) (
	*metoken.QueryRedeemFeeResponse,
	error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k := q.Keeper(&ctx)

	meToken, err := sdk.ParseCoinNormalized(req.Metoken)
	if err != nil {
		return nil, err
	}

	if err := meToken.Validate(); err != nil {
		return nil, err
	}

	index, err := k.RegisteredIndex(meToken.Denom)
	if err != nil {
		return nil, err
	}

	// get index prices
	indexPrices, err := k.Prices(index)
	if err != nil {
		return nil, err
	}

	// calculate amount to withdraw from x/metoken and x/leverage
	amountFromReserves, amountFromLeverage, err := k.calculateRedeem(index, indexPrices, meToken, req.AssetDenom)
	if err != nil {
		return nil, err
	}

	// calculate the fee for the asset amount that would be given for a redemption
	toRedeem := sdk.NewCoin(req.AssetDenom, amountFromReserves.Add(amountFromLeverage))
	_, feeAmount, err := k.redeemFee(index, indexPrices, toRedeem)
	if err != nil {
		return nil, err
	}

	return &metoken.QueryRedeemFeeResponse{Asset: feeAmount}, nil
}

// IndexBalances returns balances from the x/metoken module. If index balance denom is not specified,
// returns all the balances.
func (q Querier) IndexBalances(
	goCtx context.Context,
	req *metoken.QueryIndexBalances,
) (*metoken.QueryIndexBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k := q.Keeper(&ctx)

	var balances []metoken.IndexBalances
	if len(req.MetokenDenom) == 0 {
		balances = k.GetAllIndexesBalances()
	} else {
		balance, err := k.IndexBalances(req.MetokenDenom)
		if err != nil {
			return nil, err
		}
		balances = []metoken.IndexBalances{balance}
	}

	prices, err := q.getPrices(k, req.MetokenDenom)
	if err != nil {
		return nil, err
	}

	return &metoken.QueryIndexBalancesResponse{
		IndexBalances: balances,
		Prices:        prices,
	}, nil
}

// IndexPrices returns Index price from the x/metoken module. If index denom is not specified,
// returns prices for all the registered indexes.
func (q Querier) IndexPrices(
	goCtx context.Context,
	req *metoken.QueryIndexPrices,
) (*metoken.QueryIndexPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	prices, err := q.getPrices(q.Keeper(&ctx), req.MetokenDenom)
	if err != nil {
		return nil, err
	}

	return &metoken.QueryIndexPricesResponse{
		Prices: prices,
	}, nil
}

func (q Querier) getPrices(k Keeper, meTokenDenom string) ([]metoken.IndexPrices, error) {
	var indexes []metoken.Index
	if len(meTokenDenom) > 0 {
		if !metoken.IsMeToken(meTokenDenom) {
			return nil, sdkerrors.ErrInvalidRequest.Wrapf(
				"meToken denom %s should have the following format: me/<TokenName>",
				meTokenDenom,
			)
		}

		index, err := k.RegisteredIndex(meTokenDenom)
		if err != nil {
			return nil, err
		}

		indexes = []metoken.Index{index}
	} else {
		indexes = k.GetAllRegisteredIndexes()
	}

	prices := make([]metoken.IndexPrices, len(indexes))
	for i, index := range indexes {
		ip, err := k.Prices(index)
		if err != nil {
			return nil, err
		}

		prices[i] = ip
	}

	return prices, nil
}

// NewQuerier returns Querier object.
func NewQuerier(kb Builder) Querier {
	return Querier{Builder: kb}
}
