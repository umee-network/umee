package e2e

// type TestCosmwasm struct {
// 	IntegrationTestSuite
// 	util.Cosmwasm
// }

/*
// TODO : needs to setup live network with dockernet
func TestCosmwasmSuite(t *testing.T) {
	suite.Run(t, new(TestCosmwasm))
}

// TODO: re-enable this tests when we do dockernet integration
func (cw *TestCosmwasm) TestCW20() {
	// TODO: needs to add contracts
	accAddr, err := cw.chain.validators[0].keyInfo.GetAddress()
	cw.Require().NoError(err)
	cw.Sender = accAddr.String()

	// path := ""
	path := "/Users/gsk967/Projects/umee-network/umee-cosmwasm/artifacts/umee_cosmwasm-aarch64.wasm"
	cw.DeployWasmContract(path)

	// InstantiateContract
	cw.InstantiateContract()

	// execute contract
	tx := "{\"umee\":{\"leverage\":{\"supply_collateral\":{\"supplier\":\"umee19ppq83qzzy3f0fftdp2p3t5eyp44nm33we37n3\",\"asset\":{\"amount\":\"1000\",\"denom\":\"uumee\"}}}}}"
	cw.CWExecute(tx)

	// query the contract
	query := "{\"chain\":{\"custom\":{\"leverage_params\":{},\"assigned_query\":\"0\"}}}"
	cw.CWQuery(query)
	cw.Require().False(true)
}

*/
