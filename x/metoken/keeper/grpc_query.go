package keeper

import (
	"context"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v5/x/metoken"
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

	if err := req.Asset.Validate(); err != nil {
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
	swapFee, err := k.swapFee(index, indexPrices, req.Asset)
	if err != nil {
		return nil, err
	}

	return &metoken.QuerySwapFeeResponse{Asset: swapFee}, nil
}

// RedeemFee returns the fee for the redeem operation, given a specific amount of meTokens and the asset denom.
func (q Querier) RedeemFee(goCtx context.Context, req *metoken.QueryRedeemFee) (
	*metoken.QueryRedeemFeeResponse,
	error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k := q.Keeper(&ctx)

	if err := req.Metoken.Validate(); err != nil {
		return nil, err
	}

	index, err := k.RegisteredIndex(req.Metoken.Denom)
	if err != nil {
		return nil, err
	}

	// get index prices
	indexPrices, err := k.Prices(index)
	if err != nil {
		return nil, err
	}

	// calculate amount to withdraw from x/metoken and x/leverage
	amountFromReserves, amountFromLeverage, err := k.calculateRedeem(index, indexPrices, req.Metoken, req.AssetDenom)
	if err != nil {
		return nil, err
	}

	// calculate the fee for the asset amount that would be given for a redemption
	toRedeem := sdk.NewCoin(req.AssetDenom, amountFromReserves.Add(amountFromLeverage))
	redeemFee, err := k.redeemFee(index, indexPrices, toRedeem)
	if err != nil {
		return nil, err
	}

	return &metoken.QueryRedeemFeeResponse{Asset: redeemFee}, nil
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

	return &metoken.QueryIndexBalancesResponse{
		IndexBalances: balances,
	}, nil
}

// IndexPrice returns Index price from the x/metoken module. If index denom is not specified,
// returns prices for all the registered indexes.
func (q Querier) IndexPrice(
	goCtx context.Context,
	req *metoken.QueryIndexPrice,
) (*metoken.QueryIndexPriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k := q.Keeper(&ctx)

	var indexes []metoken.Index
	if len(req.MetokenDenom) > 0 {
		if !metoken.IsMeToken(req.MetokenDenom) {
			return nil, sdkerrors.ErrInvalidRequest.Wrapf(
				"meToken denom %s should have the following format: me/<TokenName>",
				req.MetokenDenom,
			)
		}

		index, err := k.RegisteredIndex(req.MetokenDenom)
		if err != nil {
			return nil, err
		}

		indexes = []metoken.Index{index}
	} else {
		indexes = k.GetAllRegisteredIndexes()
	}

	var prices []metoken.Price
	for _, index := range indexes {
		ip, err := k.Prices(index)
		if err != nil {
			return nil, err
		}

		price, err := ip.Price(index.Denom)
		if err != nil {
			return nil, err
		}

		prices = append(prices, price)
	}

	return &metoken.QueryIndexPriceResponse{
		Prices: prices,
	}, nil
}

// NewQuerier returns Querier object.
func NewQuerier(kb Builder) Querier {
	return Querier{Builder: kb}
}
