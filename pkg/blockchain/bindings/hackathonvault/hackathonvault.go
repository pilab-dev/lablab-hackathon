// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package hackathonvault

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

// HackathonvaultMetaData contains all meta data concerning the Hackathonvault contract.
var HackathonvaultMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"}],\"name\":\"claimAllocation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"}],\"name\":\"getBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalVaultBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unallocatedBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"}],\"name\":\"hasClaimed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"allocationPerTeam\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// HackathonvaultABI is the input ABI used to generate the binding from.
// Deprecated: Use HackathonvaultMetaData.ABI instead.
var HackathonvaultABI = HackathonvaultMetaData.ABI

// Hackathonvault is an auto generated Go binding around an Ethereum contract.
type Hackathonvault struct {
	HackathonvaultCaller     // Read-only binding to the contract
	HackathonvaultTransactor // Write-only binding to the contract
	HackathonvaultFilterer   // Log filterer for contract events
}

// HackathonvaultCaller is an auto generated read-only Go binding around an Ethereum contract.
type HackathonvaultCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HackathonvaultTransactor is an auto generated write-only Go binding around an Ethereum contract.
type HackathonvaultTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HackathonvaultFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type HackathonvaultFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HackathonvaultSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HackathonvaultSession struct {
	Contract     *Hackathonvault   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HackathonvaultCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HackathonvaultCallerSession struct {
	Contract *HackathonvaultCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// HackathonvaultTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HackathonvaultTransactorSession struct {
	Contract     *HackathonvaultTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// HackathonvaultRaw is an auto generated low-level Go binding around an Ethereum contract.
type HackathonvaultRaw struct {
	Contract *Hackathonvault // Generic contract binding to access the raw methods on
}

// HackathonvaultCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HackathonvaultCallerRaw struct {
	Contract *HackathonvaultCaller // Generic read-only contract binding to access the raw methods on
}

// HackathonvaultTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HackathonvaultTransactorRaw struct {
	Contract *HackathonvaultTransactor // Generic write-only contract binding to access the raw methods on
}

// NewHackathonvault creates a new instance of Hackathonvault, bound to a specific deployed contract.
func NewHackathonvault(address common.Address, backend bind.ContractBackend) (*Hackathonvault, error) {
	contract, err := bindHackathonvault(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Hackathonvault{HackathonvaultCaller: HackathonvaultCaller{contract: contract}, HackathonvaultTransactor: HackathonvaultTransactor{contract: contract}, HackathonvaultFilterer: HackathonvaultFilterer{contract: contract}}, nil
}

// NewHackathonvaultCaller creates a new read-only instance of Hackathonvault, bound to a specific deployed contract.
func NewHackathonvaultCaller(address common.Address, caller bind.ContractCaller) (*HackathonvaultCaller, error) {
	contract, err := bindHackathonvault(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &HackathonvaultCaller{contract: contract}, nil
}

// NewHackathonvaultTransactor creates a new write-only instance of Hackathonvault, bound to a specific deployed contract.
func NewHackathonvaultTransactor(address common.Address, transactor bind.ContractTransactor) (*HackathonvaultTransactor, error) {
	contract, err := bindHackathonvault(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &HackathonvaultTransactor{contract: contract}, nil
}

// NewHackathonvaultFilterer creates a new log filterer instance of Hackathonvault, bound to a specific deployed contract.
func NewHackathonvaultFilterer(address common.Address, filterer bind.ContractFilterer) (*HackathonvaultFilterer, error) {
	contract, err := bindHackathonvault(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &HackathonvaultFilterer{contract: contract}, nil
}

// bindHackathonvault binds a generic wrapper to an already deployed contract.
func bindHackathonvault(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := HackathonvaultMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Hackathonvault *HackathonvaultRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Hackathonvault.Contract.HackathonvaultCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Hackathonvault *HackathonvaultRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hackathonvault.Contract.HackathonvaultTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Hackathonvault *HackathonvaultRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Hackathonvault.Contract.HackathonvaultTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Hackathonvault *HackathonvaultCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Hackathonvault.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Hackathonvault *HackathonvaultTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hackathonvault.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Hackathonvault *HackathonvaultTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Hackathonvault.Contract.contract.Transact(opts, method, params...)
}

// AllocationPerTeam is a free data retrieval call binding the contract method 0x6a0bea2c.
//
// Solidity: function allocationPerTeam() view returns(uint256)
func (_Hackathonvault *HackathonvaultCaller) AllocationPerTeam(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Hackathonvault.contract.Call(opts, &out, "allocationPerTeam")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AllocationPerTeam is a free data retrieval call binding the contract method 0x6a0bea2c.
//
// Solidity: function allocationPerTeam() view returns(uint256)
func (_Hackathonvault *HackathonvaultSession) AllocationPerTeam() (*big.Int, error) {
	return _Hackathonvault.Contract.AllocationPerTeam(&_Hackathonvault.CallOpts)
}

// AllocationPerTeam is a free data retrieval call binding the contract method 0x6a0bea2c.
//
// Solidity: function allocationPerTeam() view returns(uint256)
func (_Hackathonvault *HackathonvaultCallerSession) AllocationPerTeam() (*big.Int, error) {
	return _Hackathonvault.Contract.AllocationPerTeam(&_Hackathonvault.CallOpts)
}

// GetBalance is a free data retrieval call binding the contract method 0x1e010439.
//
// Solidity: function getBalance(uint256 agentId) view returns(uint256)
func (_Hackathonvault *HackathonvaultCaller) GetBalance(opts *bind.CallOpts, agentId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Hackathonvault.contract.Call(opts, &out, "getBalance", agentId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetBalance is a free data retrieval call binding the contract method 0x1e010439.
//
// Solidity: function getBalance(uint256 agentId) view returns(uint256)
func (_Hackathonvault *HackathonvaultSession) GetBalance(agentId *big.Int) (*big.Int, error) {
	return _Hackathonvault.Contract.GetBalance(&_Hackathonvault.CallOpts, agentId)
}

// GetBalance is a free data retrieval call binding the contract method 0x1e010439.
//
// Solidity: function getBalance(uint256 agentId) view returns(uint256)
func (_Hackathonvault *HackathonvaultCallerSession) GetBalance(agentId *big.Int) (*big.Int, error) {
	return _Hackathonvault.Contract.GetBalance(&_Hackathonvault.CallOpts, agentId)
}

// HasClaimed is a free data retrieval call binding the contract method 0xce516507.
//
// Solidity: function hasClaimed(uint256 agentId) view returns(bool)
func (_Hackathonvault *HackathonvaultCaller) HasClaimed(opts *bind.CallOpts, agentId *big.Int) (bool, error) {
	var out []interface{}
	err := _Hackathonvault.contract.Call(opts, &out, "hasClaimed", agentId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasClaimed is a free data retrieval call binding the contract method 0xce516507.
//
// Solidity: function hasClaimed(uint256 agentId) view returns(bool)
func (_Hackathonvault *HackathonvaultSession) HasClaimed(agentId *big.Int) (bool, error) {
	return _Hackathonvault.Contract.HasClaimed(&_Hackathonvault.CallOpts, agentId)
}

// HasClaimed is a free data retrieval call binding the contract method 0xce516507.
//
// Solidity: function hasClaimed(uint256 agentId) view returns(bool)
func (_Hackathonvault *HackathonvaultCallerSession) HasClaimed(agentId *big.Int) (bool, error) {
	return _Hackathonvault.Contract.HasClaimed(&_Hackathonvault.CallOpts, agentId)
}

// TotalVaultBalance is a free data retrieval call binding the contract method 0xcc5d195e.
//
// Solidity: function totalVaultBalance() view returns(uint256)
func (_Hackathonvault *HackathonvaultCaller) TotalVaultBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Hackathonvault.contract.Call(opts, &out, "totalVaultBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalVaultBalance is a free data retrieval call binding the contract method 0xcc5d195e.
//
// Solidity: function totalVaultBalance() view returns(uint256)
func (_Hackathonvault *HackathonvaultSession) TotalVaultBalance() (*big.Int, error) {
	return _Hackathonvault.Contract.TotalVaultBalance(&_Hackathonvault.CallOpts)
}

// TotalVaultBalance is a free data retrieval call binding the contract method 0xcc5d195e.
//
// Solidity: function totalVaultBalance() view returns(uint256)
func (_Hackathonvault *HackathonvaultCallerSession) TotalVaultBalance() (*big.Int, error) {
	return _Hackathonvault.Contract.TotalVaultBalance(&_Hackathonvault.CallOpts)
}

// UnallocatedBalance is a free data retrieval call binding the contract method 0x9583612e.
//
// Solidity: function unallocatedBalance() view returns(uint256)
func (_Hackathonvault *HackathonvaultCaller) UnallocatedBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Hackathonvault.contract.Call(opts, &out, "unallocatedBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UnallocatedBalance is a free data retrieval call binding the contract method 0x9583612e.
//
// Solidity: function unallocatedBalance() view returns(uint256)
func (_Hackathonvault *HackathonvaultSession) UnallocatedBalance() (*big.Int, error) {
	return _Hackathonvault.Contract.UnallocatedBalance(&_Hackathonvault.CallOpts)
}

// UnallocatedBalance is a free data retrieval call binding the contract method 0x9583612e.
//
// Solidity: function unallocatedBalance() view returns(uint256)
func (_Hackathonvault *HackathonvaultCallerSession) UnallocatedBalance() (*big.Int, error) {
	return _Hackathonvault.Contract.UnallocatedBalance(&_Hackathonvault.CallOpts)
}

// ClaimAllocation is a paid mutator transaction binding the contract method 0x45d0c755.
//
// Solidity: function claimAllocation(uint256 agentId) returns()
func (_Hackathonvault *HackathonvaultTransactor) ClaimAllocation(opts *bind.TransactOpts, agentId *big.Int) (*types.Transaction, error) {
	return _Hackathonvault.contract.Transact(opts, "claimAllocation", agentId)
}

// ClaimAllocation is a paid mutator transaction binding the contract method 0x45d0c755.
//
// Solidity: function claimAllocation(uint256 agentId) returns()
func (_Hackathonvault *HackathonvaultSession) ClaimAllocation(agentId *big.Int) (*types.Transaction, error) {
	return _Hackathonvault.Contract.ClaimAllocation(&_Hackathonvault.TransactOpts, agentId)
}

// ClaimAllocation is a paid mutator transaction binding the contract method 0x45d0c755.
//
// Solidity: function claimAllocation(uint256 agentId) returns()
func (_Hackathonvault *HackathonvaultTransactorSession) ClaimAllocation(agentId *big.Int) (*types.Transaction, error) {
	return _Hackathonvault.Contract.ClaimAllocation(&_Hackathonvault.TransactOpts, agentId)
}
