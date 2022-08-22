package simulation

import (
	"bytes"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding leverage type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		prefixA := kvA.Key[:1]

		switch {
		case bytes.Equal(prefixA, types.KeyPrefixRegisteredToken):
			var registeredTokenA, registeredTokenB types.Token
			cdc.MustUnmarshal(kvA.Value, &registeredTokenA)
			cdc.MustUnmarshal(kvB.Value, &registeredTokenB)
			return fmt.Sprintf("%v\n%v", registeredTokenA, registeredTokenB)

		case bytes.Equal(prefixA, types.KeyPrefixAdjustedBorrow):
			var amountA, amountB sdk.Dec
			if err := amountA.Unmarshal(kvA.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			if err := amountB.Unmarshal(kvB.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			return fmt.Sprintf("%v\n%v", amountA, amountB)

		case bytes.Equal(prefixA, types.KeyPrefixCollateralAmount):
			var amountA, amountB sdkmath.Int
			if err := amountA.Unmarshal(kvA.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			if err := amountB.Unmarshal(kvB.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			return fmt.Sprintf("%v\n%v", amountA, amountB)

		case bytes.Equal(prefixA, types.KeyPrefixReserveAmount):
			var amountA, amountB sdkmath.Int
			if err := amountA.Unmarshal(kvA.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			if err := amountB.Unmarshal(kvB.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			return fmt.Sprintf("%v\n%v", amountA, amountB)

		case bytes.Equal(prefixA, types.KeyPrefixLastInterestTime):
			var lastInterestTimeA, lastInterestTimeB gogotypes.Int64Value
			cdc.MustUnmarshal(kvA.Value, &lastInterestTimeA)
			cdc.MustUnmarshal(kvA.Value, &lastInterestTimeB)
			return fmt.Sprintf("%v\n%v", lastInterestTimeA, lastInterestTimeB)

		case bytes.Equal(prefixA, types.KeyPrefixBadDebt):
			return fmt.Sprintf("%v\n%v", kvA, kvB) // it is bytes: []byte{0x01}

		case bytes.Equal(prefixA, types.KeyPrefixInterestScalar):
			var scalarA, scalarB sdk.Dec
			if err := scalarA.Unmarshal(kvA.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			if err := scalarB.Unmarshal(kvB.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			return fmt.Sprintf("%v\n%v", scalarA, scalarB)

		case bytes.Equal(prefixA, types.KeyPrefixAdjustedTotalBorrow):
			var totalA, totalB sdk.Dec
			if err := totalA.Unmarshal(kvA.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			if err := totalB.Unmarshal(kvB.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			return fmt.Sprintf("%v\n%v", totalA, totalB)

		default:
			panic(fmt.Sprintf("invalid leverage key prefix %X", kvA.Key[:1]))
		}
	}
}
