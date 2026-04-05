// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package reputationregistry

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

// Struct0 is an auto generated low-level Go binding around an user-defined struct.
type Struct0 struct {
	Rater        common.Address
	Score        uint8
	OutcomeRef   [32]byte
	Comment      string
	FeedbackType uint8
	Timestamp    *big.Int
}

// ReputationregistryMetaData contains all meta data concerning the Reputationregistry contract.
var ReputationregistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"score\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"outcomeRef\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"comment\",\"type\":\"string\"},{\"internalType\":\"uint8\",\"name\":\"feedbackType\",\"type\":\"uint8\"}],\"name\":\"submitFeedback\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"}],\"name\":\"getAverageScore\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"}],\"name\":\"getFeedbackHistory\",\"outputs\":[{\"internalType\":\"tuple\",\"name\":\"\",\"components\":[{\"internalType\":\"address\",\"name\":\"rater\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"score\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"outcomeRef\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"comment\",\"type\":\"string\"},{\"internalType\":\"uint8\",\"name\":\"feedbackType\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"rater\",\"type\":\"address\"}],\"name\":\"hasRated\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ReputationregistryABI is the input ABI used to generate the binding from.
// Deprecated: Use ReputationregistryMetaData.ABI instead.
var ReputationregistryABI = ReputationregistryMetaData.ABI

// Reputationregistry is an auto generated Go binding around an Ethereum contract.
type Reputationregistry struct {
	ReputationregistryCaller     // Read-only binding to the contract
	ReputationregistryTransactor // Write-only binding to the contract
	ReputationregistryFilterer   // Log filterer for contract events
}

// ReputationregistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type ReputationregistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReputationregistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ReputationregistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReputationregistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ReputationregistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReputationregistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ReputationregistrySession struct {
	Contract     *Reputationregistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ReputationregistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ReputationregistryCallerSession struct {
	Contract *ReputationregistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// ReputationregistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ReputationregistryTransactorSession struct {
	Contract     *ReputationregistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// ReputationregistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type ReputationregistryRaw struct {
	Contract *Reputationregistry // Generic contract binding to access the raw methods on
}

// ReputationregistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ReputationregistryCallerRaw struct {
	Contract *ReputationregistryCaller // Generic read-only contract binding to access the raw methods on
}

// ReputationregistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ReputationregistryTransactorRaw struct {
	Contract *ReputationregistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewReputationregistry creates a new instance of Reputationregistry, bound to a specific deployed contract.
func NewReputationregistry(address common.Address, backend bind.ContractBackend) (*Reputationregistry, error) {
	contract, err := bindReputationregistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Reputationregistry{ReputationregistryCaller: ReputationregistryCaller{contract: contract}, ReputationregistryTransactor: ReputationregistryTransactor{contract: contract}, ReputationregistryFilterer: ReputationregistryFilterer{contract: contract}}, nil
}

// NewReputationregistryCaller creates a new read-only instance of Reputationregistry, bound to a specific deployed contract.
func NewReputationregistryCaller(address common.Address, caller bind.ContractCaller) (*ReputationregistryCaller, error) {
	contract, err := bindReputationregistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ReputationregistryCaller{contract: contract}, nil
}

// NewReputationregistryTransactor creates a new write-only instance of Reputationregistry, bound to a specific deployed contract.
func NewReputationregistryTransactor(address common.Address, transactor bind.ContractTransactor) (*ReputationregistryTransactor, error) {
	contract, err := bindReputationregistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ReputationregistryTransactor{contract: contract}, nil
}

// NewReputationregistryFilterer creates a new log filterer instance of Reputationregistry, bound to a specific deployed contract.
func NewReputationregistryFilterer(address common.Address, filterer bind.ContractFilterer) (*ReputationregistryFilterer, error) {
	contract, err := bindReputationregistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ReputationregistryFilterer{contract: contract}, nil
}

// bindReputationregistry binds a generic wrapper to an already deployed contract.
func bindReputationregistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ReputationregistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Reputationregistry *ReputationregistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Reputationregistry.Contract.ReputationregistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Reputationregistry *ReputationregistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Reputationregistry.Contract.ReputationregistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Reputationregistry *ReputationregistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Reputationregistry.Contract.ReputationregistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Reputationregistry *ReputationregistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Reputationregistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Reputationregistry *ReputationregistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Reputationregistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Reputationregistry *ReputationregistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Reputationregistry.Contract.contract.Transact(opts, method, params...)
}

// GetAverageScore is a free data retrieval call binding the contract method 0xde3a2b59.
//
// Solidity: function getAverageScore(uint256 agentId) view returns(uint256)
func (_Reputationregistry *ReputationregistryCaller) GetAverageScore(opts *bind.CallOpts, agentId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Reputationregistry.contract.Call(opts, &out, "getAverageScore", agentId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAverageScore is a free data retrieval call binding the contract method 0xde3a2b59.
//
// Solidity: function getAverageScore(uint256 agentId) view returns(uint256)
func (_Reputationregistry *ReputationregistrySession) GetAverageScore(agentId *big.Int) (*big.Int, error) {
	return _Reputationregistry.Contract.GetAverageScore(&_Reputationregistry.CallOpts, agentId)
}

// GetAverageScore is a free data retrieval call binding the contract method 0xde3a2b59.
//
// Solidity: function getAverageScore(uint256 agentId) view returns(uint256)
func (_Reputationregistry *ReputationregistryCallerSession) GetAverageScore(agentId *big.Int) (*big.Int, error) {
	return _Reputationregistry.Contract.GetAverageScore(&_Reputationregistry.CallOpts, agentId)
}

// GetFeedbackHistory is a free data retrieval call binding the contract method 0x2d77b861.
//
// Solidity: function getFeedbackHistory(uint256 agentId) view returns((address,uint8,bytes32,string,uint8,uint256)[])
func (_Reputationregistry *ReputationregistryCaller) GetFeedbackHistory(opts *bind.CallOpts, agentId *big.Int) ([]Struct0, error) {
	var out []interface{}
	err := _Reputationregistry.contract.Call(opts, &out, "getFeedbackHistory", agentId)

	if err != nil {
		return *new([]Struct0), err
	}

	out0 := *abi.ConvertType(out[0], new([]Struct0)).(*[]Struct0)

	return out0, err

}

// GetFeedbackHistory is a free data retrieval call binding the contract method 0x2d77b861.
//
// Solidity: function getFeedbackHistory(uint256 agentId) view returns((address,uint8,bytes32,string,uint8,uint256)[])
func (_Reputationregistry *ReputationregistrySession) GetFeedbackHistory(agentId *big.Int) ([]Struct0, error) {
	return _Reputationregistry.Contract.GetFeedbackHistory(&_Reputationregistry.CallOpts, agentId)
}

// GetFeedbackHistory is a free data retrieval call binding the contract method 0x2d77b861.
//
// Solidity: function getFeedbackHistory(uint256 agentId) view returns((address,uint8,bytes32,string,uint8,uint256)[])
func (_Reputationregistry *ReputationregistryCallerSession) GetFeedbackHistory(agentId *big.Int) ([]Struct0, error) {
	return _Reputationregistry.Contract.GetFeedbackHistory(&_Reputationregistry.CallOpts, agentId)
}

// HasRated is a free data retrieval call binding the contract method 0x76454486.
//
// Solidity: function hasRated(uint256 agentId, address rater) view returns(bool)
func (_Reputationregistry *ReputationregistryCaller) HasRated(opts *bind.CallOpts, agentId *big.Int, rater common.Address) (bool, error) {
	var out []interface{}
	err := _Reputationregistry.contract.Call(opts, &out, "hasRated", agentId, rater)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRated is a free data retrieval call binding the contract method 0x76454486.
//
// Solidity: function hasRated(uint256 agentId, address rater) view returns(bool)
func (_Reputationregistry *ReputationregistrySession) HasRated(agentId *big.Int, rater common.Address) (bool, error) {
	return _Reputationregistry.Contract.HasRated(&_Reputationregistry.CallOpts, agentId, rater)
}

// HasRated is a free data retrieval call binding the contract method 0x76454486.
//
// Solidity: function hasRated(uint256 agentId, address rater) view returns(bool)
func (_Reputationregistry *ReputationregistryCallerSession) HasRated(agentId *big.Int, rater common.Address) (bool, error) {
	return _Reputationregistry.Contract.HasRated(&_Reputationregistry.CallOpts, agentId, rater)
}

// SubmitFeedback is a paid mutator transaction binding the contract method 0xd6bfc6ea.
//
// Solidity: function submitFeedback(uint256 agentId, uint8 score, bytes32 outcomeRef, string comment, uint8 feedbackType) returns()
func (_Reputationregistry *ReputationregistryTransactor) SubmitFeedback(opts *bind.TransactOpts, agentId *big.Int, score uint8, outcomeRef [32]byte, comment string, feedbackType uint8) (*types.Transaction, error) {
	return _Reputationregistry.contract.Transact(opts, "submitFeedback", agentId, score, outcomeRef, comment, feedbackType)
}

// SubmitFeedback is a paid mutator transaction binding the contract method 0xd6bfc6ea.
//
// Solidity: function submitFeedback(uint256 agentId, uint8 score, bytes32 outcomeRef, string comment, uint8 feedbackType) returns()
func (_Reputationregistry *ReputationregistrySession) SubmitFeedback(agentId *big.Int, score uint8, outcomeRef [32]byte, comment string, feedbackType uint8) (*types.Transaction, error) {
	return _Reputationregistry.Contract.SubmitFeedback(&_Reputationregistry.TransactOpts, agentId, score, outcomeRef, comment, feedbackType)
}

// SubmitFeedback is a paid mutator transaction binding the contract method 0xd6bfc6ea.
//
// Solidity: function submitFeedback(uint256 agentId, uint8 score, bytes32 outcomeRef, string comment, uint8 feedbackType) returns()
func (_Reputationregistry *ReputationregistryTransactorSession) SubmitFeedback(agentId *big.Int, score uint8, outcomeRef [32]byte, comment string, feedbackType uint8) (*types.Transaction, error) {
	return _Reputationregistry.Contract.SubmitFeedback(&_Reputationregistry.TransactOpts, agentId, score, outcomeRef, comment, feedbackType)
}
