package keeper_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v040 "github.com/cosmos/cosmos-sdk/x/auth/legacy/v040"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/x/peggy/keeper"
	"github.com/umee-network/umee/x/peggy/testpeggy"
	"github.com/umee-network/umee/x/peggy/types"
)

func TestPrefixRange(t *testing.T) {
	cases := map[string]struct {
		src      []byte
		expStart []byte
		expEnd   []byte
		expPanic bool
	}{
		"normal":                 {src: []byte{1, 3, 4}, expStart: []byte{1, 3, 4}, expEnd: []byte{1, 3, 5}},
		"normal short":           {src: []byte{79}, expStart: []byte{79}, expEnd: []byte{80}},
		"empty case":             {src: []byte{}},
		"roll-over example 1":    {src: []byte{17, 28, 255}, expStart: []byte{17, 28, 255}, expEnd: []byte{17, 29, 0}},
		"roll-over example 2":    {src: []byte{15, 42, 255, 255}, expStart: []byte{15, 42, 255, 255}, expEnd: []byte{15, 43, 0, 0}},
		"pathological roll-over": {src: []byte{255, 255, 255, 255}, expStart: []byte{255, 255, 255, 255}},
		"nil prohibited":         {expPanic: true},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if tc.expPanic {
				require.Panics(t, func() {
					keeper.PrefixRange(tc.src)
				})
				return
			}
			start, end := keeper.PrefixRange(tc.src)
			assert.Equal(t, tc.expStart, start)
			assert.Equal(t, tc.expEnd, end)
		})
	}
}

func TestCurrentValsetNormalization(t *testing.T) {
	specs := map[string]struct {
		srcPowers []uint64
		expPowers []uint64
	}{
		"one": {
			srcPowers: []uint64{100},
			expPowers: []uint64{4294967295},
		},
		"two": {
			srcPowers: []uint64{100, 1},
			expPowers: []uint64{4252442866, 42524428},
		},
	}
	input := testpeggy.CreateTestEnv(t)
	ctx := input.Context
	for msg, spec := range specs {
		spec := spec
		t.Run(msg, func(t *testing.T) {
			operators := make([]testpeggy.MockStakingValidatorData, len(spec.srcPowers))
			for i, v := range spec.srcPowers {
				cAddr := bytes.Repeat([]byte{byte(i)}, v040.AddrLen)
				operators[i] = testpeggy.MockStakingValidatorData{
					// any unique addr
					Operator: cAddr,
					Power:    int64(v),
				}
				input.PeggyKeeper.SetEthAddressForValidator(ctx, cAddr, common.HexToAddress("0xf71402f886b45c134743F4c00750823Bbf5Fd045"))
			}
			input.PeggyKeeper.StakingKeeper = testpeggy.NewStakingKeeperWeightedMock(operators...)
			r := input.PeggyKeeper.GetCurrentValset(ctx)
			assert.Equal(t, spec.expPowers, types.BridgeValidators(r.Members).GetPowers())
		})
	}
}

func TestAttestationIterator(t *testing.T) {
	input := testpeggy.CreateTestEnv(t)
	ctx := input.Context
	// add some attestations to the store

	att1 := &types.Attestation{
		Observed: true,
		Votes:    []string{},
	}
	dep1 := &types.MsgDepositClaim{
		EventNonce:     1,
		TokenContract:  testpeggy.TokenContractAddrs[0],
		Amount:         sdk.NewInt(100),
		EthereumSender: testpeggy.EthAddrs[0].String(),
		CosmosReceiver: testpeggy.AccAddrs[0].String(),
		Orchestrator:   testpeggy.AccAddrs[0].String(),
	}
	att2 := &types.Attestation{
		Observed: true,
		Votes:    []string{},
	}
	dep2 := &types.MsgDepositClaim{
		EventNonce:     2,
		TokenContract:  testpeggy.TokenContractAddrs[0],
		Amount:         sdk.NewInt(100),
		EthereumSender: testpeggy.EthAddrs[0].String(),
		CosmosReceiver: testpeggy.AccAddrs[0].String(),
		Orchestrator:   testpeggy.AccAddrs[0].String(),
	}
	input.PeggyKeeper.SetAttestation(ctx, dep1.EventNonce, dep1.ClaimHash(), att1)
	input.PeggyKeeper.SetAttestation(ctx, dep2.EventNonce, dep2.ClaimHash(), att2)

	attestations := []*types.Attestation{}
	input.PeggyKeeper.IterateAttestations(ctx, func(_ []byte, attestation *types.Attestation) (stop bool) {
		attestations = append(attestations, attestation)
		return false
	})

	require.Len(t, attestations, 2)
}

func TestOrchestratorAddresses(t *testing.T) {
	input := testpeggy.CreateTestEnv(t)
	ctx := input.Context
	k := input.PeggyKeeper
	var ethAddrs = []string{"0x3146D2d6Eed46Afa423969f5dDC3152DfC359b09", "0x610277F0208D342C576b991daFdCb36E36515e76", "0x835973768750b3ED2D5c3EF5AdcD5eDb44d12aD4", "0xb2A7F3E84F8FdcA1da46c810AEa110dd96BAE6bF"}
	var valAddrs = []string{"cosmosvaloper1jpz0ahls2chajf78nkqczdwwuqcu97w6z3plt4", "cosmosvaloper15n79nty2fj37ant3p2gj4wju4ls6eu6tjwmdt0", "cosmosvaloper16dnkc6ac6ruuyr6l372fc3p77jgjpet6fka0cq", "cosmosvaloper1vrptwhl3ht2txmzy28j9msqkcvmn8gjz507pgu"}
	var orchAddrs = []string{"cosmos1g0etv93428tvxqftnmj25jn06mz6dtdasj5nz7", "cosmos1rhfs24tlw4na04v35tzmjncy785kkw9j27d5kx", "cosmos10upq3tmt04zf55f6hw67m0uyrda3mp722q70rw", "cosmos1nt2uwjh5peg9vz2wfh2m3jjwqnu9kpjlhgpmen"}

	for i := range ethAddrs {
		// set some addresses
		val, err1 := sdk.ValAddressFromBech32(valAddrs[i])
		orch, err2 := sdk.AccAddressFromBech32(orchAddrs[i])
		require.NoError(t, err1)
		require.NoError(t, err2)
		// set the orchestrator address
		k.SetOrchestratorValidator(ctx, val, orch)
		// set the ethereum address
		k.SetEthAddressForValidator(ctx, val, common.HexToAddress(ethAddrs[i]))
	}

	addresses := k.GetOrchestratorAddresses(ctx)
	for i := range addresses {
		res := addresses[i]
		validatorAddr, _ := sdk.ValAddressFromBech32(valAddrs[i])
		validatorAccountAddr := sdk.AccAddress(validatorAddr.Bytes()).String()
		assert.Equal(t, validatorAccountAddr, res.Sender)
		assert.Equal(t, orchAddrs[i], res.Orchestrator)
		assert.Equal(t, ethAddrs[i], res.EthAddress)
	}

}

func TestLastSlashedValsetNonce(t *testing.T) {
	input := testpeggy.CreateTestEnv(t)
	k := input.PeggyKeeper
	ctx := input.Context

	vs := k.GetCurrentValset(ctx)

	i := 1
	for ; i < 10; i++ {
		vs.Height = uint64(i)
		vs.Nonce = uint64(i)
		k.StoreValsetUnsafe(ctx, vs)
	}

	latestValsetNonce := k.GetLatestValsetNonce(ctx)
	assert.Equal(t, latestValsetNonce, uint64(i-1))

	//  lastSlashedValsetNonce should be zero initially.
	lastSlashedValsetNonce := k.GetLastSlashedValsetNonce(ctx)
	assert.Equal(t, lastSlashedValsetNonce, uint64(0))
	unslashedValsets := k.GetUnslashedValsets(ctx, uint64(12))
	assert.Equal(t, len(unslashedValsets), 9)

	// check if last Slashed Valset nonce is set properly or not
	k.SetLastSlashedValsetNonce(ctx, uint64(3))
	lastSlashedValsetNonce = k.GetLastSlashedValsetNonce(ctx)
	assert.Equal(t, lastSlashedValsetNonce, uint64(3))

	// when maxHeight < lastSlashedValsetNonce, len(unslashedValsets) should be zero
	unslashedValsets = k.GetUnslashedValsets(ctx, uint64(2))
	assert.Equal(t, len(unslashedValsets), 0)

	// when maxHeight == lastSlashedValsetNonce, len(unslashedValsets) should be zero
	unslashedValsets = k.GetUnslashedValsets(ctx, uint64(3))
	assert.Equal(t, len(unslashedValsets), 0)

	// when maxHeight > lastSlashedValsetNonce && maxHeight <= latestValsetNonce
	unslashedValsets = k.GetUnslashedValsets(ctx, uint64(6))
	assert.Equal(t, len(unslashedValsets), 2)

	// when maxHeight > latestValsetNonce
	unslashedValsets = k.GetUnslashedValsets(ctx, uint64(15))
	assert.Equal(t, len(unslashedValsets), 6)
	fmt.Println("unslashedValsetsRange", unslashedValsets)
}
