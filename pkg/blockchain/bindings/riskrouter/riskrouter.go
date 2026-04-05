// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package riskrouter

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

// RiskRouterTradeIntent is an auto generated low-level Go binding around an user-defined struct.
type RiskRouterTradeIntent struct {
	AgentId         *big.Int
	AgentWallet     common.Address
	Pair            string
	Action          string
	AmountUsdScaled *big.Int
	MaxSlippageBps  *big.Int
	Nonce           *big.Int
	Deadline        *big.Int
}

// RiskrouterMetaData contains all meta data concerning the Riskrouter contract.
var RiskrouterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"agentWallet\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"pair\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"action\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amountUsdScaled\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxSlippageBps\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"internalType\":\"structRiskRouter.TradeIntent\",\"name\":\"intent\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"submitTradeIntent\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"agentWallet\",\"type\":\"address\"}],\"name\":\"getIntentNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"agentWallet\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"pair\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"action\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amountUsdScaled\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxSlippageBps\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"internalType\":\"structRiskRouter.TradeIntent\",\"name\":\"intent\",\"type\":\"tuple\"}],\"name\":\"simulateIntent\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// RiskrouterABI is the input ABI used to generate the binding from.
// Deprecated: Use RiskrouterMetaData.ABI instead.
var RiskrouterABI = RiskrouterMetaData.ABI

// Riskrouter is an auto generated Go binding around an Ethereum contract.
type Riskrouter struct {
	RiskrouterCaller     // Read-only binding to the contract
	RiskrouterTransactor // Write-only binding to the contract
	RiskrouterFilterer   // Log filterer for contract events
}

// RiskrouterCaller is an auto generated read-only Go binding around an Ethereum contract.
type RiskrouterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RiskrouterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RiskrouterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RiskrouterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RiskrouterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RiskrouterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RiskrouterSession struct {
	Contract     *Riskrouter       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RiskrouterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RiskrouterCallerSession struct {
	Contract *RiskrouterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// RiskrouterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RiskrouterTransactorSession struct {
	Contract     *RiskrouterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// RiskrouterRaw is an auto generated low-level Go binding around an Ethereum contract.
type RiskrouterRaw struct {
	Contract *Riskrouter // Generic contract binding to access the raw methods on
}

// RiskrouterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RiskrouterCallerRaw struct {
	Contract *RiskrouterCaller // Generic read-only contract binding to access the raw methods on
}

// RiskrouterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RiskrouterTransactorRaw struct {
	Contract *RiskrouterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRiskrouter creates a new instance of Riskrouter, bound to a specific deployed contract.
func NewRiskrouter(address common.Address, backend bind.ContractBackend) (*Riskrouter, error) {
	contract, err := bindRiskrouter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Riskrouter{RiskrouterCaller: RiskrouterCaller{contract: contract}, RiskrouterTransactor: RiskrouterTransactor{contract: contract}, RiskrouterFilterer: RiskrouterFilterer{contract: contract}}, nil
}

// NewRiskrouterCaller creates a new read-only instance of Riskrouter, bound to a specific deployed contract.
func NewRiskrouterCaller(address common.Address, caller bind.ContractCaller) (*RiskrouterCaller, error) {
	contract, err := bindRiskrouter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RiskrouterCaller{contract: contract}, nil
}

// NewRiskrouterTransactor creates a new write-only instance of Riskrouter, bound to a specific deployed contract.
func NewRiskrouterTransactor(address common.Address, transactor bind.ContractTransactor) (*RiskrouterTransactor, error) {
	contract, err := bindRiskrouter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RiskrouterTransactor{contract: contract}, nil
}

// NewRiskrouterFilterer creates a new log filterer instance of Riskrouter, bound to a specific deployed contract.
func NewRiskrouterFilterer(address common.Address, filterer bind.ContractFilterer) (*RiskrouterFilterer, error) {
	contract, err := bindRiskrouter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RiskrouterFilterer{contract: contract}, nil
}

// bindRiskrouter binds a generic wrapper to an already deployed contract.
func bindRiskrouter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := RiskrouterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Riskrouter *RiskrouterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Riskrouter.Contract.RiskrouterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Riskrouter *RiskrouterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Riskrouter.Contract.RiskrouterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Riskrouter *RiskrouterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Riskrouter.Contract.RiskrouterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Riskrouter *RiskrouterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Riskrouter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Riskrouter *RiskrouterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Riskrouter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Riskrouter *RiskrouterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Riskrouter.Contract.contract.Transact(opts, method, params...)
}

// GetIntentNonce is a free data retrieval call binding the contract method 0x14126308.
//
// Solidity: function getIntentNonce(address agentWallet) view returns(uint256)
func (_Riskrouter *RiskrouterCaller) GetIntentNonce(opts *bind.CallOpts, agentWallet common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Riskrouter.contract.Call(opts, &out, "getIntentNonce", agentWallet)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetIntentNonce is a free data retrieval call binding the contract method 0x14126308.
//
// Solidity: function getIntentNonce(address agentWallet) view returns(uint256)
func (_Riskrouter *RiskrouterSession) GetIntentNonce(agentWallet common.Address) (*big.Int, error) {
	return _Riskrouter.Contract.GetIntentNonce(&_Riskrouter.CallOpts, agentWallet)
}

// GetIntentNonce is a free data retrieval call binding the contract method 0x14126308.
//
// Solidity: function getIntentNonce(address agentWallet) view returns(uint256)
func (_Riskrouter *RiskrouterCallerSession) GetIntentNonce(agentWallet common.Address) (*big.Int, error) {
	return _Riskrouter.Contract.GetIntentNonce(&_Riskrouter.CallOpts, agentWallet)
}

// SimulateIntent is a free data retrieval call binding the contract method 0x38f5f123.
//
// Solidity: function simulateIntent((uint256,address,string,string,uint256,uint256,uint256,uint256) intent) view returns(bool approved, string reason)
func (_Riskrouter *RiskrouterCaller) SimulateIntent(opts *bind.CallOpts, intent RiskRouterTradeIntent) (struct {
	Approved bool
	Reason   string
}, error) {
	var out []interface{}
	err := _Riskrouter.contract.Call(opts, &out, "simulateIntent", intent)

	outstruct := new(struct {
		Approved bool
		Reason   string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Approved = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.Reason = *abi.ConvertType(out[1], new(string)).(*string)

	return *outstruct, err

}

// SimulateIntent is a free data retrieval call binding the contract method 0x38f5f123.
//
// Solidity: function simulateIntent((uint256,address,string,string,uint256,uint256,uint256,uint256) intent) view returns(bool approved, string reason)
func (_Riskrouter *RiskrouterSession) SimulateIntent(intent RiskRouterTradeIntent) (struct {
	Approved bool
	Reason   string
}, error) {
	return _Riskrouter.Contract.SimulateIntent(&_Riskrouter.CallOpts, intent)
}

// SimulateIntent is a free data retrieval call binding the contract method 0x38f5f123.
//
// Solidity: function simulateIntent((uint256,address,string,string,uint256,uint256,uint256,uint256) intent) view returns(bool approved, string reason)
func (_Riskrouter *RiskrouterCallerSession) SimulateIntent(intent RiskRouterTradeIntent) (struct {
	Approved bool
	Reason   string
}, error) {
	return _Riskrouter.Contract.SimulateIntent(&_Riskrouter.CallOpts, intent)
}

// SubmitTradeIntent is a paid mutator transaction binding the contract method 0x3577e7bd.
//
// Solidity: function submitTradeIntent((uint256,address,string,string,uint256,uint256,uint256,uint256) intent, bytes signature) returns()
func (_Riskrouter *RiskrouterTransactor) SubmitTradeIntent(opts *bind.TransactOpts, intent RiskRouterTradeIntent, signature []byte) (*types.Transaction, error) {
	return _Riskrouter.contract.Transact(opts, "submitTradeIntent", intent, signature)
}

// SubmitTradeIntent is a paid mutator transaction binding the contract method 0x3577e7bd.
//
// Solidity: function submitTradeIntent((uint256,address,string,string,uint256,uint256,uint256,uint256) intent, bytes signature) returns()
func (_Riskrouter *RiskrouterSession) SubmitTradeIntent(intent RiskRouterTradeIntent, signature []byte) (*types.Transaction, error) {
	return _Riskrouter.Contract.SubmitTradeIntent(&_Riskrouter.TransactOpts, intent, signature)
}

// SubmitTradeIntent is a paid mutator transaction binding the contract method 0x3577e7bd.
//
// Solidity: function submitTradeIntent((uint256,address,string,string,uint256,uint256,uint256,uint256) intent, bytes signature) returns()
func (_Riskrouter *RiskrouterTransactorSession) SubmitTradeIntent(intent RiskRouterTradeIntent, signature []byte) (*types.Transaction, error) {
	return _Riskrouter.Contract.SubmitTradeIntent(&_Riskrouter.TransactOpts, intent, signature)
}
