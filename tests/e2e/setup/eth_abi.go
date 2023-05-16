package e2esetup

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
)

const (
	AbiMethodNameBalanceOf = "balanceOf"
)

var (
	AbiAddressTy, _ = abi.NewType("address", "", nil)
	AbiUint256Ty, _ = abi.NewType("uint256", "", nil)

	EthABI = abi.ABI{
		Methods: map[string]abi.Method{
			AbiMethodNameBalanceOf: abi.NewMethod(
				AbiMethodNameBalanceOf,
				AbiMethodNameBalanceOf,
				abi.Function,
				"view",
				false,
				false,
				abi.Arguments{
					{Name: "account", Type: AbiAddressTy},
				},
				abi.Arguments{
					{Type: AbiUint256Ty},
				},
			),
		},
	}
)
