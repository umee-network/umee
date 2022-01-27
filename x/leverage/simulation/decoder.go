package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/umee-network/umee/x/leverage/types"
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

		case bytes.Equal(prefixA, types.KeyPrefixLoanToken):
			var loanTokenA, loanTokenB sdk.Coin
			cdc.MustUnmarshal(kvA.Value, &loanTokenA)
			cdc.MustUnmarshal(kvB.Value, &loanTokenB)
			return fmt.Sprintf("%v\n%v", loanTokenA, loanTokenB)

		case bytes.Equal(prefixA, types.KeyPrefixCollateralSetting):
			return fmt.Sprintf("%v\n%v", kvA, kvB) // it is bytes: []byte{0x01}

		case bytes.Equal(prefixA, types.KeyPrefixCollateralAmount):
			var collateralAmountA, collateralAmountB sdk.Coin
			cdc.MustUnmarshal(kvA.Value, &collateralAmountA)
			cdc.MustUnmarshal(kvB.Value, &collateralAmountB)
			return fmt.Sprintf("%v\n%v", collateralAmountA, collateralAmountB)

		case bytes.Equal(prefixA, types.KeyPrefixReserveAmount):
			var collateralAmountA, collateralAmountB sdk.Coin
			cdc.MustUnmarshal(kvA.Value, &collateralAmountA)
			cdc.MustUnmarshal(kvB.Value, &collateralAmountB)
			return fmt.Sprintf("%v\n%v", collateralAmountA, collateralAmountB)

		case bytes.Equal(prefixA, types.KeyPrefixLastInterestTime):
			var lastInterestTimeA, lastInterestTimeB gogotypes.Int64Value
			cdc.MustUnmarshal(kvA.Value, &lastInterestTimeA)
			cdc.MustUnmarshal(kvA.Value, &lastInterestTimeB)
			return fmt.Sprintf("%v\n%v", lastInterestTimeA, lastInterestTimeB)

		case bytes.Equal(prefixA, types.KeyPrefixExchangeRate):
			exchangeRateA, exchangeRateB := sdk.ZeroDec(), sdk.ZeroDec()
			if err := exchangeRateA.Unmarshal(kvA.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			if err := exchangeRateB.Unmarshal(kvB.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			return fmt.Sprintf("%v\n%v", exchangeRateA, exchangeRateB)

		case bytes.Equal(prefixA, types.KeyPrefixBadDebt):
			return fmt.Sprintf("%v\n%v", kvA, kvB) // it is bytes: []byte{0x01}

		case bytes.Equal(prefixA, types.KeyPrefixBorrowAPY):
			var borrowAPYA, borrowAPYB sdk.Dec
			if err := borrowAPYA.Unmarshal(kvA.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			if err := borrowAPYB.Unmarshal(kvB.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			return fmt.Sprintf("%v\n%v", borrowAPYA, borrowAPYB)

		case bytes.Equal(prefixA, types.KeyPrefixLendAPY):
			var lendAPYA, lendAPYB sdk.Dec
			if err := lendAPYA.Unmarshal(kvA.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			if err := lendAPYB.Unmarshal(kvB.Value); err != nil {
				panic(fmt.Sprintf("invalid unmarshal value %+v", err))
			}
			return fmt.Sprintf("%v\n%v", lendAPYA, lendAPYB)

		default:
			panic(fmt.Sprintf("invalid leverage key prefix %X", kvA.Key[:1]))
		}
	}
}
