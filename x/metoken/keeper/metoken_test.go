package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/leverage/types"
	"github.com/umee-network/umee/v5/x/metoken"
	"github.com/umee-network/umee/v5/x/metoken/mocks"
)

func TestIndex_AddAndUpdate(t *testing.T) {
	l := NewLeverageMock()
	k := initMeUSDKeeper(t, nil, l, nil)
	index, err := k.RegisteredIndex(mocks.MeUSDDenom)
	require.NoError(t, err)

	indexWithNotRegisteredAsset := metoken.NewIndex(
		"", sdkmath.ZeroInt(), 6, metoken.Fee{},
		[]metoken.AcceptedAsset{
			metoken.NewAcceptedAsset(
				"test", sdk.MustNewDecFromStr("0.2"),
				sdk.MustNewDecFromStr("1.0"),
			),
		},
	)

	invalid := metoken.NewIndex(
		"", sdkmath.ZeroInt(), 6, metoken.Fee{},
		[]metoken.AcceptedAsset{
			metoken.NewAcceptedAsset(
				mocks.TestDenom1, sdk.MustNewDecFromStr("0.2"),
				sdk.MustNewDecFromStr("1.0"),
			),
		},
	)

	changedExponent := mocks.StableIndex(mocks.MeUSDDenom)
	changedExponent.Exponent = 10

	deletedAsset := mocks.StableIndex(mocks.MeUSDDenom)
	deletedAsset.AcceptedAssets = deletedAsset.AcceptedAssets[:1]

	addDuplicatedAsset := mocks.StableIndex("me/Test")

	tcs := []struct {
		name        string
		addIndex    []metoken.Index
		updateIndex []metoken.Index
		errMsg      string
	}{
		{
			name:        "add: duplicated index",
			addIndex:    []metoken.Index{index},
			updateIndex: nil,
			errMsg:      "already exists",
		},
		{
			name:        "add: asset not registered in x/leverage",
			addIndex:    []metoken.Index{indexWithNotRegisteredAsset},
			updateIndex: nil,
			errMsg:      types.ErrNotRegisteredToken.Error(),
		},
		{
			name:        "add: index don't pass validation",
			addIndex:    []metoken.Index{invalid},
			updateIndex: nil,
			errMsg:      "should have the following format",
		},
		{
			name:        "add: asset accepted by another index",
			addIndex:    []metoken.Index{addDuplicatedAsset},
			updateIndex: nil,
			errMsg:      "is already accepted in another index",
		},
		{
			name:        "add: valid",
			addIndex:    []metoken.Index{mocks.NonStableIndex("me/TH")},
			updateIndex: nil,
			errMsg:      "",
		},
		{
			name:        "update: index not found",
			addIndex:    nil,
			updateIndex: []metoken.Index{mocks.StableIndex("me/NotFound")},
			errMsg:      "not found",
		},
		{
			name:        "update: changed exponent after minting meTokens",
			addIndex:    nil,
			updateIndex: []metoken.Index{changedExponent},
			errMsg:      "exponent cannot be changed when supply is greater than zero",
		},
		{
			name:        "update: deleted asset",
			addIndex:    nil,
			updateIndex: []metoken.Index{deletedAsset},
			errMsg:      "cannot be deleted from an index",
		},
		{
			name:        "update: valid",
			addIndex:    nil,
			updateIndex: []metoken.Index{mocks.StableIndex(mocks.MeUSDDenom)},
			errMsg:      "",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				err := k.UpdateIndexes(tc.addIndex, tc.updateIndex)
				if tc.errMsg != "" {
					assert.ErrorContains(t, err, tc.errMsg)
				} else {
					assert.NilError(t, err)
				}
			},
		)
	}
}
