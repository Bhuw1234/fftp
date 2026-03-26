package proofofcompute

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/deparrow/dpc/x/proofofcompute/keeper"
	"github.com/deparrow/dpc/x/proofofcompute/types"
)

// NewHandler returns a handler for "proofofcompute" type messages.
func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgSubmitJob:
			res, err := msgServer.SubmitJob(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgSubmitProof:
			res, err := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgCancelJob:
			res, err := msgServer.CancelJob(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgClaimReward:
			res, err := msgServer.ClaimReward(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}

// msgServerImpl implements the types.MsgServer interface
type msgServerImpl struct {
	k keeper.Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper keeper.Keeper) types.MsgServer {
	return &msgServerImpl{keeper}
}

var _ types.MsgServer = msgServerImpl{}

// SubmitJob submits a new compute job
func (m msgServerImpl) SubmitJob(ctx context.Context, msg *types.MsgSubmitJob) (*types.MsgSubmitJobResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	job, err := m.k.SubmitJob(sdkCtx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitJobResponse{
		JobId:  job.ID,
		Status: string(job.Status),
	}, nil
}

// SubmitProof submits a compute proof for verification
func (m msgServerImpl) SubmitProof(ctx context.Context, msg *types.MsgSubmitProof) (*types.MsgSubmitProofResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := m.k.SubmitProof(sdkCtx, msg)
	if err != nil {
		return nil, err
	}

	// Get the completed job
	job, err := m.k.GetJob(sdkCtx, msg.JobID)
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitProofResponse{
		JobId:  msg.JobID,
		Status: string(job.Status),
		Reward: job.Reward.String(),
	}, nil
}

// CancelJob cancels a pending job
func (m msgServerImpl) CancelJob(ctx context.Context, msg *types.MsgCancelJob) (*types.MsgCancelJobResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := m.k.CancelJob(sdkCtx, msg.JobID, msg.Submitter)
	if err != nil {
		return nil, err
	}

	return &types.MsgCancelJobResponse{
		JobId:  msg.JobID,
		Status: "cancelled",
	}, nil
}

// ClaimReward claims pending rewards
func (m msgServerImpl) ClaimReward(ctx context.Context, msg *types.MsgClaimReward) (*types.MsgClaimRewardResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	reward, err := m.k.ClaimReward(sdkCtx, msg.NodeAddress)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimRewardResponse{
		Amount: reward.String(),
	}, nil
}