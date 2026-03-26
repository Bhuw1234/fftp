package types

// Event types for the agentwallet module
const (
	EventTypeWalletCreated      = "wallet_created"
	EventTypeWalletUpdated      = "wallet_updated"
	EventTypeWalletDeleted      = "wallet_deleted"
	EventTypeFundsDeposited     = "funds_deposited"
	EventTypeFundsWithdrawn     = "funds_withdrawn"
	EventTypeFundsTransferred   = "funds_transferred"
	EventTypeSpendingRuleAdded  = "spending_rule_added"
	EventTypeSpendingRuleUpdated = "spending_rule_updated"
	EventTypeSpendingRuleRemoved = "spending_rule_removed"
	EventTypeAutomationRuleAdded = "automation_rule_added"
	EventTypeAutomationRuleUpdated = "automation_rule_updated"
	EventTypeAutomationRuleRemoved = "automation_rule_removed"
	EventTypeAutomationTriggered = "automation_triggered"
	EventTypeEmergencyReserveUpdated = "emergency_reserve_updated"
	EventTypeDailySpendingReset = "daily_spending_reset"
	EventTypeOperationBlocked   = "operation_blocked"
	EventTypeAgentRegistered    = "agent_registered"
	EventTypeAgentUnregistered  = "agent_unregistered"
	EventTypeAutonomousTxExecuted = "autonomous_tx_executed"

	AttributeKeyDID             = "did"
	AttributeKeyAddress         = "address"
	AttributeKeyAmount          = "amount"
	AttributeKeyBalance         = "balance"
	AttributeKeyRecipient       = "recipient"
	AttributeKeySender          = "sender"
	AttributeKeyOperation       = "operation"
	AttributeKeyRuleIndex       = "rule_index"
	AttributeKeyMaxPerTx        = "max_per_tx"
	AttributeKeyDailyBudget     = "daily_budget"
	AttributeKeyAllowedOps      = "allowed_ops"
	AttributeKeyBlockedOps      = "blocked_ops"
	AttributeKeyTrigger         = "trigger"
	AttributeKeyAction          = "action"
	AttributeKeyEnabled         = "enabled"
	AttributeKeyReserve         = "reserve"
	AttributeKeyDailySpent      = "daily_spent"
	AttributeKeyReason          = "reason"
	AttributeKeyTxHash          = "tx_hash"
	AttributeKeyTimestamp       = "timestamp"
	AttributeKeyAgentType       = "agent_type"
	AttributeKeyMetadata        = "metadata"
)

// Operation types for spending rules
const (
	OperationSubmitJob      = "submit_job"
	OperationPayService     = "pay_service"
	OperationBuyCompute     = "buy_compute"
	OperationTransfer       = "transfer"
	OperationExternalTransfer = "external_transfer"
	OperationWithdraw       = "withdraw"
	OperationDeposit        = "deposit"
)

// Automation trigger types
const (
	TriggerBalanceBelow    = "balance_below"
	TriggerBalanceAbove    = "balance_above"
	TriggerJobCompleted    = "job_completed"
	TriggerIdleTimeout     = "idle_timeout"
	TriggerScheduled       = "scheduled"
	TriggerDailyReset      = "daily_reset"
	TriggerEmergencyLow    = "emergency_low"
)

// Automation action types
const (
	ActionBuyCompute      = "buy_compute"
	ActionTransferToReserve = "transfer_to_reserve"
	ActionNotify          = "notify"
	ActionPauseOperations = "pause_operations"
	ActionResumeOperations = "resume_operations"
	ActionRequestFunding  = "request_funding"
)

// Agent types for AI agents
const (
	AgentTypeCompute  = "compute"
	AgentTypeService  = "service"
	AgentTypeOrchestrator = "orchestrator"
	AgentTypeCustom   = "custom"
)
