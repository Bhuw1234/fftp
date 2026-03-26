package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "cosmossdk.io/errors"
)

// MsgRegisterProvider defines a message to register as a compute provider
type MsgRegisterProvider struct {
	Address      string              `json:"address"`
	Stake        sdk.Coin            `json:"stake"`
	Capabilities ProviderCapabilities `json:"capabilities"`
}

// Route implements sdk.Msg
func (m MsgRegisterProvider) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgRegisterProvider) Type() string {
	return "register_provider"
}

// GetSigners implements sdk.Msg
func (m MsgRegisterProvider) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgRegisterProvider) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgRegisterProvider) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}

	if !m.Stake.IsValid() {
		return sdkerrors.Wrap(ErrInvalidStake, "invalid stake coin")
	}

	return nil
}

// MsgUnregisterProvider defines a message to unregister as a compute provider
type MsgUnregisterProvider struct {
	Address string `json:"address"`
}

// Route implements sdk.Msg
func (m MsgUnregisterProvider) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgUnregisterProvider) Type() string {
	return "unregister_provider"
}

// GetSigners implements sdk.Msg
func (m MsgUnregisterProvider) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgUnregisterProvider) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgUnregisterProvider) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}

	return nil
}

// MsgStakeProvider defines a message to add more stake to a provider
type MsgStakeProvider struct {
	Address string   `json:"address"`
	Amount  sdk.Coin `json:"amount"`
}

// Route implements sdk.Msg
func (m MsgStakeProvider) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgStakeProvider) Type() string {
	return "stake_provider"
}

// GetSigners implements sdk.Msg
func (m MsgStakeProvider) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgStakeProvider) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgStakeProvider) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}

	if !m.Amount.IsValid() {
		return sdkerrors.Wrap(ErrInvalidStake, "invalid stake coin")
	}

	return nil
}

// MsgUnstakeProvider defines a message to unstake from a provider
type MsgUnstakeProvider struct {
	Address string   `json:"address"`
	Amount  sdk.Coin `json:"amount"`
}

// Route implements sdk.Msg
func (m MsgUnstakeProvider) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgUnstakeProvider) Type() string {
	return "unstake_provider"
}

// GetSigners implements sdk.Msg
func (m MsgUnstakeProvider) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgUnstakeProvider) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgUnstakeProvider) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}

	if !m.Amount.IsValid() {
		return sdkerrors.Wrap(ErrInvalidStake, "invalid stake coin")
	}

	return nil
}

// MsgCreateEscrow defines a message to create an escrow for a job
type MsgCreateEscrow struct {
	JobID     string   `json:"job_id"`
	Submitter string   `json:"submitter"`
	Provider  string   `json:"provider"`
	Amount    sdk.Coin `json:"amount"`
	Duration  int64    `json:"duration"` // Duration in seconds
}

// Route implements sdk.Msg
func (m MsgCreateEscrow) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgCreateEscrow) Type() string {
	return "create_escrow"
}

// GetSigners implements sdk.Msg
func (m MsgCreateEscrow) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Submitter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgCreateEscrow) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgCreateEscrow) ValidateBasic() error {
	if m.JobID == "" {
		return sdkerrors.Wrap(ErrInvalidJobID, "job ID cannot be empty")
	}

	if m.Submitter == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "submitter cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(m.Submitter)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, "invalid submitter address")
	}

	if m.Provider == "" {
		return sdkerrors.Wrap(ErrInvalidProvider, "provider cannot be empty")
	}

	_, err = sdk.AccAddressFromBech32(m.Provider)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidProvider, "invalid provider address")
	}

	if !m.Amount.IsValid() {
		return sdkerrors.Wrap(ErrInsufficientFunds, "invalid amount")
	}

	if m.Duration <= 0 {
		return sdkerrors.Wrap(ErrInvalidParams, "duration must be positive")
	}

	return nil
}

// MsgReleaseEscrow defines a message to release escrow to provider
type MsgReleaseEscrow struct {
	EscrowID  string `json:"escrow_id"`
	Submitter string `json:"submitter"`
}

// Route implements sdk.Msg
func (m MsgReleaseEscrow) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgReleaseEscrow) Type() string {
	return "release_escrow"
}

// GetSigners implements sdk.Msg
func (m MsgReleaseEscrow) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Submitter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgReleaseEscrow) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgReleaseEscrow) ValidateBasic() error {
	if m.EscrowID == "" {
		return sdkerrors.Wrap(ErrEscrowNotFound, "escrow ID cannot be empty")
	}

	if m.Submitter == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "submitter cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(m.Submitter)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}

	return nil
}

// MsgRefundEscrow defines a message to refund escrow to submitter
type MsgRefundEscrow struct {
	EscrowID string `json:"escrow_id"`
	Caller   string `json:"caller"`
}

// Route implements sdk.Msg
func (m MsgRefundEscrow) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgRefundEscrow) Type() string {
	return "refund_escrow"
}

// GetSigners implements sdk.Msg
func (m MsgRefundEscrow) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Caller)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgRefundEscrow) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgRefundEscrow) ValidateBasic() error {
	if m.EscrowID == "" {
		return sdkerrors.Wrap(ErrEscrowNotFound, "escrow ID cannot be empty")
	}

	if m.Caller == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "caller cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(m.Caller)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}

	return nil
}

// MsgOpenDispute defines a message to open a dispute
type MsgOpenDispute struct {
	EscrowID string `json:"escrow_id"`
	Disputer string `json:"disputer"`
	Reason   string `json:"reason"`
	Evidence []byte `json:"evidence"`
}

// Route implements sdk.Msg
func (m MsgOpenDispute) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgOpenDispute) Type() string {
	return "open_dispute"
}

// GetSigners implements sdk.Msg
func (m MsgOpenDispute) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Disputer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgOpenDispute) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgOpenDispute) ValidateBasic() error {
	if m.EscrowID == "" {
		return sdkerrors.Wrap(ErrEscrowNotFound, "escrow ID cannot be empty")
	}

	if m.Disputer == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "disputer cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(m.Disputer)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}

	if m.Reason == "" {
		return sdkerrors.Wrap(ErrInvalidDisputeReason, "reason cannot be empty")
	}

	return nil
}

// MsgResolveDispute defines a message to resolve a dispute
type MsgResolveDispute struct {
	DisputeID  string `json:"dispute_id"`
	Resolver   string `json:"resolver"`
	Resolution string `json:"resolution"`
	Winner     string `json:"winner"`
}

// Route implements sdk.Msg
func (m MsgResolveDispute) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgResolveDispute) Type() string {
	return "resolve_dispute"
}

// GetSigners implements sdk.Msg
func (m MsgResolveDispute) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Resolver)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgResolveDispute) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgResolveDispute) ValidateBasic() error {
	if m.DisputeID == "" {
		return sdkerrors.Wrap(ErrDisputeNotFound, "dispute ID cannot be empty")
	}

	if m.Resolver == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "resolver cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(m.Resolver)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}

	if m.Resolution == "" {
		return sdkerrors.Wrap(ErrInvalidParams, "resolution cannot be empty")
	}

	return nil
}