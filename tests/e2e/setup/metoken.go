package setup

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/metoken"
)

func (s *E2ETestSuite) TxMetokenSwap(umeeAddr string, asset sdk.Coin, meTokenDenom string) error {
	req := &metoken.MsgSwap{
		User:         umeeAddr,
		Asset:        asset,
		MetokenDenom: meTokenDenom,
	}

	return s.BroadcastTxWithRetry(req)
}

func (s *E2ETestSuite) TxMetokenRedeem(umeeAddr string, meToken sdk.Coin, assetDenom string) error {
	req := &metoken.MsgRedeem{
		User:       umeeAddr,
		Metoken:    meToken,
		AssetDenom: assetDenom,
	}

	return s.BroadcastTxWithRetry(req)
}
