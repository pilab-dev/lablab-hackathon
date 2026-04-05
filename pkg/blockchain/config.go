package blockchain

type ContractAddresses struct {
	AgentRegistry      string
	HackathonVault     string
	RiskRouter         string
	ReputationRegistry string
	ValidationRegistry string
}

type Config struct {
	RPCURL       string
	ChainID      uint64
	OperatorPK   string
	AgentPK      string
	AgentID      string
	Contracts    ContractAddresses
	GasLimit     uint64
	GasPriceGwei uint64
}
