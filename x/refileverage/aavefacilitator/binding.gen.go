// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package aavefacilitator

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// AavefacilitatorMetaData contains all meta data concerning the Aavefacilitator contract.
var AavefacilitatorMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_ghoToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_aaveGovernance\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_bridge\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"aaveGovernance\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"mintLimit\",\"type\":\"uint128\"},{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"}],\"name\":\"addFaciliator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bridge\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ghoToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"onAxelarGmp\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"newLimit\",\"type\":\"uint128\"}],\"name\":\"updateMintLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// AavefacilitatorABI is the input ABI used to generate the binding from.
// Deprecated: Use AavefacilitatorMetaData.ABI instead.
var AavefacilitatorABI = AavefacilitatorMetaData.ABI

// Aavefacilitator is an auto generated Go binding around an Ethereum contract.
type Aavefacilitator struct {
	AavefacilitatorCaller     // Read-only binding to the contract
	AavefacilitatorTransactor // Write-only binding to the contract
	AavefacilitatorFilterer   // Log filterer for contract events
}

// AavefacilitatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type AavefacilitatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AavefacilitatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AavefacilitatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AavefacilitatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AavefacilitatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AavefacilitatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AavefacilitatorSession struct {
	Contract     *Aavefacilitator  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AavefacilitatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AavefacilitatorCallerSession struct {
	Contract *AavefacilitatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// AavefacilitatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AavefacilitatorTransactorSession struct {
	Contract     *AavefacilitatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// AavefacilitatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type AavefacilitatorRaw struct {
	Contract *Aavefacilitator // Generic contract binding to access the raw methods on
}

// AavefacilitatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AavefacilitatorCallerRaw struct {
	Contract *AavefacilitatorCaller // Generic read-only contract binding to access the raw methods on
}

// AavefacilitatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AavefacilitatorTransactorRaw struct {
	Contract *AavefacilitatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAavefacilitator creates a new instance of Aavefacilitator, bound to a specific deployed contract.
func NewAavefacilitator(address common.Address, backend bind.ContractBackend) (*Aavefacilitator, error) {
	contract, err := bindAavefacilitator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Aavefacilitator{AavefacilitatorCaller: AavefacilitatorCaller{contract: contract}, AavefacilitatorTransactor: AavefacilitatorTransactor{contract: contract}, AavefacilitatorFilterer: AavefacilitatorFilterer{contract: contract}}, nil
}

// NewAavefacilitatorCaller creates a new read-only instance of Aavefacilitator, bound to a specific deployed contract.
func NewAavefacilitatorCaller(address common.Address, caller bind.ContractCaller) (*AavefacilitatorCaller, error) {
	contract, err := bindAavefacilitator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AavefacilitatorCaller{contract: contract}, nil
}

// NewAavefacilitatorTransactor creates a new write-only instance of Aavefacilitator, bound to a specific deployed contract.
func NewAavefacilitatorTransactor(address common.Address, transactor bind.ContractTransactor) (*AavefacilitatorTransactor, error) {
	contract, err := bindAavefacilitator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AavefacilitatorTransactor{contract: contract}, nil
}

// NewAavefacilitatorFilterer creates a new log filterer instance of Aavefacilitator, bound to a specific deployed contract.
func NewAavefacilitatorFilterer(address common.Address, filterer bind.ContractFilterer) (*AavefacilitatorFilterer, error) {
	contract, err := bindAavefacilitator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AavefacilitatorFilterer{contract: contract}, nil
}

// bindAavefacilitator binds a generic wrapper to an already deployed contract.
func bindAavefacilitator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AavefacilitatorABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Aavefacilitator *AavefacilitatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Aavefacilitator.Contract.AavefacilitatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Aavefacilitator *AavefacilitatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Aavefacilitator.Contract.AavefacilitatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Aavefacilitator *AavefacilitatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Aavefacilitator.Contract.AavefacilitatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Aavefacilitator *AavefacilitatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Aavefacilitator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Aavefacilitator *AavefacilitatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Aavefacilitator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Aavefacilitator *AavefacilitatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Aavefacilitator.Contract.contract.Transact(opts, method, params...)
}

// AaveGovernance is a free data retrieval call binding the contract method 0x41b71d49.
//
// Solidity: function aaveGovernance() view returns(address)
func (_Aavefacilitator *AavefacilitatorCaller) AaveGovernance(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Aavefacilitator.contract.Call(opts, &out, "aaveGovernance")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AaveGovernance is a free data retrieval call binding the contract method 0x41b71d49.
//
// Solidity: function aaveGovernance() view returns(address)
func (_Aavefacilitator *AavefacilitatorSession) AaveGovernance() (common.Address, error) {
	return _Aavefacilitator.Contract.AaveGovernance(&_Aavefacilitator.CallOpts)
}

// AaveGovernance is a free data retrieval call binding the contract method 0x41b71d49.
//
// Solidity: function aaveGovernance() view returns(address)
func (_Aavefacilitator *AavefacilitatorCallerSession) AaveGovernance() (common.Address, error) {
	return _Aavefacilitator.Contract.AaveGovernance(&_Aavefacilitator.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_Aavefacilitator *AavefacilitatorCaller) Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Aavefacilitator.contract.Call(opts, &out, "bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_Aavefacilitator *AavefacilitatorSession) Bridge() (common.Address, error) {
	return _Aavefacilitator.Contract.Bridge(&_Aavefacilitator.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_Aavefacilitator *AavefacilitatorCallerSession) Bridge() (common.Address, error) {
	return _Aavefacilitator.Contract.Bridge(&_Aavefacilitator.CallOpts)
}

// GhoToken is a free data retrieval call binding the contract method 0x5996db91.
//
// Solidity: function ghoToken() view returns(address)
func (_Aavefacilitator *AavefacilitatorCaller) GhoToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Aavefacilitator.contract.Call(opts, &out, "ghoToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GhoToken is a free data retrieval call binding the contract method 0x5996db91.
//
// Solidity: function ghoToken() view returns(address)
func (_Aavefacilitator *AavefacilitatorSession) GhoToken() (common.Address, error) {
	return _Aavefacilitator.Contract.GhoToken(&_Aavefacilitator.CallOpts)
}

// GhoToken is a free data retrieval call binding the contract method 0x5996db91.
//
// Solidity: function ghoToken() view returns(address)
func (_Aavefacilitator *AavefacilitatorCallerSession) GhoToken() (common.Address, error) {
	return _Aavefacilitator.Contract.GhoToken(&_Aavefacilitator.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Aavefacilitator *AavefacilitatorCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Aavefacilitator.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Aavefacilitator *AavefacilitatorSession) Owner() (common.Address, error) {
	return _Aavefacilitator.Contract.Owner(&_Aavefacilitator.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Aavefacilitator *AavefacilitatorCallerSession) Owner() (common.Address, error) {
	return _Aavefacilitator.Contract.Owner(&_Aavefacilitator.CallOpts)
}

// AddFaciliator is a paid mutator transaction binding the contract method 0xdefdb73a.
//
// Solidity: function addFaciliator(uint128 mintLimit, string label) returns()
func (_Aavefacilitator *AavefacilitatorTransactor) AddFaciliator(opts *bind.TransactOpts, mintLimit *big.Int, label string) (*types.Transaction, error) {
	return _Aavefacilitator.contract.Transact(opts, "addFaciliator", mintLimit, label)
}

// AddFaciliator is a paid mutator transaction binding the contract method 0xdefdb73a.
//
// Solidity: function addFaciliator(uint128 mintLimit, string label) returns()
func (_Aavefacilitator *AavefacilitatorSession) AddFaciliator(mintLimit *big.Int, label string) (*types.Transaction, error) {
	return _Aavefacilitator.Contract.AddFaciliator(&_Aavefacilitator.TransactOpts, mintLimit, label)
}

// AddFaciliator is a paid mutator transaction binding the contract method 0xdefdb73a.
//
// Solidity: function addFaciliator(uint128 mintLimit, string label) returns()
func (_Aavefacilitator *AavefacilitatorTransactorSession) AddFaciliator(mintLimit *big.Int, label string) (*types.Transaction, error) {
	return _Aavefacilitator.Contract.AddFaciliator(&_Aavefacilitator.TransactOpts, mintLimit, label)
}

// OnAxelarGmp is a paid mutator transaction binding the contract method 0xca0b0a6c.
//
// Solidity: function onAxelarGmp(address recipient, uint256 amount) returns()
func (_Aavefacilitator *AavefacilitatorTransactor) OnAxelarGmp(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Aavefacilitator.contract.Transact(opts, "onAxelarGmp", recipient, amount)
}

// OnAxelarGmp is a paid mutator transaction binding the contract method 0xca0b0a6c.
//
// Solidity: function onAxelarGmp(address recipient, uint256 amount) returns()
func (_Aavefacilitator *AavefacilitatorSession) OnAxelarGmp(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Aavefacilitator.Contract.OnAxelarGmp(&_Aavefacilitator.TransactOpts, recipient, amount)
}

// OnAxelarGmp is a paid mutator transaction binding the contract method 0xca0b0a6c.
//
// Solidity: function onAxelarGmp(address recipient, uint256 amount) returns()
func (_Aavefacilitator *AavefacilitatorTransactorSession) OnAxelarGmp(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Aavefacilitator.Contract.OnAxelarGmp(&_Aavefacilitator.TransactOpts, recipient, amount)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Aavefacilitator *AavefacilitatorTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Aavefacilitator.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Aavefacilitator *AavefacilitatorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Aavefacilitator.Contract.RenounceOwnership(&_Aavefacilitator.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Aavefacilitator *AavefacilitatorTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Aavefacilitator.Contract.RenounceOwnership(&_Aavefacilitator.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Aavefacilitator *AavefacilitatorTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Aavefacilitator.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Aavefacilitator *AavefacilitatorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Aavefacilitator.Contract.TransferOwnership(&_Aavefacilitator.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Aavefacilitator *AavefacilitatorTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Aavefacilitator.Contract.TransferOwnership(&_Aavefacilitator.TransactOpts, newOwner)
}

// UpdateMintLimit is a paid mutator transaction binding the contract method 0x6af74a1a.
//
// Solidity: function updateMintLimit(uint128 newLimit) returns()
func (_Aavefacilitator *AavefacilitatorTransactor) UpdateMintLimit(opts *bind.TransactOpts, newLimit *big.Int) (*types.Transaction, error) {
	return _Aavefacilitator.contract.Transact(opts, "updateMintLimit", newLimit)
}

// UpdateMintLimit is a paid mutator transaction binding the contract method 0x6af74a1a.
//
// Solidity: function updateMintLimit(uint128 newLimit) returns()
func (_Aavefacilitator *AavefacilitatorSession) UpdateMintLimit(newLimit *big.Int) (*types.Transaction, error) {
	return _Aavefacilitator.Contract.UpdateMintLimit(&_Aavefacilitator.TransactOpts, newLimit)
}

// UpdateMintLimit is a paid mutator transaction binding the contract method 0x6af74a1a.
//
// Solidity: function updateMintLimit(uint128 newLimit) returns()
func (_Aavefacilitator *AavefacilitatorTransactorSession) UpdateMintLimit(newLimit *big.Int) (*types.Transaction, error) {
	return _Aavefacilitator.Contract.UpdateMintLimit(&_Aavefacilitator.TransactOpts, newLimit)
}

// AavefacilitatorOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Aavefacilitator contract.
type AavefacilitatorOwnershipTransferredIterator struct {
	Event *AavefacilitatorOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AavefacilitatorOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AavefacilitatorOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AavefacilitatorOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AavefacilitatorOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AavefacilitatorOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AavefacilitatorOwnershipTransferred represents a OwnershipTransferred event raised by the Aavefacilitator contract.
type AavefacilitatorOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Aavefacilitator *AavefacilitatorFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AavefacilitatorOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Aavefacilitator.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AavefacilitatorOwnershipTransferredIterator{contract: _Aavefacilitator.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Aavefacilitator *AavefacilitatorFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AavefacilitatorOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Aavefacilitator.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AavefacilitatorOwnershipTransferred)
				if err := _Aavefacilitator.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Aavefacilitator *AavefacilitatorFilterer) ParseOwnershipTransferred(log types.Log) (*AavefacilitatorOwnershipTransferred, error) {
	event := new(AavefacilitatorOwnershipTransferred)
	if err := _Aavefacilitator.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
