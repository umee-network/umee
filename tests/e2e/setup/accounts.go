package setup

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/client"
	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

type testAccount struct {
	mnemonic string
	keyInfo  keyring.Record
	addr     sdk.AccAddress
	client   client.Client
}

var (
	// Initial coins to give to each validator
	valCoins = sdk.NewCoins(
		coin.New(appparams.BaseDenom, 1_000000_000000),
	)

	// Number of test accounts to initialize in chain.TestAccounts
	numTestAccounts = 1

	// Initial coins to give to each test account.
	testAccountCoins = sdk.NewCoins(
		coin.New(appparams.BaseDenom, 100000_000000),
		coin.New(PhotonDenom, 100000_000000),
		coin.New(mocks.USDTBaseDenom, 100000_000000),
	)

	// initial validator stakes (before delegation from test account 0)
	stakeAmountCoin = coin.New(appparams.BaseDenom, 1_000000)
)

// create a test account, which is an address with a mnemonic stored only in memory, to be used with the network.
// these are created randomly each time and added to the suite, so they should be accessed by c.TestAccounts[i >= 0]
// and queries by the desired account's address
func (c *chain) createTestAccount(cdc codec.Codec) error {
	mnemonic, info, err := createMemoryKey(cdc)
	if err != nil {
		return err
	}
	ta := testAccount{}
	ta.keyInfo = *info
	ta.mnemonic = mnemonic
	ta.addr, err = info.GetAddress()
	if err != nil {
		return err
	}
	ta.client, err = c.initDedicatedClient("testAccount"+ta.addr.String(), mnemonic)
	if err != nil {
		return err
	}
	c.TestAccounts = append(c.TestAccounts, &ta)
	return nil
}
