package computemarket

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/deparrow/dpc/x/computemarket/keeper"
	"github.com/deparrow/dpc/x/computemarket/types"
)

// NewHandler returns a handler for "computemarket" type messages.
func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgRegisterProvider:
			res, err := msgServer.RegisterProvider(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgUnregisterProvider:
			res, err := msgServer.UnregisterProvider(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgStakeProvider:
			res, err := msgServer.StakeProvider(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgUnstakeProvider:
			res, err := msgServer.UnstakeProvider(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgCreateEscrow:
			res, err := msgServer.CreateEscrow(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgReleaseEscrow:
			res, err := msgServer.ReleaseEscrow(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgRefundEscrow:
			res, err := msgServer.RefundEscrow(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgOpenDispute:
			res, err := msgServer.OpenDispute(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgResolveDispute:
			res, err := msgServer.ResolveDispute(sdk.WrapSDKContext(ctx), msg)
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

// RegisterProvider registers a new compute provider
func (m msgServerImpl) RegisterProvider(ctx context.Context, msg *types.MsgRegisterProvider) (*types.MsgRegisterProviderResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	provider, err := m.k.RegisterProvider(sdkCtx, msg.Address, msg.Stake, msg.Capabilities)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterProviderResponse{
		Address:         provider.Address,
		ReputationScore: provider.ReputationScore,
	}, nil
}

// UnregisterProvider unregisters a compute provider
func (m msgServerImpl) UnregisterProvider(ctx context.Context, msg *types.MsgUnregisterProvider) (*types.MsgUnregisterProviderResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := m.k.UnregisterProvider(sdkCtx, msg.Address)
	if err != nil {
		return nil, err
	}

	return &types.MsgUnregisterProviderResponse{}, nil
}

// StakeProvider adds more stake to a provider
func (m msgServerImpl) StakeProvider(ctx context.Context, msg *types.MsgStakeProvider) (*types.MsgStakeProviderResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	provider, err := m.k.StakeProvider(sdkCtx, msg.Address, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgStakeProviderResponse{
		Address:      provider.Address,
		TotalStake:   provider.StakedAmount.String(),
	}, nil
}

// UnstakeProvider removes stake from a provider
func (m msgServerImpl) UnstakeProvider(ctx context.Context, msg *types.MsgUnstakeProvider) (*types.MsgUnstakeProviderResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	provider, err := m.k.UnstakeProvider(sdkCtx, msg.Address, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgUnstakeProviderResponse{
		Address:      provider.Address,
		RemainingStake: provider.StakedAmount.String(),
	}, nil
}

// CreateEscrow creates a new escrow for a job
func (m msgServerImpl) CreateEscrow(ctx context.Context, msg *types.MsgCreateEscrow) (*types.MsgCreateEscrowResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	escrow, err := m.k.CreateEscrow(sdkCtx, msg.JobID, msg.Submitter, msg.Provider, msg.Amount, msg.Duration)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateEscrowResponse{
		EscrowId: escrow.ID,
		Status:   escrow.Status.String(),
	}, nil
}

// ReleaseEscrow releases escrow funds to the provider
func (m msgServerImpl) ReleaseEscrow(ctx context.Context, msg *types.MsgReleaseEscrow) (*types.MsgReleaseEscrowResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := m.k.ReleaseEscrow(sdkCtx, msg.EscrowID, msg.Submitter)
	if err != nil {
		return nil, err
	}

	return &types.MsgReleaseEscrowResponse{
		EscrowId: msg.EscrowID,
		Status:   "released",
	}, nil
}

// RefundEscrow refunds escrow funds to the submitter
func (m msgServerImpl) RefundEscrow(ctx context.Context, msg *types.MsgRefundEscrow) (*types.MsgRefundEscrowResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := m.k.RefundEscrow(sdkCtx, msg.EscrowID, msg.Caller)
	if err != nil {
		return nil, err
	}

	return &types.MsgRefundEscrowResponse{
		EscrowId: msg.EscrowID,
		Status:   "refunded",
	}, nil
}

// OpenDispute opens a dispute for an escrow
func (m msgServerImpl) OpenDispute(ctx context.Context, msg *types.MsgOpenDispute) (*types.MsgOpenDisputeResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	dispute, err := m.k.OpenDispute(sdkCtx, msg.EscrowID, msg.Disputer, msg.Reason, msg.Evidence)
	if err != nil {
		return nil, err
	}

	return &types.MsgOpenDisputeResponse{
		DisputeId: dispute.ID,
		Status:    dispute.Status.String(),
	}, nil
}

// ResolveDispute resolves a dispute
func (m msgServerImpl) ResolveDispute(ctx context.Context, msg *types.MsgResolveDispute) (*types.MsgResolveDisputeResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := m.k.ResolveDispute(sdkCtx, msg.DisputeID, msg.Resolver, msg.Resolution, msg.Winner)
	if err != nil {
		return nil, err
	}

	return &types.MsgResolveDisputeResponse{
		DisputeId: msg.DisputeID,
		Resolution: msg.Resolution,
	}, nil
}
