// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package agentregistry

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

// AgentregistryMetaData contains all meta data concerning the Agentregistry contract.
var AgentregistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"},{\"internalType\":\"string[]\",\"name\":\"capabilities\",\"type\":\"string[]\"},{\"internalType\":\"string\",\"name\":\"metadataURI\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"agentWallet\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operatorWallet\",\"type\":\"address\"}],\"name\":\"registerAgent\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"}],\"name\":\"isRegistered\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"agentId\",\"type\":\"uint256\"}],\"name\":\"getAgent\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operatorWallet\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"agentWallet\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"registeredAt\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// AgentregistryABI is the input ABI used to generate the binding from.
// Deprecated: Use AgentregistryMetaData.ABI instead.
var AgentregistryABI = AgentregistryMetaData.ABI

// Agentregistry is an auto generated Go binding around an Ethereum contract.
type Agentregistry struct {
	AgentregistryCaller     // Read-only binding to the contract
	AgentregistryTransactor // Write-only binding to the contract
	AgentregistryFilterer   // Log filterer for contract events
}

// AgentregistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type AgentregistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AgentregistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AgentregistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AgentregistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AgentregistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AgentregistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AgentregistrySession struct {
	Contract     *Agentregistry    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AgentregistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AgentregistryCallerSession struct {
	Contract *AgentregistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// AgentregistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AgentregistryTransactorSession struct {
	Contract     *AgentregistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// AgentregistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type AgentregistryRaw struct {
	Contract *Agentregistry // Generic contract binding to access the raw methods on
}

// AgentregistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AgentregistryCallerRaw struct {
	Contract *AgentregistryCaller // Generic read-only contract binding to access the raw methods on
}

// AgentregistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AgentregistryTransactorRaw struct {
	Contract *AgentregistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAgentregistry creates a new instance of Agentregistry, bound to a specific deployed contract.
func NewAgentregistry(address common.Address, backend bind.ContractBackend) (*Agentregistry, error) {
	contract, err := bindAgentregistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Agentregistry{AgentregistryCaller: AgentregistryCaller{contract: contract}, AgentregistryTransactor: AgentregistryTransactor{contract: contract}, AgentregistryFilterer: AgentregistryFilterer{contract: contract}}, nil
}

// NewAgentregistryCaller creates a new read-only instance of Agentregistry, bound to a specific deployed contract.
func NewAgentregistryCaller(address common.Address, caller bind.ContractCaller) (*AgentregistryCaller, error) {
	contract, err := bindAgentregistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AgentregistryCaller{contract: contract}, nil
}

// NewAgentregistryTransactor creates a new write-only instance of Agentregistry, bound to a specific deployed contract.
func NewAgentregistryTransactor(address common.Address, transactor bind.ContractTransactor) (*AgentregistryTransactor, error) {
	contract, err := bindAgentregistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AgentregistryTransactor{contract: contract}, nil
}

// NewAgentregistryFilterer creates a new log filterer instance of Agentregistry, bound to a specific deployed contract.
func NewAgentregistryFilterer(address common.Address, filterer bind.ContractFilterer) (*AgentregistryFilterer, error) {
	contract, err := bindAgentregistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AgentregistryFilterer{contract: contract}, nil
}

// bindAgentregistry binds a generic wrapper to an already deployed contract.
func bindAgentregistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AgentregistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Agentregistry *AgentregistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Agentregistry.Contract.AgentregistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Agentregistry *AgentregistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Agentregistry.Contract.AgentregistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Agentregistry *AgentregistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Agentregistry.Contract.AgentregistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Agentregistry *AgentregistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Agentregistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Agentregistry *AgentregistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Agentregistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Agentregistry *AgentregistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Agentregistry.Contract.contract.Transact(opts, method, params...)
}

// GetAgent is a free data retrieval call binding the contract method 0x2de5aaf7.
//
// Solidity: function getAgent(uint256 agentId) view returns(address owner, address operatorWallet, address agentWallet, string name, string description, uint256 registeredAt)
func (_Agentregistry *AgentregistryCaller) GetAgent(opts *bind.CallOpts, agentId *big.Int) (struct {
	Owner          common.Address
	OperatorWallet common.Address
	AgentWallet    common.Address
	Name           string
	Description    string
	RegisteredAt   *big.Int
}, error) {
	var out []interface{}
	err := _Agentregistry.contract.Call(opts, &out, "getAgent", agentId)

	outstruct := new(struct {
		Owner          common.Address
		OperatorWallet common.Address
		AgentWallet    common.Address
		Name           string
		Description    string
		RegisteredAt   *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Owner = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.OperatorWallet = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.AgentWallet = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.Name = *abi.ConvertType(out[3], new(string)).(*string)
	outstruct.Description = *abi.ConvertType(out[4], new(string)).(*string)
	outstruct.RegisteredAt = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetAgent is a free data retrieval call binding the contract method 0x2de5aaf7.
//
// Solidity: function getAgent(uint256 agentId) view returns(address owner, address operatorWallet, address agentWallet, string name, string description, uint256 registeredAt)
func (_Agentregistry *AgentregistrySession) GetAgent(agentId *big.Int) (struct {
	Owner          common.Address
	OperatorWallet common.Address
	AgentWallet    common.Address
	Name           string
	Description    string
	RegisteredAt   *big.Int
}, error) {
	return _Agentregistry.Contract.GetAgent(&_Agentregistry.CallOpts, agentId)
}

// GetAgent is a free data retrieval call binding the contract method 0x2de5aaf7.
//
// Solidity: function getAgent(uint256 agentId) view returns(address owner, address operatorWallet, address agentWallet, string name, string description, uint256 registeredAt)
func (_Agentregistry *AgentregistryCallerSession) GetAgent(agentId *big.Int) (struct {
	Owner          common.Address
	OperatorWallet common.Address
	AgentWallet    common.Address
	Name           string
	Description    string
	RegisteredAt   *big.Int
}, error) {
	return _Agentregistry.Contract.GetAgent(&_Agentregistry.CallOpts, agentId)
}

// IsRegistered is a free data retrieval call binding the contract method 0x579a6988.
//
// Solidity: function isRegistered(uint256 agentId) view returns(bool)
func (_Agentregistry *AgentregistryCaller) IsRegistered(opts *bind.CallOpts, agentId *big.Int) (bool, error) {
	var out []interface{}
	err := _Agentregistry.contract.Call(opts, &out, "isRegistered", agentId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsRegistered is a free data retrieval call binding the contract method 0x579a6988.
//
// Solidity: function isRegistered(uint256 agentId) view returns(bool)
func (_Agentregistry *AgentregistrySession) IsRegistered(agentId *big.Int) (bool, error) {
	return _Agentregistry.Contract.IsRegistered(&_Agentregistry.CallOpts, agentId)
}

// IsRegistered is a free data retrieval call binding the contract method 0x579a6988.
//
// Solidity: function isRegistered(uint256 agentId) view returns(bool)
func (_Agentregistry *AgentregistryCallerSession) IsRegistered(agentId *big.Int) (bool, error) {
	return _Agentregistry.Contract.IsRegistered(&_Agentregistry.CallOpts, agentId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Agentregistry *AgentregistryCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Agentregistry.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Agentregistry *AgentregistrySession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _Agentregistry.Contract.OwnerOf(&_Agentregistry.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Agentregistry *AgentregistryCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _Agentregistry.Contract.OwnerOf(&_Agentregistry.CallOpts, tokenId)
}

// RegisterAgent is a paid mutator transaction binding the contract method 0xf1a61f35.
//
// Solidity: function registerAgent(string name, string description, string[] capabilities, string metadataURI, address agentWallet, address operatorWallet) returns(uint256 agentId)
func (_Agentregistry *AgentregistryTransactor) RegisterAgent(opts *bind.TransactOpts, name string, description string, capabilities []string, metadataURI string, agentWallet common.Address, operatorWallet common.Address) (*types.Transaction, error) {
	return _Agentregistry.contract.Transact(opts, "registerAgent", name, description, capabilities, metadataURI, agentWallet, operatorWallet)
}

// RegisterAgent is a paid mutator transaction binding the contract method 0xf1a61f35.
//
// Solidity: function registerAgent(string name, string description, string[] capabilities, string metadataURI, address agentWallet, address operatorWallet) returns(uint256 agentId)
func (_Agentregistry *AgentregistrySession) RegisterAgent(name string, description string, capabilities []string, metadataURI string, agentWallet common.Address, operatorWallet common.Address) (*types.Transaction, error) {
	return _Agentregistry.Contract.RegisterAgent(&_Agentregistry.TransactOpts, name, description, capabilities, metadataURI, agentWallet, operatorWallet)
}

// RegisterAgent is a paid mutator transaction binding the contract method 0xf1a61f35.
//
// Solidity: function registerAgent(string name, string description, string[] capabilities, string metadataURI, address agentWallet, address operatorWallet) returns(uint256 agentId)
func (_Agentregistry *AgentregistryTransactorSession) RegisterAgent(name string, description string, capabilities []string, metadataURI string, agentWallet common.Address, operatorWallet common.Address) (*types.Transaction, error) {
	return _Agentregistry.Contract.RegisterAgent(&_Agentregistry.TransactOpts, name, description, capabilities, metadataURI, agentWallet, operatorWallet)
}
