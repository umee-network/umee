package setup

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/metoken"
)

func (s *E2ETestSuite) QueryMetokenBalances(denom string) (metoken.QueryIndexBalancesResponse, error) {
	endpoint := fmt.Sprintf("%s/umee/metoken/v1/index_balances?metoken_denom=%s", s.UmeeREST(), denom)
	var resp metoken.QueryIndexBalancesResponse

	return resp, s.QueryREST(endpoint, &resp)
}

func (s *E2ETestSuite) QueryMetokenIndexes(denom string) (metoken.QueryIndexesResponse, error) {
	endpoint := fmt.Sprintf("%s/umee/metoken/v1/indexes?metoken_denom=%s", s.UmeeREST(), denom)
	var resp metoken.QueryIndexesResponse

	return resp, s.QueryREST(endpoint, &resp)
}

func (s *E2ETestSuite) TxMetokenSwap(umeeAddr string, asset sdk.Coin, meTokenDenom string) error {
	req := &metoken.MsgSwap{
		User:         umeeAddr,
		Asset:        asset,
		MetokenDenom: meTokenDenom,
	}

	return s.broadcastTxWithRetry(req)
}

func (s *E2ETestSuite) TxMetokenRedeem(umeeAddr string, meToken sdk.Coin, assetDenom string) error {
	req := &metoken.MsgRedeem{
		User:       umeeAddr,
		Metoken:    meToken,
		AssetDenom: assetDenom,
	}

	return s.broadcastTxWithRetry(req)
}
