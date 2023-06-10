// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package main

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
	_ = abi.ConvertType
)

// ReFiFacilitatorMetaData contains all meta data concerning the ReFiFacilitator contract.
var ReFiFacilitatorMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_ghoToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_aaveGovernance\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_bridge\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"aaveGovernance\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"mintLimit\",\"type\":\"uint128\"},{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"}],\"name\":\"addFaciliator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bridge\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ghoToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"onAxelarGmp\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"newLimit\",\"type\":\"uint128\"}],\"name\":\"updateMintLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ReFiFacilitatorABI is the input ABI used to generate the binding from.
// Deprecated: Use ReFiFacilitatorMetaData.ABI instead.
var ReFiFacilitatorABI = ReFiFacilitatorMetaData.ABI

// ReFiFacilitator is an auto generated Go binding around an Ethereum contract.
type ReFiFacilitator struct {
	ReFiFacilitatorCaller     // Read-only binding to the contract
	ReFiFacilitatorTransactor // Write-only binding to the contract
	ReFiFacilitatorFilterer   // Log filterer for contract events
}

// ReFiFacilitatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type ReFiFacilitatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReFiFacilitatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ReFiFacilitatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReFiFacilitatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ReFiFacilitatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReFiFacilitatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ReFiFacilitatorSession struct {
	Contract     *ReFiFacilitator  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ReFiFacilitatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ReFiFacilitatorCallerSession struct {
	Contract *ReFiFacilitatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// ReFiFacilitatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ReFiFacilitatorTransactorSession struct {
	Contract     *ReFiFacilitatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// ReFiFacilitatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type ReFiFacilitatorRaw struct {
	Contract *ReFiFacilitator // Generic contract binding to access the raw methods on
}

// ReFiFacilitatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ReFiFacilitatorCallerRaw struct {
	Contract *ReFiFacilitatorCaller // Generic read-only contract binding to access the raw methods on
}

// ReFiFacilitatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ReFiFacilitatorTransactorRaw struct {
	Contract *ReFiFacilitatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewReFiFacilitator creates a new instance of ReFiFacilitator, bound to a specific deployed contract.
func NewReFiFacilitator(address common.Address, backend bind.ContractBackend) (*ReFiFacilitator, error) {
	contract, err := bindReFiFacilitator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ReFiFacilitator{ReFiFacilitatorCaller: ReFiFacilitatorCaller{contract: contract}, ReFiFacilitatorTransactor: ReFiFacilitatorTransactor{contract: contract}, ReFiFacilitatorFilterer: ReFiFacilitatorFilterer{contract: contract}}, nil
}

// NewReFiFacilitatorCaller creates a new read-only instance of ReFiFacilitator, bound to a specific deployed contract.
func NewReFiFacilitatorCaller(address common.Address, caller bind.ContractCaller) (*ReFiFacilitatorCaller, error) {
	contract, err := bindReFiFacilitator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ReFiFacilitatorCaller{contract: contract}, nil
}

// NewReFiFacilitatorTransactor creates a new write-only instance of ReFiFacilitator, bound to a specific deployed contract.
func NewReFiFacilitatorTransactor(address common.Address, transactor bind.ContractTransactor) (*ReFiFacilitatorTransactor, error) {
	contract, err := bindReFiFacilitator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ReFiFacilitatorTransactor{contract: contract}, nil
}

// NewReFiFacilitatorFilterer creates a new log filterer instance of ReFiFacilitator, bound to a specific deployed contract.
func NewReFiFacilitatorFilterer(address common.Address, filterer bind.ContractFilterer) (*ReFiFacilitatorFilterer, error) {
	contract, err := bindReFiFacilitator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ReFiFacilitatorFilterer{contract: contract}, nil
}

// bindReFiFacilitator binds a generic wrapper to an already deployed contract.
func bindReFiFacilitator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ReFiFacilitatorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ReFiFacilitator *ReFiFacilitatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ReFiFacilitator.Contract.ReFiFacilitatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ReFiFacilitator *ReFiFacilitatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.ReFiFacilitatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ReFiFacilitator *ReFiFacilitatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.ReFiFacilitatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ReFiFacilitator *ReFiFacilitatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ReFiFacilitator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ReFiFacilitator *ReFiFacilitatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ReFiFacilitator *ReFiFacilitatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.contract.Transact(opts, method, params...)
}

// AaveGovernance is a free data retrieval call binding the contract method 0x41b71d49.
//
// Solidity: function aaveGovernance() view returns(address)
func (_ReFiFacilitator *ReFiFacilitatorCaller) AaveGovernance(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ReFiFacilitator.contract.Call(opts, &out, "aaveGovernance")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AaveGovernance is a free data retrieval call binding the contract method 0x41b71d49.
//
// Solidity: function aaveGovernance() view returns(address)
func (_ReFiFacilitator *ReFiFacilitatorSession) AaveGovernance() (common.Address, error) {
	return _ReFiFacilitator.Contract.AaveGovernance(&_ReFiFacilitator.CallOpts)
}

// AaveGovernance is a free data retrieval call binding the contract method 0x41b71d49.
//
// Solidity: function aaveGovernance() view returns(address)
func (_ReFiFacilitator *ReFiFacilitatorCallerSession) AaveGovernance() (common.Address, error) {
	return _ReFiFacilitator.Contract.AaveGovernance(&_ReFiFacilitator.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_ReFiFacilitator *ReFiFacilitatorCaller) Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ReFiFacilitator.contract.Call(opts, &out, "bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_ReFiFacilitator *ReFiFacilitatorSession) Bridge() (common.Address, error) {
	return _ReFiFacilitator.Contract.Bridge(&_ReFiFacilitator.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_ReFiFacilitator *ReFiFacilitatorCallerSession) Bridge() (common.Address, error) {
	return _ReFiFacilitator.Contract.Bridge(&_ReFiFacilitator.CallOpts)
}

// GhoToken is a free data retrieval call binding the contract method 0x5996db91.
//
// Solidity: function ghoToken() view returns(address)
func (_ReFiFacilitator *ReFiFacilitatorCaller) GhoToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ReFiFacilitator.contract.Call(opts, &out, "ghoToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GhoToken is a free data retrieval call binding the contract method 0x5996db91.
//
// Solidity: function ghoToken() view returns(address)
func (_ReFiFacilitator *ReFiFacilitatorSession) GhoToken() (common.Address, error) {
	return _ReFiFacilitator.Contract.GhoToken(&_ReFiFacilitator.CallOpts)
}

// GhoToken is a free data retrieval call binding the contract method 0x5996db91.
//
// Solidity: function ghoToken() view returns(address)
func (_ReFiFacilitator *ReFiFacilitatorCallerSession) GhoToken() (common.Address, error) {
	return _ReFiFacilitator.Contract.GhoToken(&_ReFiFacilitator.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ReFiFacilitator *ReFiFacilitatorCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ReFiFacilitator.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ReFiFacilitator *ReFiFacilitatorSession) Owner() (common.Address, error) {
	return _ReFiFacilitator.Contract.Owner(&_ReFiFacilitator.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ReFiFacilitator *ReFiFacilitatorCallerSession) Owner() (common.Address, error) {
	return _ReFiFacilitator.Contract.Owner(&_ReFiFacilitator.CallOpts)
}

// AddFaciliator is a paid mutator transaction binding the contract method 0xdefdb73a.
//
// Solidity: function addFaciliator(uint128 mintLimit, string label) returns()
func (_ReFiFacilitator *ReFiFacilitatorTransactor) AddFaciliator(opts *bind.TransactOpts, mintLimit *big.Int, label string) (*types.Transaction, error) {
	return _ReFiFacilitator.contract.Transact(opts, "addFaciliator", mintLimit, label)
}

// AddFaciliator is a paid mutator transaction binding the contract method 0xdefdb73a.
//
// Solidity: function addFaciliator(uint128 mintLimit, string label) returns()
func (_ReFiFacilitator *ReFiFacilitatorSession) AddFaciliator(mintLimit *big.Int, label string) (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.AddFaciliator(&_ReFiFacilitator.TransactOpts, mintLimit, label)
}

// AddFaciliator is a paid mutator transaction binding the contract method 0xdefdb73a.
//
// Solidity: function addFaciliator(uint128 mintLimit, string label) returns()
func (_ReFiFacilitator *ReFiFacilitatorTransactorSession) AddFaciliator(mintLimit *big.Int, label string) (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.AddFaciliator(&_ReFiFacilitator.TransactOpts, mintLimit, label)
}

// OnAxelarGmp is a paid mutator transaction binding the contract method 0xca0b0a6c.
//
// Solidity: function onAxelarGmp(address recipient, uint256 amount) returns()
func (_ReFiFacilitator *ReFiFacilitatorTransactor) OnAxelarGmp(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ReFiFacilitator.contract.Transact(opts, "onAxelarGmp", recipient, amount)
}

// OnAxelarGmp is a paid mutator transaction binding the contract method 0xca0b0a6c.
//
// Solidity: function onAxelarGmp(address recipient, uint256 amount) returns()
func (_ReFiFacilitator *ReFiFacilitatorSession) OnAxelarGmp(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.OnAxelarGmp(&_ReFiFacilitator.TransactOpts, recipient, amount)
}

// OnAxelarGmp is a paid mutator transaction binding the contract method 0xca0b0a6c.
//
// Solidity: function onAxelarGmp(address recipient, uint256 amount) returns()
func (_ReFiFacilitator *ReFiFacilitatorTransactorSession) OnAxelarGmp(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.OnAxelarGmp(&_ReFiFacilitator.TransactOpts, recipient, amount)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ReFiFacilitator *ReFiFacilitatorTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReFiFacilitator.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ReFiFacilitator *ReFiFacilitatorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.RenounceOwnership(&_ReFiFacilitator.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ReFiFacilitator *ReFiFacilitatorTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.RenounceOwnership(&_ReFiFacilitator.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ReFiFacilitator *ReFiFacilitatorTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ReFiFacilitator.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ReFiFacilitator *ReFiFacilitatorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.TransferOwnership(&_ReFiFacilitator.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ReFiFacilitator *ReFiFacilitatorTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.TransferOwnership(&_ReFiFacilitator.TransactOpts, newOwner)
}

// UpdateMintLimit is a paid mutator transaction binding the contract method 0x6af74a1a.
//
// Solidity: function updateMintLimit(uint128 newLimit) returns()
func (_ReFiFacilitator *ReFiFacilitatorTransactor) UpdateMintLimit(opts *bind.TransactOpts, newLimit *big.Int) (*types.Transaction, error) {
	return _ReFiFacilitator.contract.Transact(opts, "updateMintLimit", newLimit)
}

// UpdateMintLimit is a paid mutator transaction binding the contract method 0x6af74a1a.
//
// Solidity: function updateMintLimit(uint128 newLimit) returns()
func (_ReFiFacilitator *ReFiFacilitatorSession) UpdateMintLimit(newLimit *big.Int) (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.UpdateMintLimit(&_ReFiFacilitator.TransactOpts, newLimit)
}

// UpdateMintLimit is a paid mutator transaction binding the contract method 0x6af74a1a.
//
// Solidity: function updateMintLimit(uint128 newLimit) returns()
func (_ReFiFacilitator *ReFiFacilitatorTransactorSession) UpdateMintLimit(newLimit *big.Int) (*types.Transaction, error) {
	return _ReFiFacilitator.Contract.UpdateMintLimit(&_ReFiFacilitator.TransactOpts, newLimit)
}

// ReFiFacilitatorOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ReFiFacilitator contract.
type ReFiFacilitatorOwnershipTransferredIterator struct {
	Event *ReFiFacilitatorOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ReFiFacilitatorOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReFiFacilitatorOwnershipTransferred)
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
		it.Event = new(ReFiFacilitatorOwnershipTransferred)
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
func (it *ReFiFacilitatorOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReFiFacilitatorOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReFiFacilitatorOwnershipTransferred represents a OwnershipTransferred event raised by the ReFiFacilitator contract.
type ReFiFacilitatorOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ReFiFacilitator *ReFiFacilitatorFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ReFiFacilitatorOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ReFiFacilitator.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ReFiFacilitatorOwnershipTransferredIterator{contract: _ReFiFacilitator.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ReFiFacilitator *ReFiFacilitatorFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ReFiFacilitatorOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ReFiFacilitator.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReFiFacilitatorOwnershipTransferred)
				if err := _ReFiFacilitator.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_ReFiFacilitator *ReFiFacilitatorFilterer) ParseOwnershipTransferred(log types.Log) (*ReFiFacilitatorOwnershipTransferred, error) {
	event := new(ReFiFacilitatorOwnershipTransferred)
	if err := _ReFiFacilitator.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
