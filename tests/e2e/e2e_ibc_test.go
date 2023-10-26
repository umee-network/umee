package e2e

import (
	"fmt"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v6/app/params"
	setup "github.com/umee-network/umee/v6/tests/e2e/setup"
	"github.com/umee-network/umee/v6/tests/grpc"
	"github.com/umee-network/umee/v6/x/uibc"
)

const (
	// ibc hash of gaia stake token on umee
	stakeIBCHash = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
	// ibc hash of uumee token on gaia
	umeeIBCHash = "ibc/9F53D255F5320A4BE124FF20C29D46406E126CE8A09B00CA8D3CFF7905119728"
	// ibc hash of uatom token on umee
	uatomIBCHash = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"

	atomSymbol = "ATOM"
	umeeSymbol = "UMEE"
)

var powerReduction = sdk.MustNewDecFromStr("10").Power(6)

// mulCoin multiplies the amount of a coin by a dec (given as string)
func mulCoin(c sdk.Coin, d string) sdk.Coin {
	newAmount := sdk.MustNewDecFromStr(d).MulInt(c.Amount).RoundInt()
	return sdk.NewCoin(c.Denom, newAmount)
}

func (s *E2ETest) checkOutflowByPercentage(endpoint, excDenom string, outflow, amount, perDiff sdk.Dec) {
	// get historic average price for denom (SYMBOL_DENOM)
	histoAvgPrice, err := s.QueryHistAvgPrice(endpoint, excDenom)
	s.Require().NoError(err)
	totalPrice := amount.Quo(powerReduction).Mul(histoAvgPrice)
	s.T().Log("exchangeRate total price ", totalPrice.String(), "outflow value", outflow.String())
	percentageDiff := totalPrice.Mul(perDiff)
	// Note: checking outflow >= total_price with percentageDiff
	// either total_price >= outflow with percentageDiff
	s.Require().True(outflow.GTE(totalPrice.Sub(percentageDiff)) || totalPrice.GTE(outflow.Sub(percentageDiff)))
}

func (s *E2ETest) checkOutflows(umeeAPIEndpoint, denom string, checkWithExcRate bool, amount sdk.Dec, excDenom string) {
	s.Require().Eventually(
		func() bool {
			a, err := s.QueryOutflows(umeeAPIEndpoint, denom)
			if err != nil {
				return false
			}
			if checkWithExcRate {
				s.checkOutflowByPercentage(umeeAPIEndpoint, excDenom, a, amount, sdk.MustNewDecFromStr("0.01"))
			}
			return true
		},
		30*time.Second,
		500*time.Millisecond,
	)
}

func (s *E2ETest) checkSupply(endpoint, ibcDenom string, amount math.Int) {
	actualSupply := math.ZeroInt()
	var err error
	s.Require().Eventually(
		func() bool {
			var supply sdk.Coins
			supply, err = s.QueryTotalSupply(endpoint)
			if err != nil {
				return false
			}
			actualSupply = supply.AmountOf(ibcDenom)
			return actualSupply.Equal(amount)
		},
		2*time.Minute,
		1*time.Second,
		"check supply: %s (expected %s, actual %s) err: %v", ibcDenom, amount, actualSupply, err,
	)
}

func (s *E2ETest) TestIBCTokenTransfer() {
	// IBC inbound transfer of non x/leverage registered tokens must fail, because
	// because we won't have price for it.
	s.Run("send_stake_to_umee", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		umeeAPIEndpoint := s.UmeeREST()
		recipient := s.AccountAddr(0).String()

		token := sdk.NewInt64Coin("stake", 3300000000) // 3300stake
		s.SendIBC(setup.GaiaChainID, s.Chain.ID, recipient, token, false, "")
		// Zero, since not a registered token
		s.checkSupply(umeeAPIEndpoint, stakeIBCHash, sdk.ZeroInt())
	})

	s.Run("ibc_transfer_quota", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		gaiaAPIEndpoint := s.GaiaREST()
		umeeAPIEndpoint := s.UmeeREST()
		// totalQuota := math.NewInt(120)
		tokenQuota := math.NewInt(100)

		var atomPrice math.LegacyDec
		// compute the amount of ATOM sent from umee to gaia which would meet atom's token quota
		s.Require().Eventually(func() bool {
			var err error
			atomPrice, err = s.QueryHistAvgPrice(umeeAPIEndpoint, atomSymbol)
			if err != nil {
				return false
			}
			return atomPrice.GT(sdk.OneDec())
		},
			2*time.Minute,
			1*time.Second,
			"price of atom should be greater than 1",
		)

		atomQuota := sdk.NewCoin(uatomIBCHash,
			sdk.NewDecFromInt(tokenQuota).Quo(atomPrice).Mul(powerReduction).RoundInt(),
		)

		//<<<< INFLOW : gaia -> umee >>
		// send $500 ATOM from gaia to umee. (ibc_quota will not check token limit)
		atomFromGaia := mulCoin(atomQuota, "5.0")
		atomFromGaia.Denom = "uatom"
		s.SendIBC(setup.GaiaChainID, s.Chain.ID, "", atomFromGaia, false, "")
		s.checkSupply(umeeAPIEndpoint, uatomIBCHash, atomFromGaia.Amount)

		// <<< OUTLOW : umee -> gaia >>
		// compute the amout of UMEE sent to gaia which would meet umee's token quota
		umeePrice, err := s.QueryHistAvgPrice(umeeAPIEndpoint, umeeSymbol)
		s.Require().NoError(err)
		s.Require().True(umeePrice.GT(sdk.MustNewDecFromStr("0.001")),
			"umee price should be non zero, and expecting higher than 0.001, got: %s", umeePrice)
		umeeQuota := sdk.NewCoin(appparams.BondDenom,
			sdk.NewDecFromInt(tokenQuota).Quo(umeePrice).Mul(powerReduction).RoundInt(),
		)

		// << TOKEN QUOTA EXCCEED >>
		// send $110 UMEE from umee to gaia (token_quota is 100$)
		exceedUmee := mulCoin(umeeQuota, "1.1")
		s.SendIBC(s.Chain.ID, setup.GaiaChainID, "", exceedUmee, true, "")
		// check the ibc (umee) quota after ibc txs - this one should have failed
		// supply don't change
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, math.ZeroInt())

		// send $110 ATOM from umee to gaia
		exceedAtom := mulCoin(atomQuota, "1.1")
		// supply will be not be decreased because sending amount is more than token quota so it will fail
		s.SendIBC(s.Chain.ID, setup.GaiaChainID, "", exceedAtom, true, "uatom from umee to gaia")
		s.checkSupply(umeeAPIEndpoint, uatomIBCHash, atomFromGaia.Amount)

		// << BELOW TOKEN QUOTA >>
		// send $90 UMEE from umee to gaia (ibc_quota will check)
		// Note: receiver is null so hermes will default send to key_name (from config) of target chain (gaia)
		sendUmee := mulCoin(umeeQuota, "0.9")
		s.SendIBC(s.Chain.ID, setup.GaiaChainID, "", sendUmee, false, fmt.Sprintf(
			"sending %s (less than token quota) ", sendUmee.String()))
		s.checkOutflows(umeeAPIEndpoint, appparams.BondDenom, true, sdk.NewDecFromInt(sendUmee.Amount), appparams.Name)
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, sendUmee.Amount)

		// << BELOW TOKEN QUOTA 40$ but ATOM_QUOTA (40$)+ UMEE_QUOTA(90$) >= TOTAL QUOTA (120$) >>
		// send $40 ATOM from umee to gaia
		atom40 := mulCoin(atomQuota, "0.4")
		s.SendIBC(s.Chain.ID, setup.GaiaChainID, "", atom40, true, "below token quota but not total quota")
		// supply will be not be decreased because sending more than total quota from umee to gaia
		s.checkSupply(umeeAPIEndpoint, uatomIBCHash, atomFromGaia.Amount)

		// ✅ << BELOW TOKEN QUTOA 5$ but ATOM_QUOTA (5$)+ UMEE_QUOTA(90$) <= TOTAL QUOTA (120$) >>
		// send $15 ATOM from umee to gaia
		sendAtom := mulCoin(atomQuota, "0.05")
		s.SendIBC(s.Chain.ID, setup.GaiaChainID, "", sendAtom, false, "below both quotas")
		// remaing supply decreased uatom on umee
		s.checkSupply(umeeAPIEndpoint, uatomIBCHash, atomFromGaia.Amount.Sub(sendAtom.Amount))
		s.checkOutflows(umeeAPIEndpoint, uatomIBCHash, true, sdk.NewDecFromInt(sendAtom.Amount), atomSymbol)

		// send $45 UMEE from gaia to umee
		returnUmee := mulCoin(sendUmee, "0.5")
		returnUmee.Denom = umeeIBCHash
		coins, err := s.QueryTotalSupply(gaiaAPIEndpoint) // before sending back
		remainingTokens := coins.AmountOf(umeeIBCHash).Sub(returnUmee.Amount)
		s.Require().NoError(err)
		s.SendIBC(setup.GaiaChainID, s.Chain.ID, "", returnUmee, false, "send back some umee")
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, remainingTokens)

		// sending back remaining amount
		s.SendIBC(setup.GaiaChainID, s.Chain.ID, "", sdk.NewCoin(umeeIBCHash, remainingTokens), false, "send back remaining umee")
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, math.ZeroInt())

		// reset the outflows
		s.T().Logf("waiting until quota reset, basically it will take around 300 seconds to do quota reset")
		s.Require().Eventually(
			func() bool {
				amount, err := s.QueryOutflows(umeeAPIEndpoint, appparams.BondDenom)
				if err != nil {
					return false
				}
				if amount.IsZero() {
					s.T().Logf("quota is reset : %s is 0", appparams.BondDenom)
					return true
				}
				return false
			},
			4*time.Minute,
			3*time.Second,
		)

		/****
			IBC_Status : disble (making ibc_transfer quota check disabled)
			No Outflows will updated
		***/
		// Make gov proposal to disable the quota check on ibc-transfer

		for i := 0; i < 10; i++ {
			err = grpc.UIBCIBCTransferStatusUpdate(
				s.AccountClient(0),
				uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_DISABLED,
			)

			if err == nil {
				break
			}

			time.Sleep(time.Duration(i+1) * time.Second)
		}

		s.Require().NoError(err)
		// Get the uibc params for quota checking
		uibcParams, err := s.AccountClient(0).QueryUIBCParams()
		s.Require().NoError(err)
		s.Require().Equal(uibcParams.IbcStatus, uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_DISABLED)

		// sending the umee tokens - they would have exceeded quota before
		s.SendIBC(s.Chain.ID, setup.GaiaChainID, "", exceedUmee, false, "sending umee")
		s.checkSupply(gaiaAPIEndpoint, umeeIBCHash, exceedUmee.Amount)
		// Check the outflows
		s.Require().Eventually(
			func() bool {
				a, err := s.QueryOutflows(umeeAPIEndpoint, appparams.BondDenom)
				if err != nil {
					return false
				}
				return a.Equal(sdk.ZeroDec())
			},
			30*time.Second,
			1*time.Second,
		)
	})
}
