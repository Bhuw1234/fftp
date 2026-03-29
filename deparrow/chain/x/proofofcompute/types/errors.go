package types

import (
	"errors"
)

// Module errors
var (
	// ErrJobNotFound is returned when a job cannot be found
	ErrJobNotFound = errors.New("job not found")
	// ErrInvalidProof is returned when a proof is invalid
	ErrInvalidProof = errors.New("invalid proof")
	// ErrInsufficientStake is returned when stake is below minimum
	ErrInsufficientStake = errors.New("insufficient stake")
	// ErrJobAlreadyCompleted is returned when trying to modify a completed job
	ErrJobAlreadyCompleted = errors.New("job already completed")
	// ErrJobNotRunning is returned when trying to submit proof for non-running job
	ErrJobNotRunning = errors.New("job not in running state")
	// ErrInvalidComputeUnits is returned when compute units are invalid
	ErrInvalidComputeUnits = errors.New("invalid compute units")
	// ErrInvalidSubmitter is returned when submitter doesn't match
	ErrInvalidSubmitter = errors.New("invalid submitter")
	// ErrMaxSupplyExceeded is returned when minting would exceed max supply
	ErrMaxSupplyExceeded = errors.New("max supply exceeded")
	// ErrNodeNotFound is returned when a node cannot be found
	ErrNodeNotFound = errors.New("node not found")
	// ErrNoPendingReward is returned when node has no pending rewards
	ErrNoPendingReward = errors.New("no pending rewards")
)
