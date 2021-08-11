package e2e

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
)

const (
	abiMethodNameBalanceOf = "balanceOf"
)

var (
	abiAddressTy, _ = abi.NewType("address", "", nil)
	abiUint256Ty, _ = abi.NewType("uint256", "", nil)

	ethABI = abi.ABI{
		Methods: map[string]abi.Method{
			abiMethodNameBalanceOf: abi.NewMethod(
				abiMethodNameBalanceOf,
				abiMethodNameBalanceOf,
				abi.Function,
				"view",
				false,
				false,
				abi.Arguments{
					{Name: "account", Type: abiAddressTy},
				},
				abi.Arguments{
					{Type: abiUint256Ty},
				},
			),
		},
	}
)
