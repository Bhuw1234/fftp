package types

import "errors"

// Module errors
var (
	// ErrProviderNotFound is returned when a provider cannot be found
	ErrProviderNotFound = errors.New("provider not found")
	// ErrProviderAlreadyRegistered is returned when provider already exists
	ErrProviderAlreadyRegistered = errors.New("provider already registered")
	// ErrProviderNotActive is returned when provider is not active
	ErrProviderNotActive = errors.New("provider not active")
	// ErrInsufficientStake is returned when stake is below minimum
	ErrInsufficientStake = errors.New("insufficient stake")
	// ErrEscrowNotFound is returned when an escrow cannot be found
	ErrEscrowNotFound = errors.New("escrow not found")
	// ErrEscrowNotLocked is returned when escrow is not in locked state
	ErrEscrowNotLocked = errors.New("escrow not locked")
	// ErrEscrowAlreadyReleased is returned when escrow is already released
	ErrEscrowAlreadyReleased = errors.New("escrow already released")
	// ErrDisputeNotFound is returned when a dispute cannot be found
	ErrDisputeNotFound = errors.New("dispute not found")
	// ErrDisputeAlreadyResolved is returned when dispute is already resolved
	ErrDisputeAlreadyResolved = errors.New("dispute already resolved")
	// ErrInvalidDisputeStatus is returned when dispute status is invalid
	ErrInvalidDisputeStatus = errors.New("invalid dispute status")
	// ErrUnauthorized is returned when caller is not authorized
	ErrUnauthorized = errors.New("unauthorized")
	// ErrInvalidAddress is returned when address is invalid
	ErrInvalidAddress = errors.New("invalid address")
	// ErrInvalidAmount is returned when amount is invalid
	ErrInvalidAmount = errors.New("invalid amount")
	// ErrLowReputation is returned when provider reputation is too low
	ErrLowReputation = errors.New("reputation below minimum")
)
