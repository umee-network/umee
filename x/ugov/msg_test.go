package ugov

import (
	fmt "fmt"
	"testing"

	"github.com/umee-network/umee/v5/util/bpmath"

	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/umee-network/umee/v5/tests/accs"
	"github.com/umee-network/umee/v5/util/coin"
)

var govAddr = authtypes.NewModuleAddress("gov").String()

func validMsgGovUpdateMinGasPrice() MsgGovUpdateMinGasPrice {
	return MsgGovUpdateMinGasPrice{
		Authority:   govAddr,
		MinGasPrice: coin.Atom1_25dec,
	}
}

func TestMsgGovUpdateMinGasPrice(t *testing.T) {
	t.Parallel()

	msg := validMsgGovUpdateMinGasPrice()
	assert.NilError(t, msg.ValidateBasic())

	assert.Equal(t,
		`<authority: umee10d07y265gmmuvt4z0w9aw880jnsr700jg5w6jp, min_gas_price: 1.250000000000000000ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9>`,
		msg.String())

	signers := msg.GetSigners()
	assert.Equal(t, len(signers), 1)
	assert.Equal(t, msg.Authority, signers[0].String())

	msg.MinGasPrice.Amount = sdk.MustNewDecFromStr("0.0000123")
	assert.NilError(t, msg.ValidateBasic(), "fractional amount should be allowed")

	msg.MinGasPrice.Amount = sdk.NewDec(0)
	assert.NilError(t, msg.ValidateBasic(), "zero amount should be allowed")

	// error cases
	msg.MinGasPrice.Amount = sdk.NewDec(-1)
	assert.ErrorContains(t, msg.ValidateBasic(), "amount cannot be negative")

	msg = validMsgGovUpdateMinGasPrice()
	msg.Authority = accs.Alice.String()
	assert.ErrorIs(t, msg.ValidateBasic(), govtypes.ErrInvalidSigner, "must fail on a non gov account")
}

func validMsgGovSetEmergencyGroup() MsgGovSetEmergencyGroup {
	return MsgGovSetEmergencyGroup{
		Authority:      govAddr,
		EmergencyGroup: accs.Alice.String(),
	}
}

func TestMsgGovSetEmergencyGroup(t *testing.T) {
	t.Parallel()

	msg := validMsgGovSetEmergencyGroup()
	assert.Equal(t, fmt.Sprintf("authority:%q emergency_group:%q ", msg.Authority, msg.EmergencyGroup),
		msg.String())
	assert.NilError(t, msg.ValidateBasic())

	signers := msg.GetSigners()
	assert.Equal(t, len(signers), 1)
	assert.Equal(t, msg.Authority, signers[0].String())

	msg.Authority = accs.Bob.String()
	assert.ErrorIs(t, msg.ValidateBasic(), govtypes.ErrInvalidSigner, "must fail on a non gov account")
	assert.ErrorIs(t, msg.ValidateBasic(), govtypes.ErrInvalidSigner, "must fail on a non gov account")
	msg = validMsgGovSetEmergencyGroup()
	msg.EmergencyGroup = "umee1yesmdu06f7strl67kjvg2w7t5kacc"
	assert.ErrorContains(t, msg.ValidateBasic(), "bech32 failed", "must fail with bad emergency_group address")
}

func validMsgGovUpdateInflationParams() MsgGovUpdateInflationParams {
	return MsgGovUpdateInflationParams{
		Authority: govAddr,
		Params:    DefaultInflationParams(),
	}
}

func TestMsgGovUpdateInflationParams(t *testing.T) {
	t.Parallel()

	msg := validMsgGovUpdateInflationParams()
	assert.NilError(t, msg.ValidateBasic())

	signers := msg.GetSigners()
	assert.Equal(t, len(signers), 1)
	assert.Equal(t, msg.Authority, signers[0].String())

	msg.Authority = "umee1yesmdu06f7strl67kjvg2w7t5kacc"
	assert.ErrorIs(t, msg.ValidateBasic(), govtypes.ErrInvalidSigner, "must fail on a non gov account")
	msg.Params.InflationReductionRate = bpmath.FixedBP(10)
	assert.ErrorContains(t, msg.Params.Validate(), "inflation reduction must be between 100bp to 10'000bp")
	msg.Params.MaxSupply = coin.Negative1("test")
	assert.ErrorContains(t, msg.Params.Validate(), "max_supply must be positive")
}
