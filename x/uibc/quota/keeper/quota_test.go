package keeper

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	ibcutil "github.com/umee-network/umee/v6/util/ibc"
)

func TestUnitGetQuotas(t *testing.T) {
	k := initKeeperSimpleMock(t)

	quotas, err := k.GetAllOutflows()
	assert.NilError(t, err)
	assert.Equal(t, len(quotas), 0)

	setQuotas := sdk.DecCoins{sdk.NewInt64DecCoin("test_uumee", 10000)}

	k.SetTokenOutflows(setQuotas)
	quotas, err = k.GetAllOutflows()
	assert.NilError(t, err)
	assert.DeepEqual(t, setQuotas, quotas)

	// get the quota of denom
	quota := k.GetTokenOutflows(setQuotas[0].Denom)
	assert.Equal(t, quota.Denom, setQuotas[0].Denom)
}

func TestUnitGetLocalDenom(t *testing.T) {
	out := ibcutil.GetLocalDenom("umee")
	assert.Equal(t, "umee", out)
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
	k.SetTotalInflow(sdk.NewDec(50))
	k.SetTokenInflow(sdk.NewDecCoin(umee, math.NewInt(50)))

	err := k.CheckAndUpdateQuota(umee, sdk.NewInt(1))
	assert.NilError(t, err)
	k.checkOutflows(umee, 8, 52)

	// transferring 2 umee => 4USD, will exceed the quota (8+4 > 10)
	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(2))
	assert.ErrorContains(t, err, "quota")
	k.checkOutflows(umee, 8, 52)

	// transferring 1 umee => 2USD, will will be still OK (8+2 <= 10)
	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(1))
	assert.NilError(t, err)
	k.checkOutflows(umee, 10, 54)

	// 2. Setting TokenQuota param to 0 should unlimit the token quota check
	//
	k.setQuotaParams(0, 100)

	// transferring 20 umee => 40USD, will skip the token quota check, but will update outflows
	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(20))
	assert.NilError(t, err)
	k.checkOutflows(umee, 50, 94)

	// transferring additional 5 umee => 10USD, will fail total quota check but it will pass inflow quota check
	// sum of outflows <= $1M +  params.InflowOutflowQuotaRate * sum of all inflows =  (10_000_000)+  (50*0) = 10_000_000
	// 104 <= 10_000_000
	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(5))
	assert.ErrorContains(t, err, "quota")
	k.checkOutflows(umee, 50, 94)

	// it will fail total quota check and inflow quota check also
	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(5000000000))
	assert.ErrorContains(t, err, "quota")
	k.checkOutflows(umee, 50, 94)

	// 3. Setting TotalQuota param to 0 should unlimit the total quota check
	//
	k.setQuotaParams(0, 0)
	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(5))
	assert.NilError(t, err)
	k.checkOutflows(umee, 60, 104)

	// 4. Setting TokenQuota to 65
	//
	k.setQuotaParams(65, 0)
	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(1))
	assert.NilError(t, err)
	k.checkOutflows(umee, 62, 106)

	err = k.CheckAndUpdateQuota(umee, sdk.NewInt(2)) // exceeds token quota
	assert.ErrorContains(t, err, "quota")
}

func TestUnitGetExchangePrice(t *testing.T) {
	k := initKeeperSimpleMock(t)
	p, err := k.getExchangePrice(umee, sdk.NewInt(12))
	assert.NilError(t, err)
	assert.DeepEqual(t, sdk.NewDec(24), p)

	// ATOM is leverage registered token but price is not avaiable
	p, err = k.getExchangePrice(atom, sdk.NewInt(3))
	assert.NilError(t, err)
	assert.DeepEqual(t, sdk.ZeroDec(), p)

	_, err = k.getExchangePrice("notexisting", sdk.NewInt(10))
	assert.ErrorContains(t, err, "not found")
}

func TestSetAndGetIBCInflows(t *testing.T) {
	k := initKeeperSimpleMock(t)
	inflowSum := sdk.MustNewDecFromStr("123123")
	k.SetTotalInflow(inflowSum)

	rv := k.GetTotalInflow()
	assert.DeepEqual(t, inflowSum, rv)

	// inflow of token
	inflowOfToken := sdk.NewDecCoin("abcd", math.NewInt(1000000))
	k.SetTokenInflow(inflowOfToken)

	val := k.GetTokenInflow(inflowOfToken.Denom)
	assert.DeepEqual(t, val, inflowOfToken)

	inflows, err := k.GetAllInflows()
	assert.NilError(t, err)
	assert.DeepEqual(t, inflows[0], inflowOfToken)
}
