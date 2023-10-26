package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	ibcutil "github.com/umee-network/umee/v6/util/ibc"
)

func TestUnitGetQuotas(t *testing.T) {
	k := initKeeperSimpleMock(t)

	quotas, err := k.GetAllOutflows()
	require.NoError(t, err)
	require.Equal(t, len(quotas), 0)

	setQuotas := sdk.DecCoins{sdk.NewInt64DecCoin("test_uumee", 10000)}

	k.SetTokenOutflows(setQuotas)
	quotas, err = k.GetAllOutflows()
	require.NoError(t, err)
	require.Equal(t, setQuotas, quotas)

	// get the quota of denom
	quota := k.GetTokenOutflows(setQuotas[0].Denom)
	require.Equal(t, quota.Denom, setQuotas[0].Denom)
}

func TestUnitGetLocalDenom(t *testing.T) {
	out := ibcutil.GetLocalDenom("umee")
	require.Equal(t, "umee", out)
}

func TestUnitCheckAndUpdateQuota(t *testing.T) {
	k := initKeeperSimpleMock(t)

	// initUmeeKeeper sets umee price: 2usd

	// 1. We set the quota param to 10 and existing outflow sum to 6 USD.
	// Transfer of 2 USD in Umee should work, but additional transfer of 4 USD should fail.
	//
	k.setQuotaParams(10, 100)
	k.SetTokenOutflow(sdk.NewInt64DecCoin(umee, 6))
	k.SetTokenInflow(sdk.NewInt64DecCoin(umee, 6))
	k.SetTotalOutflowSum(sdk.NewDec(50))
	k.SetTotalInflowSum(sdk.NewDec(50))

	err := k.CheckAndUpdateQuota(umee, sdk.NewInt(1))
	require.NoError(t, err)
	k.checkOutflows(umee, 8, 52)

	// transferring 2 umee => 4USD, will exceed the quota (8+4 > 10)
	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(2))
	require.ErrorContains(t, err, "quota")
	k.checkOutflows(umee, 8, 52)

	// transferring 1 umee => 2USD, will will be still OK (8+2 <= 10)
	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(1))
	require.NoError(t, err)
	k.checkOutflows(umee, 10, 54)

	// 2. Setting TokenQuota param to 0 should unlimit the token quota check
	//
	k.setQuotaParams(0, 100)

	// transferring 20 umee => 40USD, will skip the token quota check, but will update outflows
	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(20))
	require.NoError(t, err)
	k.checkOutflows(umee, 50, 94)

	// transferring additional 5 umee => 10USD, will fail total quota check
	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(5))
	require.ErrorContains(t, err, "quota")
	k.checkOutflows(umee, 50, 94)

	// 3. Setting TotalQuota param to 0 should unlimit the total quota check
	//
	k.setQuotaParams(0, 0)
	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(5))
	require.NoError(t, err)
	k.checkOutflows(umee, 60, 104)

	// 4. Setting TokenQuota to 65
	//
	k.setQuotaParams(65, 0)
	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(1))
	require.NoError(t, err)
	k.checkOutflows(umee, 62, 106)

	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(2)) // exceeds token quota
	require.ErrorContains(t, err, "quota")
}

func TestUnitGetExchangePrice(t *testing.T) {
	k := initKeeperSimpleMock(t)
	p, err := k.getExchangePrice(umee, sdk.NewInt(12))
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(24), p)

	p, err = k.getExchangePrice(atom, sdk.NewInt(3))
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(30), p)

	_, err = k.getExchangePrice("notexisting", sdk.NewInt(10))
	require.ErrorContains(t, err, "not found")
}
