package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/proofofcompute module sentinel errors
var (
	ErrInvalidJobID          = sdkerrors.Register(ModuleName, 1, "invalid job ID")
	ErrJobNotFound           = sdkerrors.Register(ModuleName, 2, "job not found")
	ErrInvalidSubmitter      = sdkerrors.Register(ModuleName, 3, "invalid submitter address")
	ErrInvalidComputeNode    = sdkerrors.Register(ModuleName, 4, "invalid compute node address")
	ErrInvalidStake          = sdkerrors.Register(ModuleName, 5, "invalid stake amount")
	ErrInvalidProof          = sdkerrors.Register(ModuleName, 6, "invalid compute proof")
	ErrProofVerification     = sdkerrors.Register(ModuleName, 7, "proof verification failed")
	ErrJobAlreadyCompleted   = sdkerrors.Register(ModuleName, 8, "job already completed")
	ErrJobNotRunning         = sdkerrors.Register(ModuleName, 9, "job is not running")
	ErrInsufficientCompute   = sdkerrors.Register(ModuleName, 10, "insufficient compute units")
	ErrMaxSupplyExceeded     = sdkerrors.Register(ModuleName, 11, "max DPC supply exceeded")
	ErrInvalidComputeUnits   = sdkerrors.Register(ModuleName, 12, "invalid compute units")
	ErrJobCancelled          = sdkerrors.Register(ModuleName, 13, "job was cancelled")
	ErrJobFailed             = sdkerrors.Register(ModuleName, 14, "job failed")
	ErrUnauthorized          = sdkerrors.Register(ModuleName, 15, "unauthorized operation")
	ErrInvalidSignature      = sdkerrors.Register(ModuleName, 16, "invalid signature")
	ErrOutputHashMismatch    = sdkerrors.Register(ModuleName, 17, "output hash mismatch")
	ErrInvalidReward         = sdkerrors.Register(ModuleName, 18, "invalid reward amount")
	ErrDuplicateProof        = sdkerrors.Register(ModuleName, 19, "duplicate proof submitted")
	ErrDifficultyTooHigh     = sdkerrors.Register(ModuleName, 20, "difficulty too high")
	ErrInvalidDifficulty     = sdkerrors.Register(ModuleName, 21, "invalid difficulty value")
)
