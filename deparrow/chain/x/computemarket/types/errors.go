package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/computemarket module sentinel errors
var (
	ErrInvalidProvider       = sdkerrors.Register(ModuleName, 1, "invalid provider")
	ErrProviderNotFound      = sdkerrors.Register(ModuleName, 2, "provider not found")
	ErrProviderAlreadyExists = sdkerrors.Register(ModuleName, 3, "provider already exists")
	ErrProviderInactive      = sdkerrors.Register(ModuleName, 4, "provider is inactive")
	ErrInsufficientStake     = sdkerrors.Register(ModuleName, 5, "insufficient stake")
	ErrInvalidStake          = sdkerrors.Register(ModuleName, 6, "invalid stake amount")
	ErrEscrowNotFound        = sdkerrors.Register(ModuleName, 7, "escrow not found")
	ErrEscrowLocked          = sdkerrors.Register(ModuleName, 8, "escrow is locked")
	ErrEscrowReleased        = sdkerrors.Register(ModuleName, 9, "escrow already released")
	ErrEscrowExpired         = sdkerrors.Register(ModuleName, 10, "escrow has expired")
	ErrEscrowDisputed        = sdkerrors.Register(ModuleName, 11, "escrow is disputed")
	ErrInvalidEscrowStatus   = sdkerrors.Register(ModuleName, 12, "invalid escrow status")
	ErrDisputeNotFound       = sdkerrors.Register(ModuleName, 13, "dispute not found")
	ErrDisputeAlreadyResolved = sdkerrors.Register(ModuleName, 14, "dispute already resolved")
	ErrDisputePeriodExpired  = sdkerrors.Register(ModuleName, 15, "dispute period expired")
	ErrUnauthorizedEscrow    = sdkerrors.Register(ModuleName, 16, "unauthorized escrow operation")
	ErrInvalidJobID          = sdkerrors.Register(ModuleName, 17, "invalid job ID")
	ErrJobMatchNotFound      = sdkerrors.Register(ModuleName, 18, "job match not found")
	ErrNoProvidersAvailable  = sdkerrors.Register(ModuleName, 19, "no providers available")
	ErrInvalidCapabilities   = sdkerrors.Register(ModuleName, 20, "invalid capabilities")
	ErrReputationTooLow      = sdkerrors.Register(ModuleName, 21, "reputation score too low")
	ErrSlashingFailed        = sdkerrors.Register(ModuleName, 22, "slashing failed")
	ErrInvalidDisputeReason  = sdkerrors.Register(ModuleName, 23, "invalid dispute reason")
	ErrProviderSlashed       = sdkerrors.Register(ModuleName, 24, "provider has been slashed")
	ErrInvalidAddress        = sdkerrors.Register(ModuleName, 25, "invalid address")
	ErrInsufficientFunds     = sdkerrors.Register(ModuleName, 26, "insufficient funds")
	ErrInvalidParams         = sdkerrors.Register(ModuleName, 27, "invalid module parameters")
	ErrMatchingFailed        = sdkerrors.Register(ModuleName, 28, "job matching failed")
)