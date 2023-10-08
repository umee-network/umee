package setup

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

type testAccount struct {
	mnemonic string
	keyInfo  keyring.Record
	addr     sdk.AccAddress
}

const (
	// Initial coins to give to validators. Staking only (test txs should use testAccounts instead of validators.
	ValidatorInitBalanceStr = "510000000000" + appparams.BondDenom + ",100000000000" + PhotonDenom + ",100000000000" + mocks.USDTBaseDenom

	// Number of test accounts to initialize in chain.TestAccounts
	numTestAccounts = 3
	// Initial balance of test accounts
	AccountInitBalanceStr = "510000000000" + appparams.BondDenom + ",100000000000" + PhotonDenom + ",100000000000" + mocks.USDTBaseDenom
)

var (
	stakeAmount, _  = sdk.NewIntFromString("100000000000")
	stakeAmountCoin = sdk.NewCoin(appparams.BondDenom, stakeAmount)

	// TODO: reduce
	stakeAmount2, _  = sdk.NewIntFromString("500000000000")
	stakeAmountCoin2 = sdk.NewCoin(appparams.BondDenom, stakeAmount2)
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
	ta.addr, err = info.GetAddress() // TODO: verify this address starts with "umee1"
	if err != nil {
		return err
	}
	c.TestAccounts = append(c.TestAccounts, &ta)
	return nil
}
