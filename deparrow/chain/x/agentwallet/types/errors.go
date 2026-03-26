package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/agentwallet module sentinel errors
var (
	ErrInvalidWallet          = sdkerrors.Register(ModuleName, 1, "invalid wallet")
	ErrWalletNotFound         = sdkerrors.Register(ModuleName, 2, "wallet not found")
	ErrWalletAlreadyExists    = sdkerrors.Register(ModuleName, 3, "wallet already exists")
	ErrInvalidDID             = sdkerrors.Register(ModuleName, 4, "invalid DID format")
	ErrInvalidAddress         = sdkerrors.Register(ModuleName, 5, "invalid address")
	ErrInsufficientFunds      = sdkerrors.Register(ModuleName, 6, "insufficient funds")
	ErrSpendingRuleViolation  = sdkerrors.Register(ModuleName, 7, "spending rule violation")
	ErrDailyBudgetExceeded    = sdkerrors.Register(ModuleName, 8, "daily budget exceeded")
	ErrMaxPerTxExceeded       = sdkerrors.Register(ModuleName, 9, "max per transaction exceeded")
	ErrOperationBlocked       = sdkerrors.Register(ModuleName, 10, "operation blocked by rules")
	ErrOperationNotAllowed    = sdkerrors.Register(ModuleName, 11, "operation not allowed")
	ErrEmergencyReserve       = sdkerrors.Register(ModuleName, 12, "cannot spend emergency reserve")
	ErrInvalidRule            = sdkerrors.Register(ModuleName, 13, "invalid spending rule")
	ErrRuleLimitExceeded      = sdkerrors.Register(ModuleName, 14, "rule limit exceeded")
	ErrAutomationFailed       = sdkerrors.Register(ModuleName, 15, "automation execution failed")
	ErrInvalidAutomationRule  = sdkerrors.Register(ModuleName, 16, "invalid automation rule")
	ErrAutomationDisabled     = sdkerrors.Register(ModuleName, 17, "automation rule is disabled")
	ErrInvalidTrigger         = sdkerrors.Register(ModuleName, 18, "invalid automation trigger")
	ErrInvalidAction          = sdkerrors.Register(ModuleName, 19, "invalid automation action")
	ErrUnauthorized           = sdkerrors.Register(ModuleName, 20, "unauthorized wallet operation")
	ErrExternalTransferBlocked = sdkerrors.Register(ModuleName, 21, "external transfers are blocked")
	ErrInvalidAmount          = sdkerrors.Register(ModuleName, 22, "invalid amount")
	ErrInvalidParams          = sdkerrors.Register(ModuleName, 23, "invalid module parameters")
	ErrAgentNotRegistered     = sdkerrors.Register(ModuleName, 24, "agent not registered")
	ErrAgentAlreadyRegistered = sdkerrors.Register(ModuleName, 25, "agent already registered")
	ErrInvalidSignature       = sdkerrors.Register(ModuleName, 26, "invalid autonomous signature")
	ErrSelfTransfer           = sdkerrors.Register(ModuleName, 27, "cannot transfer to self")
	ErrInsufficientReserve    = sdkerrors.Register(ModuleName, 28, "insufficient emergency reserve")
	ErrInvalidReserveAmount   = sdkerrors.Register(ModuleName, 29, "invalid reserve amount")
)
