package types

import "errors"

// Module errors
var (
	// ErrWalletNotFound is returned when a wallet cannot be found
	ErrWalletNotFound = errors.New("wallet not found")
	// ErrWalletAlreadyExists is returned when wallet already exists
	ErrWalletAlreadyExists = errors.New("wallet already exists")
	// ErrInsufficientBalance is returned when balance is insufficient
	ErrInsufficientBalance = errors.New("insufficient balance")
	// ErrSpendingRuleViolation is returned when spending violates rules
	ErrSpendingRuleViolation = errors.New("spending rule violation")
	// ErrMaxRulesExceeded is returned when max rules are exceeded
	ErrMaxRulesExceeded = errors.New("maximum rules exceeded")
	// ErrInvalidDID is returned when DID is invalid
	ErrInvalidDID = errors.New("invalid DID")
	// ErrInvalidAddress is returned when address is invalid
	ErrInvalidAddress = errors.New("invalid address")
	// ErrOperationNotAllowed is returned when operation is blocked
	ErrOperationNotAllowed = errors.New("operation not allowed")
	// ErrEmergencyReserve is returned when emergency reserve is too low
	ErrEmergencyReserve = errors.New("emergency reserve below minimum")
)
