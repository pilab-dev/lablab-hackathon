package blockchain

type ContractAddresses struct {
	AgentRegistry      string `json:"agent_registry"`
	HackathonVault     string `json:"hackathon_vault"`
	RiskRouter         string `json:"risk_router"`
	ReputationRegistry string `json:"reputation_registry"`
	ValidationRegistry string `json:"validation_registry"`
}

type Config struct {
	RPCURL       string            `json:"rpc_url"`
	ChainID      uint64            `json:"chain_id"`
	OperatorPK   string            `json:"operator_pk"`
	AgentPK      string            `json:"agent_pk"`
	AgentID      string            `json:"agent_id"`
	Contracts    ContractAddresses `json:"contracts"`
	GasLimit     uint64            `json:"gas_limit"`
	GasPriceGwei uint64            `json:"gas_price_gwei"`
}
