package agentwallet

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/deparrow/dpc/x/agentwallet/keeper"
	"github.com/deparrow/dpc/x/agentwallet/types"
)

// NewHandler returns a handler for "agentwallet" type messages.
func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgCreateWallet:
			res, err := msgServer.CreateWallet(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgDeleteWallet:
			res, err := msgServer.DeleteWallet(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgDeposit:
			res, err := msgServer.Deposit(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgWithdraw:
			res, err := msgServer.Withdraw(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgTransfer:
			res, err := msgServer.Transfer(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgAutonomousSpend:
			res, err := msgServer.AutonomousSpend(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgAddSpendingRule:
			res, err := msgServer.AddSpendingRule(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgUpdateSpendingRule:
			res, err := msgServer.UpdateSpendingRule(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgRemoveSpendingRule:
			res, err := msgServer.RemoveSpendingRule(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgAddAutomationRule:
			res, err := msgServer.AddAutomationRule(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgUpdateAutomationRule:
			res, err := msgServer.UpdateAutomationRule(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgRemoveAutomationRule:
			res, err := msgServer.RemoveAutomationRule(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgSetEmergencyReserve:
			res, err := msgServer.SetEmergencyReserve(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgRegisterAgent:
			res, err := msgServer.RegisterAgent(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgUnregisterAgent:
			res, err := msgServer.UnregisterAgent(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgUpdateParams:
			res, err := msgServer.UpdateParams(sdk.WrapSDKContext(ctx), msg)
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

// CreateWallet creates a new agent wallet
func (m msgServerImpl) CreateWallet(ctx context.Context, msg *types.MsgCreateWallet) (*types.MsgCreateWalletResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	wallet, err := m.k.CreateWallet(sdkCtx, msg.DID, msg.Address, msg.InitialFunds)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateWalletResponse{
		DID:     wallet.DID,
		Address: wallet.Address,
	}, nil
}

// DeleteWallet deletes an agent wallet
func (m msgServerImpl) DeleteWallet(ctx context.Context, msg *types.MsgDeleteWallet) (*types.MsgDeleteWalletResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	amount, err := m.k.DeleteWallet(sdkCtx, msg.Address)
	if err != nil {
		return nil, err
	}

	return &types.MsgDeleteWalletResponse{
		WithdrawnAmount: amount.String(),
	}, nil
}

// Deposit deposits funds into a wallet
func (m msgServerImpl) Deposit(ctx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	wallet, err := m.k.Deposit(sdkCtx, msg.Address, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgDepositResponse{
		Balance: wallet.Balance.String(),
	}, nil
}

// Withdraw withdraws funds from a wallet
func (m msgServerImpl) Withdraw(ctx context.Context, msg *types.MsgWithdraw) (*types.MsgWithdrawResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	wallet, err := m.k.Withdraw(sdkCtx, msg.Address, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgWithdrawResponse{
		Balance: wallet.Balance.String(),
	}, nil
}

// Transfer transfers funds between wallets
func (m msgServerImpl) Transfer(ctx context.Context, msg *types.MsgTransfer) (*types.MsgTransferResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	operation := msg.Operation
	if operation == "" {
		operation = types.OperationTransfer
	}

	sender, recipient, err := m.k.Transfer(sdkCtx, msg.Sender, msg.Recipient, msg.Amount, operation)
	if err != nil {
		return nil, err
	}

	return &types.MsgTransferResponse{
		SenderBalance:    sender.Balance.String(),
		RecipientBalance: recipient.Balance.String(),
	}, nil
}

// AutonomousSpend performs an autonomous spend operation
func (m msgServerImpl) AutonomousSpend(ctx context.Context, msg *types.MsgAutonomousSpend) (*types.MsgAutonomousSpendResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	wallet, txHash, err := m.k.AutonomousSpend(sdkCtx, msg.DID, msg.Address, msg.Recipient, msg.Amount, msg.Operation, msg.Signature)
	if err != nil {
		return nil, err
	}

	return &types.MsgAutonomousSpendResponse{
		Success: true,
		Balance: wallet.Balance.String(),
		TxHash:  txHash,
	}, nil
}

// AddSpendingRule adds a spending rule to a wallet
func (m msgServerImpl) AddSpendingRule(ctx context.Context, msg *types.MsgAddSpendingRule) (*types.MsgAddSpendingRuleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ruleIndex, err := m.k.AddSpendingRule(sdkCtx, msg.Address, msg.Rule)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddSpendingRuleResponse{
		RuleIndex: ruleIndex,
	}, nil
}

// UpdateSpendingRule updates a spending rule
func (m msgServerImpl) UpdateSpendingRule(ctx context.Context, msg *types.MsgUpdateSpendingRule) (*types.MsgUpdateSpendingRuleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := m.k.UpdateSpendingRule(sdkCtx, msg.Address, msg.RuleIndex, msg.Rule)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateSpendingRuleResponse{
		Success: true,
	}, nil
}

// RemoveSpendingRule removes a spending rule
func (m msgServerImpl) RemoveSpendingRule(ctx context.Context, msg *types.MsgRemoveSpendingRule) (*types.MsgRemoveSpendingRuleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := m.k.RemoveSpendingRule(sdkCtx, msg.Address, msg.RuleIndex)
	if err != nil {
		return nil, err
	}

	return &types.MsgRemoveSpendingRuleResponse{
		Success: true,
	}, nil
}

// AddAutomationRule adds an automation rule to a wallet
func (m msgServerImpl) AddAutomationRule(ctx context.Context, msg *types.MsgAddAutomationRule) (*types.MsgAddAutomationRuleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ruleIndex, err := m.k.AddAutomationRule(sdkCtx, msg.Address, msg.Rule)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddAutomationRuleResponse{
		RuleIndex: ruleIndex,
	}, nil
}

// UpdateAutomationRule updates an automation rule
func (m msgServerImpl) UpdateAutomationRule(ctx context.Context, msg *types.MsgUpdateAutomationRule) (*types.MsgUpdateAutomationRuleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := m.k.UpdateAutomationRule(sdkCtx, msg.Address, msg.RuleIndex, msg.Rule)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateAutomationRuleResponse{
		Success: true,
	}, nil
}

// RemoveAutomationRule removes an automation rule
func (m msgServerImpl) RemoveAutomationRule(ctx context.Context, msg *types.MsgRemoveAutomationRule) (*types.MsgRemoveAutomationRuleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := m.k.RemoveAutomationRule(sdkCtx, msg.Address, msg.RuleIndex)
	if err != nil {
		return nil, err
	}

	return &types.MsgRemoveAutomationRuleResponse{
		Success: true,
	}, nil
}

// SetEmergencyReserve sets the emergency reserve for a wallet
func (m msgServerImpl) SetEmergencyReserve(ctx context.Context, msg *types.MsgSetEmergencyReserve) (*types.MsgSetEmergencyReserveResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	wallet, err := m.k.SetEmergencyReserve(sdkCtx, msg.Address, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetEmergencyReserveResponse{
		Reserve: wallet.EmergencyReserve.String(),
	}, nil
}

// RegisterAgent registers a new AI agent
func (m msgServerImpl) RegisterAgent(ctx context.Context, msg *types.MsgRegisterAgent) (*types.MsgRegisterAgentResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	agent, err := m.k.RegisterAgent(sdkCtx, msg.DID, msg.Address, msg.AgentType, msg.Metadata)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterAgentResponse{
		DID:     agent.DID,
		Address: agent.Address,
	}, nil
}

// UnregisterAgent unregisters an AI agent
func (m msgServerImpl) UnregisterAgent(ctx context.Context, msg *types.MsgUnregisterAgent) (*types.MsgUnregisterAgentResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := m.k.UnregisterAgent(sdkCtx, msg.DID)
	if err != nil {
		return nil, err
	}

	return &types.MsgUnregisterAgentResponse{
		Success: true,
	}, nil
}

// UpdateParams updates module parameters
func (m msgServerImpl) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check authority
	if msg.Authority != m.k.GetAuthority() {
		return nil, sdkerrors.ErrUnauthorized
	}

	err := m.k.SetParams(sdkCtx, msg.Params)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{
		Success: true,
	}, nil
}
