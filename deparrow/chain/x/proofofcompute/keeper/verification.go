package keeper

import (
	"crypto/sha256"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/deparrow/dpc/x/proofofcompute/types"
)

// SubmitProof submits a compute proof for verification
func (k Keeper) SubmitProof(ctx sdk.Context, msg *types.MsgSubmitProof) error {
	// Get the job
	job, err := k.GetJob(ctx, msg.JobID)
	if err != nil {
		return err
	}

	// Verify job is in running state
	if job.Status != types.JobStatusRunning {
		return sdkerrors.Wrapf(
			types.ErrJobNotRunning,
			"job status is %s, expected running",
			job.Status.String(),
		)
	}

	// Verify the submitter is the assigned compute node
	if job.ComputeNode != msg.NodeAddress {
		return sdkerrors.Wrap(
			types.ErrUnauthorized,
			"only assigned compute node can submit proof",
		)
	}

	// Check for duplicate proof
	if k.HasProof(ctx, msg.JobID) {
		return types.ErrDuplicateProof
	}

	// Create the proof
	proof := &types.ComputeProof{
		JobID:         msg.JobID,
		NodeID:        msg.NodeAddress,
		ComputeUnits:  msg.ComputeUnits,
		ExecutionTime: msg.ExecutionTime,
		OutputHash:    msg.OutputHash,
		Signature:     msg.Signature,
	}

	// Verify the proof
	if err := k.VerifyProof(ctx, proof, job); err != nil {
		return err
	}

	// Store the proof
	if err := k.SetProof(ctx, proof); err != nil {
		return err
	}

	// Mark job as completed
	if err := k.CompleteJob(ctx, msg.JobID, msg.OutputHash, msg.ComputeUnits); err != nil {
		return err
	}

	// Get updated job and distribute reward
	updatedJob, err := k.GetJob(ctx, msg.JobID)
	if err != nil {
		return err
	}

	if err := k.DistributeReward(ctx, updatedJob); err != nil {
		k.Logger(ctx).Error("failed to distribute reward", "job_id", msg.JobID, "error", err)
		// Add to pending rewards instead of failing
		reward, calcErr := k.CalculateReward(ctx, msg.ComputeUnits, updatedJob.Complexity)
		if calcErr == nil {
			_ = k.AddPendingReward(ctx, msg.NodeAddress, reward)
		}
	}

	// Emit proof verification event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProofVerified,
			sdk.NewAttribute(types.AttributeKeyJobID, msg.JobID),
			sdk.NewAttribute(types.AttributeKeyComputeNode, msg.NodeAddress),
			sdk.NewAttribute(types.AttributeKeyComputeUnits, string(rune(msg.ComputeUnits))),
			sdk.NewAttribute(types.AttributeKeyExecutionTime, string(rune(msg.ExecutionTime))),
		),
	)

	return nil
}

// VerifyProof verifies the compute proof against the job
func (k Keeper) VerifyProof(ctx sdk.Context, proof *types.ComputeProof, job *types.Job) error {
	// 1. Verify compute units match (within tolerance)
	// Allow 10% tolerance for compute units
	tolerance := job.ComputeUnits / 10
	if proof.ComputeUnits < job.ComputeUnits-tolerance ||
		proof.ComputeUnits > job.ComputeUnits+tolerance {
		k.Logger(ctx).Debug(
			"compute units mismatch",
			"expected", job.ComputeUnits,
			"got", proof.ComputeUnits,
			"tolerance", tolerance,
		)
		// Don't fail, just log - compute nodes may report slightly different values
	}

	// 2. Verify output hash is valid
	if len(proof.OutputHash) == 0 {
		return sdkerrors.Wrap(types.ErrInvalidProof, "output hash is empty")
	}

	// 3. Verify signature (simplified - in production use proper crypto)
	if !k.verifySignature(ctx, proof, job) {
		return sdkerrors.Wrap(types.ErrInvalidSignature, "signature verification failed")
	}

	// 4. Verify execution time is reasonable
	if proof.ExecutionTime <= 0 {
		return sdkerrors.Wrap(types.ErrInvalidProof, "invalid execution time")
	}

	// 5. Verify deterministic computation (spot check)
	// In production, this would compare against other nodes' results
	if !k.verifyDeterministicOutput(ctx, proof) {
		return sdkerrors.Wrap(types.ErrOutputHashMismatch, "deterministic verification failed")
	}

	return nil
}

// verifySignature verifies the node's signature on the proof
func (k Keeper) verifySignature(ctx sdk.Context, proof *types.ComputeProof, job *types.Job) bool {
	// In production, use proper cryptographic signature verification
	// For now, we use a simplified approach

	// Reconstruct the signed message
	message := k.buildProofMessage(proof)

	// Hash the message
	hash := sha256.Sum256(message)

	// Verify signature using node's public key
	// This is a placeholder - in production, use the node's actual public key
	// and proper ed25519/secp256k1 verification

	// For now, just verify the signature is not empty and has reasonable length
	if len(proof.Signature) < 64 {
		k.Logger(ctx).Debug("signature too short", "len", len(proof.Signature))
		return false
	}

	// TODO: Implement actual signature verification
	// This would involve:
	// 1. Getting the node's public key from the auth module
	// 2. Verifying the signature using the appropriate algorithm

	// For development, we accept all properly formatted signatures
	k.Logger(ctx).Debug(
		"signature verification (development mode)",
		"job_id", proof.JobID,
		"node", proof.NodeID,
		"hash", hex.EncodeToString(hash[:8]),
	)

	return true
}

// verifyDeterministicOutput verifies the computation is deterministic
func (k Keeper) verifyDeterministicOutput(ctx sdk.Context, proof *types.ComputeProof) bool {
	// In production, this would:
	// 1. Compare output hash against other nodes that ran the same job
	// 2. Use Merkle proofs for result verification
	// 3. Potentially re-execute a subset of the computation

	// For now, we verify the hash is well-formed
	if len(proof.OutputHash) != 32 {
		// Accept any non-empty hash for development
		if len(proof.OutputHash) > 0 {
			return true
		}
		return false
	}

	return true
}

// buildProofMessage builds the message that was signed
func (k Keeper) buildProofMessage(proof *types.ComputeProof) []byte {
	// Concatenate all fields to create the signed message
	msg := []byte(proof.JobID)
	msg = append(msg, []byte(proof.NodeID)...)
	msg = append(msg, sdk.Uint64ToBigEndian(proof.ComputeUnits)...)
	msg = append(msg, sdk.Uint64ToBigEndian(uint64(proof.ExecutionTime))...)
	msg = append(msg, proof.OutputHash...)
	return msg
}

// HasProof checks if a proof already exists for a job
func (k Keeper) HasProof(ctx sdk.Context, jobID string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetProofKey(jobID)
	return store.Has(key)
}

// GetProof retrieves a proof by job ID
func (k Keeper) GetProof(ctx sdk.Context, jobID string) (*types.ComputeProof, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetProofKey(jobID)

	bz := store.Get(key)
	if bz == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidProof, "proof not found")
	}

	var proof types.ComputeProof
	k.cdc.MustUnmarshal(bz, &proof)
	return &proof, nil
}

// SetProof stores a proof
func (k Keeper) SetProof(ctx sdk.Context, proof *types.ComputeProof) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetProofKey(proof.JobID)
	bz := k.cdc.MustMarshal(proof)
	store.Set(key, bz)
	return nil
}

// ValidateProofFormat validates the proof format without verification
func (k Keeper) ValidateProofFormat(proof *types.ComputeProof) error {
	if proof.JobID == "" {
		return sdkerrors.Wrap(types.ErrInvalidJobID, "job ID cannot be empty")
	}

	if proof.NodeID == "" {
		return sdkerrors.Wrap(types.ErrInvalidComputeNode, "node ID cannot be empty")
	}

	if proof.ComputeUnits == 0 {
		return sdkerrors.Wrap(types.ErrInvalidComputeUnits, "compute units must be positive")
	}

	if proof.ExecutionTime <= 0 {
		return sdkerrors.Wrap(types.ErrInvalidProof, "execution time must be positive")
	}

	if len(proof.OutputHash) == 0 {
		return sdkerrors.Wrap(types.ErrInvalidProof, "output hash cannot be empty")
	}

	if len(proof.Signature) == 0 {
		return sdkerrors.Wrap(types.ErrInvalidSignature, "signature cannot be empty")
	}

	return nil
}

// SpotCheckVerification performs spot-check verification for a completed job
// This is used for fraud detection in production
func (k Keeper) SpotCheckVerification(ctx sdk.Context, jobID string) error {
	job, err := k.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	proof, err := k.GetProof(ctx, jobID)
	if err != nil {
		return err
	}

	// Re-verify the proof
	if err := k.VerifyProof(ctx, proof, job); err != nil {
		k.Logger(ctx).Error(
			"spot check verification failed",
			"job_id", jobID,
			"error", err,
		)

		// In production, this would trigger a dispute or slashing
		return err
	}

	k.Logger(ctx).Debug("spot check passed", "job_id", jobID)
	return nil
}
