package e2e

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/assert"

	"github.com/umee-network/umee/v6/tests/accs"
	setup "github.com/umee-network/umee/v6/tests/e2e/setup"
	"github.com/umee-network/umee/v6/tests/tsdk"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/uibc"
)

func (s *E2ETest) testIBCTokenTransferWithMemo(umeeAPIEndpoint string, atomQuota sdk.Coin) {
	assert := assert.New(s.T())
	totalSupply, err := s.QueryTotalSupply(umeeAPIEndpoint)
	s.T().Logf("total supply : %s", totalSupply.String())
	prevIBCAtomBalance := totalSupply.AmountOf(uatomIBCHash)
	s.T().Logf("total balance of IBC ATOM : %s", prevIBCAtomBalance.String())

	//<<<< Valid MEMO  : gaia -> umee >>
	atomFromGaia := mulCoin(atomQuota, "5.0")
	atomFromGaia.Denom = "uatom"

	atomIBCDenom := atomFromGaia
	atomIBCDenom.Denom = uatomIBCHash
	cdc := tsdk.NewCodec(uibc.RegisterInterfaces, ltypes.RegisterInterfaces)

	// INVALID MEMO : with fallback_addr
	//  Collteralize msg is not supported
	msgCollateralize := []sdk.Msg{
		ltypes.NewMsgCollateralize(accs.Alice, atomIBCDenom),
	}
	anyMsgOfCollateralize, err := tx.SetMsgs(msgCollateralize)
	assert.Nil(err)
	fallbackAddr := "umee1mjk79fjjgpplak5wq838w0yd982gzkyf3qjpef"
	invalidMemo := uibc.ICS20Memo{Messages: anyMsgOfCollateralize, FallbackAddr: fallbackAddr}

	// fallback_addr balance
	iniBalance, err := s.QueryUmeeDenomBalance(umeeAPIEndpoint, fallbackAddr, uatomIBCHash)
	assert.Equal(true, iniBalance.Amount.Equal(math.ZeroInt()))

	invalidMemoBZ, err := cdc.MarshalJSON(&invalidMemo)
	assert.Nil(err)
	s.SendIBC(setup.GaiaChainID, s.Chain.ID, accs.Alice.String(), atomFromGaia, "", string(invalidMemoBZ), "")
	updatedIBCAtomBalance := atomFromGaia.Amount.Add(prevIBCAtomBalance)
	s.checkSupply(umeeAPIEndpoint, uatomIBCHash, updatedIBCAtomBalance)
	s.checkLeverageAccountBalance(umeeAPIEndpoint, fallbackAddr, uatomIBCHash, math.ZeroInt())
	// fallback_addr has to get the sending amount
	bAmount, err := s.QueryUmeeDenomBalance(umeeAPIEndpoint, fallbackAddr, uatomIBCHash)
	assert.Equal(true, atomIBCDenom.Equal(bAmount))
	// receiver doesn't receive the sending amount because due to invalid memo , recv address is override by fallback_addr
	recvAmount, err := s.QueryUmeeDenomBalance(umeeAPIEndpoint, accs.Alice.String(), uatomIBCHash)
	assert.Equal(true, recvAmount.Amount.Equal(math.ZeroInt()))

	// INVALID MEMO : without fallback_addr
	// receiver has to get the sending amount
	invalidMemo = uibc.ICS20Memo{Messages: anyMsgOfCollateralize, FallbackAddr: ""}
	invalidMemoBZ, err = cdc.MarshalJSON(&invalidMemo)
	assert.Nil(err)
	s.SendIBC(setup.GaiaChainID, s.Chain.ID, accs.Alice.String(), atomFromGaia, "", string(invalidMemoBZ), "")
	updatedIBCAtomBalance = updatedIBCAtomBalance.Add(atomFromGaia.Amount)
	s.checkSupply(umeeAPIEndpoint, uatomIBCHash, updatedIBCAtomBalance)
	s.checkLeverageAccountBalance(umeeAPIEndpoint, fallbackAddr, uatomIBCHash, math.ZeroInt())
	// fallback_addr doesn't get the sending amount
	bAmount, err = s.QueryUmeeDenomBalance(umeeAPIEndpoint, fallbackAddr, uatomIBCHash)
	// same as previous amount (already fallback_addr have the amount)
	assert.Equal(true, atomIBCDenom.Equal(bAmount))
	// receiver has to  receive the sending amount
	recvAmount, err = s.QueryUmeeDenomBalance(umeeAPIEndpoint, accs.Alice.String(), uatomIBCHash)
	assert.Equal(true, atomIBCDenom.Equal(recvAmount))

	// VALID MEMO : without fallback_addr
	msgs := []sdk.Msg{
		ltypes.NewMsgSupplyCollateral(accs.Alice, atomIBCDenom),
	}
	anyMsg, err := tx.SetMsgs(msgs)
	assert.Nil(err)
	memo := uibc.ICS20Memo{Messages: anyMsg}

	bz, err := cdc.MarshalJSON(&memo)
	assert.Nil(err)
	s.SendIBC(setup.GaiaChainID, s.Chain.ID, accs.Alice.String(), atomFromGaia, "", string(bz), "")
	updatedIBCAtomBalance = updatedIBCAtomBalance.Add(atomFromGaia.Amount)
	s.checkSupply(umeeAPIEndpoint, uatomIBCHash, updatedIBCAtomBalance)
	s.checkLeverageAccountBalance(umeeAPIEndpoint, accs.Alice.String(), uatomIBCHash, atomFromGaia.Amount)
}
