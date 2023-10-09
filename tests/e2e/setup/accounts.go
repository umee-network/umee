package setup

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

type testAccount struct {
	mnemonic string
	keyInfo  keyring.Record
	addr     sdk.AccAddress
}

var (
	// Initial coins to give to validator 0 (which it uses to fund test accounts)
	valCoins = sdk.NewCoins(
		coin.New(appparams.BondDenom, 1_000000_000000),
		coin.New(PhotonDenom, 1_000000_000000),
		coin.New(mocks.USDTBaseDenom, 1_000000_000000),
	)

	// TODO: stake less on the validators, and instead delegate from a test account
	stakeAmountCoin  = coin.New(appparams.BondDenom, 1_000000)
	stakeAmountCoin2 = coin.New(appparams.BondDenom, 5_000000)
)
