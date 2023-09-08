package client

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/metoken"
)

func (c Client) MetokenQueryClient() metoken.QueryClient {
	return metoken.NewQueryClient(c.Query.GrpcConn)
}

func (c Client) QueryMetokenBalances(denom string) (*metoken.QueryIndexBalancesResponse, error) {
	ctx, cancel := c.Query.NewCtx()
	defer cancel()

	return c.MetokenQueryClient().IndexBalances(ctx, &metoken.QueryIndexBalances{MetokenDenom: denom})
}

func (c Client) QueryMetokenIndexes(denom string) (*metoken.QueryIndexesResponse, error) {
	ctx, cancel := c.Query.NewCtx()
	defer cancel()

	return c.MetokenQueryClient().Indexes(ctx, &metoken.QueryIndexes{MetokenDenom: denom})
}

func (c Client) TxMetokenSwap(
	umeeAddr sdk.AccAddress,
	asset sdk.Coin,
	meTokenDenom string,
) error {
	return c.broadcastWithRetry(metoken.NewMsgSwap(umeeAddr, asset, meTokenDenom))
}

func (c Client) TxMetokenRedeem(
	umeeAddr sdk.AccAddress,
	meToken sdk.Coin,
	assetDenom string,
) error {
	return c.broadcastWithRetry(metoken.NewMsgRedeem(umeeAddr, meToken, assetDenom))
}

func (c Client) broadcastWithRetry(msg sdk.Msg) error {
	var err error
	for retry := 0; retry < 3; retry++ {
		// retry if txs fails, because sometimes account sequence mismatch occurs due to txs pending
		if _, err = c.Tx.BroadcastTx([]sdk.Msg{msg}...); err == nil {
			break
		}
		time.Sleep(time.Millisecond * 300)
	}

	return err
}
