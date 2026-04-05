// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package validationregistry

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
	AgentId        *big.Int
	CheckpointHash [32]byte
	Score          uint8
	ProofType      uint8
	Proof          []byte
	Notes          string
}

// ValidationregistryMetaData contains all meta data concerning the Validationregistry contract.
var ValidationregistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"checkpointHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"score\",\"type\":\"uint8\"},{\"internalType\":\"string\",\"name\":\"notes\",\"type\":\"string\"}],\"name\":\"postEIP712Attestation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"checkpointHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"score\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"proofType\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"proof\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"notes\",\"type\":\"string\"}],\"name\":\"postAttestation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"}],\"name\":\"getAttestations\",\"outputs\":[{\"internalType\":\"tuple\",\"name\":\"\",\"components\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"checkpointHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"score\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"proofType\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"proof\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"notes\",\"type\":\"string\"}],\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"}],\"name\":\"getAverageValidationScore\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ValidationregistryABI is the input ABI used to generate the binding from.
// Deprecated: Use ValidationregistryMetaData.ABI instead.
var ValidationregistryABI = ValidationregistryMetaData.ABI

// Validationregistry is an auto generated Go binding around an Ethereum contract.
type Validationregistry struct {
	ValidationregistryCaller     // Read-only binding to the contract
	ValidationregistryTransactor // Write-only binding to the contract
	ValidationregistryFilterer   // Log filterer for contract events
}

// ValidationregistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type ValidationregistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidationregistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ValidationregistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidationregistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ValidationregistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidationregistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ValidationregistrySession struct {
	Contract     *Validationregistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ValidationregistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ValidationregistryCallerSession struct {
	Contract *ValidationregistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// ValidationregistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ValidationregistryTransactorSession struct {
	Contract     *ValidationregistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// ValidationregistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type ValidationregistryRaw struct {
	Contract *Validationregistry // Generic contract binding to access the raw methods on
}

// ValidationregistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ValidationregistryCallerRaw struct {
	Contract *ValidationregistryCaller // Generic read-only contract binding to access the raw methods on
}

// ValidationregistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ValidationregistryTransactorRaw struct {
	Contract *ValidationregistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewValidationregistry creates a new instance of Validationregistry, bound to a specific deployed contract.
func NewValidationregistry(address common.Address, backend bind.ContractBackend) (*Validationregistry, error) {
	contract, err := bindValidationregistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Validationregistry{ValidationregistryCaller: ValidationregistryCaller{contract: contract}, ValidationregistryTransactor: ValidationregistryTransactor{contract: contract}, ValidationregistryFilterer: ValidationregistryFilterer{contract: contract}}, nil
}

// NewValidationregistryCaller creates a new read-only instance of Validationregistry, bound to a specific deployed contract.
func NewValidationregistryCaller(address common.Address, caller bind.ContractCaller) (*ValidationregistryCaller, error) {
	contract, err := bindValidationregistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ValidationregistryCaller{contract: contract}, nil
}

// NewValidationregistryTransactor creates a new write-only instance of Validationregistry, bound to a specific deployed contract.
func NewValidationregistryTransactor(address common.Address, transactor bind.ContractTransactor) (*ValidationregistryTransactor, error) {
	contract, err := bindValidationregistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ValidationregistryTransactor{contract: contract}, nil
}

// NewValidationregistryFilterer creates a new log filterer instance of Validationregistry, bound to a specific deployed contract.
func NewValidationregistryFilterer(address common.Address, filterer bind.ContractFilterer) (*ValidationregistryFilterer, error) {
	contract, err := bindValidationregistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ValidationregistryFilterer{contract: contract}, nil
}

// bindValidationregistry binds a generic wrapper to an already deployed contract.
func bindValidationregistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ValidationregistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Validationregistry *ValidationregistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Validationregistry.Contract.ValidationregistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Validationregistry *ValidationregistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Validationregistry.Contract.ValidationregistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Validationregistry *ValidationregistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Validationregistry.Contract.ValidationregistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Validationregistry *ValidationregistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Validationregistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Validationregistry *ValidationregistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Validationregistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Validationregistry *ValidationregistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Validationregistry.Contract.contract.Transact(opts, method, params...)
}

// GetAttestations is a free data retrieval call binding the contract method 0x3ee4d3d2.
//
// Solidity: function getAttestations(uint256 agentId) view returns((uint256,bytes32,uint8,uint8,bytes,string)[])
func (_Validationregistry *ValidationregistryCaller) GetAttestations(opts *bind.CallOpts, agentId *big.Int) ([]Struct0, error) {
	var out []interface{}
	err := _Validationregistry.contract.Call(opts, &out, "getAttestations", agentId)

	if err != nil {
		return *new([]Struct0), err
	}

	out0 := *abi.ConvertType(out[0], new([]Struct0)).(*[]Struct0)

	return out0, err

}

// GetAttestations is a free data retrieval call binding the contract method 0x3ee4d3d2.
//
// Solidity: function getAttestations(uint256 agentId) view returns((uint256,bytes32,uint8,uint8,bytes,string)[])
func (_Validationregistry *ValidationregistrySession) GetAttestations(agentId *big.Int) ([]Struct0, error) {
	return _Validationregistry.Contract.GetAttestations(&_Validationregistry.CallOpts, agentId)
}

// GetAttestations is a free data retrieval call binding the contract method 0x3ee4d3d2.
//
// Solidity: function getAttestations(uint256 agentId) view returns((uint256,bytes32,uint8,uint8,bytes,string)[])
func (_Validationregistry *ValidationregistryCallerSession) GetAttestations(agentId *big.Int) ([]Struct0, error) {
	return _Validationregistry.Contract.GetAttestations(&_Validationregistry.CallOpts, agentId)
}

// GetAverageValidationScore is a free data retrieval call binding the contract method 0x17876c65.
//
// Solidity: function getAverageValidationScore(uint256 agentId) view returns(uint256)
func (_Validationregistry *ValidationregistryCaller) GetAverageValidationScore(opts *bind.CallOpts, agentId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Validationregistry.contract.Call(opts, &out, "getAverageValidationScore", agentId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAverageValidationScore is a free data retrieval call binding the contract method 0x17876c65.
//
// Solidity: function getAverageValidationScore(uint256 agentId) view returns(uint256)
func (_Validationregistry *ValidationregistrySession) GetAverageValidationScore(agentId *big.Int) (*big.Int, error) {
	return _Validationregistry.Contract.GetAverageValidationScore(&_Validationregistry.CallOpts, agentId)
}

// GetAverageValidationScore is a free data retrieval call binding the contract method 0x17876c65.
//
// Solidity: function getAverageValidationScore(uint256 agentId) view returns(uint256)
func (_Validationregistry *ValidationregistryCallerSession) GetAverageValidationScore(agentId *big.Int) (*big.Int, error) {
	return _Validationregistry.Contract.GetAverageValidationScore(&_Validationregistry.CallOpts, agentId)
}

// PostAttestation is a paid mutator transaction binding the contract method 0x7d91de26.
//
// Solidity: function postAttestation(uint256 agentId, bytes32 checkpointHash, uint8 score, uint8 proofType, bytes proof, string notes) returns()
func (_Validationregistry *ValidationregistryTransactor) PostAttestation(opts *bind.TransactOpts, agentId *big.Int, checkpointHash [32]byte, score uint8, proofType uint8, proof []byte, notes string) (*types.Transaction, error) {
	return _Validationregistry.contract.Transact(opts, "postAttestation", agentId, checkpointHash, score, proofType, proof, notes)
}

// PostAttestation is a paid mutator transaction binding the contract method 0x7d91de26.
//
// Solidity: function postAttestation(uint256 agentId, bytes32 checkpointHash, uint8 score, uint8 proofType, bytes proof, string notes) returns()
func (_Validationregistry *ValidationregistrySession) PostAttestation(agentId *big.Int, checkpointHash [32]byte, score uint8, proofType uint8, proof []byte, notes string) (*types.Transaction, error) {
	return _Validationregistry.Contract.PostAttestation(&_Validationregistry.TransactOpts, agentId, checkpointHash, score, proofType, proof, notes)
}

// PostAttestation is a paid mutator transaction binding the contract method 0x7d91de26.
//
// Solidity: function postAttestation(uint256 agentId, bytes32 checkpointHash, uint8 score, uint8 proofType, bytes proof, string notes) returns()
func (_Validationregistry *ValidationregistryTransactorSession) PostAttestation(agentId *big.Int, checkpointHash [32]byte, score uint8, proofType uint8, proof []byte, notes string) (*types.Transaction, error) {
	return _Validationregistry.Contract.PostAttestation(&_Validationregistry.TransactOpts, agentId, checkpointHash, score, proofType, proof, notes)
}

// PostEIP712Attestation is a paid mutator transaction binding the contract method 0x0a28a5e6.
//
// Solidity: function postEIP712Attestation(uint256 agentId, bytes32 checkpointHash, uint8 score, string notes) returns()
func (_Validationregistry *ValidationregistryTransactor) PostEIP712Attestation(opts *bind.TransactOpts, agentId *big.Int, checkpointHash [32]byte, score uint8, notes string) (*types.Transaction, error) {
	return _Validationregistry.contract.Transact(opts, "postEIP712Attestation", agentId, checkpointHash, score, notes)
}

// PostEIP712Attestation is a paid mutator transaction binding the contract method 0x0a28a5e6.
//
// Solidity: function postEIP712Attestation(uint256 agentId, bytes32 checkpointHash, uint8 score, string notes) returns()
func (_Validationregistry *ValidationregistrySession) PostEIP712Attestation(agentId *big.Int, checkpointHash [32]byte, score uint8, notes string) (*types.Transaction, error) {
	return _Validationregistry.Contract.PostEIP712Attestation(&_Validationregistry.TransactOpts, agentId, checkpointHash, score, notes)
}

// PostEIP712Attestation is a paid mutator transaction binding the contract method 0x0a28a5e6.
//
// Solidity: function postEIP712Attestation(uint256 agentId, bytes32 checkpointHash, uint8 score, string notes) returns()
func (_Validationregistry *ValidationregistryTransactorSession) PostEIP712Attestation(agentId *big.Int, checkpointHash [32]byte, score uint8, notes string) (*types.Transaction, error) {
	return _Validationregistry.Contract.PostEIP712Attestation(&_Validationregistry.TransactOpts, agentId, checkpointHash, score, notes)
}
