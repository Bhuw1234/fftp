package types

import (
	context "context"
)

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// Wallet management
	CreateWallet(context.Context, *MsgCreateWallet) (*MsgCreateWalletResponse, error)
	DeleteWallet(context.Context, *MsgDeleteWallet) (*MsgDeleteWalletResponse, error)
	
	// Fund operations
	Deposit(context.Context, *MsgDeposit) (*MsgDepositResponse, error)
	Withdraw(context.Context, *MsgWithdraw) (*MsgWithdrawResponse, error)
	Transfer(context.Context, *MsgTransfer) (*MsgTransferResponse, error)
	
	// Autonomous operations
	AutonomousSpend(context.Context, *MsgAutonomousSpend) (*MsgAutonomousSpendResponse, error)
	
	// Spending rules
	AddSpendingRule(context.Context, *MsgAddSpendingRule) (*MsgAddSpendingRuleResponse, error)
	UpdateSpendingRule(context.Context, *MsgUpdateSpendingRule) (*MsgUpdateSpendingRuleResponse, error)
	RemoveSpendingRule(context.Context, *MsgRemoveSpendingRule) (*MsgRemoveSpendingRuleResponse, error)
	
	// Automation rules
	AddAutomationRule(context.Context, *MsgAddAutomationRule) (*MsgAddAutomationRuleResponse, error)
	UpdateAutomationRule(context.Context, *MsgUpdateAutomationRule) (*MsgUpdateAutomationRuleResponse, error)
	RemoveAutomationRule(context.Context, *MsgRemoveAutomationRule) (*MsgRemoveAutomationRuleResponse, error)
	
	// Emergency reserve
	SetEmergencyReserve(context.Context, *MsgSetEmergencyReserve) (*MsgSetEmergencyReserveResponse, error)
	
	// Agent management
	RegisterAgent(context.Context, *MsgRegisterAgent) (*MsgRegisterAgentResponse, error)
	UnregisterAgent(context.Context, *MsgUnregisterAgent) (*MsgUnregisterAgentResponse, error)
	
	// Governance
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

// Response types

// MsgCreateWalletResponse is the response for MsgCreateWallet
type MsgCreateWalletResponse struct {
	DID     string `json:"did"`
	Address string `json:"address"`
}

// MsgDeleteWalletResponse is the response for MsgDeleteWallet
type MsgDeleteWalletResponse struct {
	WithdrawnAmount string `json:"withdrawn_amount"`
}

// MsgDepositResponse is the response for MsgDeposit
type MsgDepositResponse struct {
	Balance string `json:"balance"`
}

// MsgWithdrawResponse is the response for MsgWithdraw
type MsgWithdrawResponse struct {
	Balance string `json:"balance"`
}

// MsgTransferResponse is the response for MsgTransfer
type MsgTransferResponse struct {
	SenderBalance    string `json:"sender_balance"`
	RecipientBalance string `json:"recipient_balance"`
}

// MsgAutonomousSpendResponse is the response for MsgAutonomousSpend
type MsgAutonomousSpendResponse struct {
	Success bool   `json:"success"`
	Balance string `json:"balance"`
	TxHash  string `json:"tx_hash"`
}

// MsgAddSpendingRuleResponse is the response for MsgAddSpendingRule
type MsgAddSpendingRuleResponse struct {
	RuleIndex uint32 `json:"rule_index"`
}

// MsgUpdateSpendingRuleResponse is the response for MsgUpdateSpendingRule
type MsgUpdateSpendingRuleResponse struct {
	Success bool `json:"success"`
}

// MsgRemoveSpendingRuleResponse is the response for MsgRemoveSpendingRule
type MsgRemoveSpendingRuleResponse struct {
	Success bool `json:"success"`
}

// MsgAddAutomationRuleResponse is the response for MsgAddAutomationRule
type MsgAddAutomationRuleResponse struct {
	RuleIndex uint32 `json:"rule_index"`
}

// MsgUpdateAutomationRuleResponse is the response for MsgUpdateAutomationRule
type MsgUpdateAutomationRuleResponse struct {
	Success bool `json:"success"`
}

// MsgRemoveAutomationRuleResponse is the response for MsgRemoveAutomationRule
type MsgRemoveAutomationRuleResponse struct {
	Success bool `json:"success"`
}

// MsgSetEmergencyReserveResponse is the response for MsgSetEmergencyReserve
type MsgSetEmergencyReserveResponse struct {
	Reserve string `json:"reserve"`
}

// MsgRegisterAgentResponse is the response for MsgRegisterAgent
type MsgRegisterAgentResponse struct {
	DID     string `json:"did"`
	Address string `json:"address"`
}

// MsgUnregisterAgentResponse is the response for MsgUnregisterAgent
type MsgUnregisterAgentResponse struct {
	Success bool `json:"success"`
}

// MsgUpdateParamsResponse is the response for MsgUpdateParams
type MsgUpdateParamsResponse struct {
	Success bool `json:"success"`
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct{}

func (UnimplementedMsgServer) CreateWallet(ctx context.Context, req *MsgCreateWallet) (*MsgCreateWalletResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) DeleteWallet(ctx context.Context, req *MsgDeleteWallet) (*MsgDeleteWalletResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) Deposit(ctx context.Context, req *MsgDeposit) (*MsgDepositResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) Withdraw(ctx context.Context, req *MsgWithdraw) (*MsgWithdrawResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) Transfer(ctx context.Context, req *MsgTransfer) (*MsgTransferResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) AutonomousSpend(ctx context.Context, req *MsgAutonomousSpend) (*MsgAutonomousSpendResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) AddSpendingRule(ctx context.Context, req *MsgAddSpendingRule) (*MsgAddSpendingRuleResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) UpdateSpendingRule(ctx context.Context, req *MsgUpdateSpendingRule) (*MsgUpdateSpendingRuleResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) RemoveSpendingRule(ctx context.Context, req *MsgRemoveSpendingRule) (*MsgRemoveSpendingRuleResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) AddAutomationRule(ctx context.Context, req *MsgAddAutomationRule) (*MsgAddAutomationRuleResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) UpdateAutomationRule(ctx context.Context, req *MsgUpdateAutomationRule) (*MsgUpdateAutomationRuleResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) RemoveAutomationRule(ctx context.Context, req *MsgRemoveAutomationRule) (*MsgRemoveAutomationRuleResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) SetEmergencyReserve(ctx context.Context, req *MsgSetEmergencyReserve) (*MsgSetEmergencyReserveResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) RegisterAgent(ctx context.Context, req *MsgRegisterAgent) (*MsgRegisterAgentResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) UnregisterAgent(ctx context.Context, req *MsgUnregisterAgent) (*MsgUnregisterAgentResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) UpdateParams(ctx context.Context, req *MsgUpdateParams) (*MsgUpdateParamsResponse, error) {
	return nil, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// Wallet queries
	Wallet(context.Context, *QueryWalletRequest) (*QueryWalletResponse, error)
	WalletByDID(context.Context, *QueryWalletByDIDRequest) (*QueryWalletResponse, error)
	Wallets(context.Context, *QueryWalletsRequest) (*QueryWalletsResponse, error)
	
	// Balance queries
	Balance(context.Context, *QueryBalanceRequest) (*QueryBalanceResponse, error)
	AvailableBalance(context.Context, *QueryAvailableBalanceRequest) (*QueryAvailableBalanceResponse, error)
	
	// Spending rule queries
	SpendingRules(context.Context, *QuerySpendingRulesRequest) (*QuerySpendingRulesResponse, error)
	DailySpending(context.Context, *QueryDailySpendingRequest) (*QueryDailySpendingResponse, error)
	
	// Automation rule queries
	AutomationRules(context.Context, *QueryAutomationRulesRequest) (*QueryAutomationRulesResponse, error)
	PendingAutomations(context.Context, *QueryPendingAutomationsRequest) (*QueryPendingAutomationsResponse, error)
	
	// Agent queries
	Agent(context.Context, *QueryAgentRequest) (*QueryAgentResponse, error)
	Agents(context.Context, *QueryAgentsRequest) (*QueryAgentsResponse, error)
	AgentByAddress(context.Context, *QueryAgentByAddressRequest) (*QueryAgentResponse, error)
	
	// Stats and params
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	Stats(context.Context, *QueryStatsRequest) (*QueryStatsResponse, error)
}

// Query request types

type QueryWalletRequest struct {
	Address string `json:"address"`
}

type QueryWalletByDIDRequest struct {
	Did string `json:"did"`
}

type QueryWalletsRequest struct {
	Page     uint64 `json:"page"`
	PageSize uint64 `json:"page_size"`
}

type QueryBalanceRequest struct {
	Address string `json:"address"`
}

type QueryAvailableBalanceRequest struct {
	Address string `json:"address"`
}

type QuerySpendingRulesRequest struct {
	Address string `json:"address"`
}

type QueryDailySpendingRequest struct {
	Address string `json:"address"`
}

type QueryAutomationRulesRequest struct {
	Address string `json:"address"`
}

type QueryPendingAutomationsRequest struct {
	Address string `json:"address,omitempty"`
}

type QueryAgentRequest struct {
	Did string `json:"did"`
}

type QueryAgentsRequest struct {
	Page     uint64 `json:"page"`
	PageSize uint64 `json:"page_size"`
	Active   bool   `json:"active"`
}

type QueryAgentByAddressRequest struct {
	Address string `json:"address"`
}

type QueryParamsRequest struct{}

type QueryStatsRequest struct{}

// Query response types

type QueryWalletResponse struct {
	Wallet AgentWalletExtended `json:"wallet"`
}

type QueryWalletsResponse struct {
	Wallets    []AgentWalletExtended `json:"wallets"`
	TotalCount uint64                `json:"total_count"`
}

type QueryBalanceResponse struct {
	Balance          string `json:"balance"`
	AvailableBalance string `json:"available_balance"`
	EmergencyReserve string `json:"emergency_reserve"`
}

type QueryAvailableBalanceResponse struct {
	AvailableBalance string `json:"available_balance"`
}

type QuerySpendingRulesResponse struct {
	Rules []SpendingRuleExtended `json:"rules"`
}

type QueryDailySpendingResponse struct {
	DailySpent string `json:"daily_spent"`
	DailyBudget string `json:"daily_budget"`
	Remaining  string `json:"remaining"`
}

type QueryAutomationRulesResponse struct {
	Rules []AutomationRuleExtended `json:"rules"`
}

type QueryPendingAutomationsResponse struct {
	Automations []PendingAutomation `json:"automations"`
}

type QueryAgentResponse struct {
	Agent AgentInfo `json:"agent"`
}

type QueryAgentsResponse struct {
	Agents     []AgentInfo `json:"agents"`
	TotalCount uint64      `json:"total_count"`
}

type QueryParamsResponse struct {
	Params Params `json:"params"`
}

type QueryStatsResponse struct {
	TotalWallets       uint64 `json:"total_wallets"`
	TotalAgents        uint64 `json:"total_agents"`
	ActiveAgents       uint64 `json:"active_agents"`
	TotalDPCDeposited  string `json:"total_dpc_deposited"`
	TotalDPCWithdrawn  string `json:"total_dpc_withdrawn"`
	TotalTransactions  uint64 `json:"total_transactions"`
	AverageDailySpend  string `json:"average_daily_spend"`
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct{}

func (UnimplementedQueryServer) Wallet(ctx context.Context, req *QueryWalletRequest) (*QueryWalletResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) WalletByDID(ctx context.Context, req *QueryWalletByDIDRequest) (*QueryWalletResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Wallets(ctx context.Context, req *QueryWalletsRequest) (*QueryWalletsResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Balance(ctx context.Context, req *QueryBalanceRequest) (*QueryBalanceResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) AvailableBalance(ctx context.Context, req *QueryAvailableBalanceRequest) (*QueryAvailableBalanceResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) SpendingRules(ctx context.Context, req *QuerySpendingRulesRequest) (*QuerySpendingRulesResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) DailySpending(ctx context.Context, req *QueryDailySpendingRequest) (*QueryDailySpendingResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) AutomationRules(ctx context.Context, req *QueryAutomationRulesRequest) (*QueryAutomationRulesResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) PendingAutomations(ctx context.Context, req *QueryPendingAutomationsRequest) (*QueryPendingAutomationsResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Agent(ctx context.Context, req *QueryAgentRequest) (*QueryAgentResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Agents(ctx context.Context, req *QueryAgentsRequest) (*QueryAgentsResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) AgentByAddress(ctx context.Context, req *QueryAgentByAddressRequest) (*QueryAgentResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Stats(ctx context.Context, req *QueryStatsRequest) (*QueryStatsResponse, error) {
	return nil, nil
}

// Service descriptors (placeholders for proto-generated code)
var _Msg_serviceDesc = _Msg_serviceDesc_placeholder{}
var _Query_serviceDesc = _Query_serviceDesc_placeholder{}

type _Msg_serviceDesc_placeholder struct{}
type _Query_serviceDesc_placeholder struct{}
