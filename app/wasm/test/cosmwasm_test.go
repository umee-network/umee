package wasm_test

import "testing"

func TestCosmwasm(t *testing.T) {
	its := new(IntegrationTestSuite)
	// setup the test configuration
	its.SetupTest(t)

	// testing cw_plus base contract
	its.TestCw20Store()
	its.TestCw20Instantiate()
	its.TestCw20ContractInfo()
	its.TestCw20CheckBalance()
	its.TestCw20Transfer()

	// testing the umee cosmwasm queries
	its.InitiateUmeeCosmwasm()
	its.TestLeverageQueries()
	its.TestOracleQueries()
	its.TestLeverageTxs()
}
