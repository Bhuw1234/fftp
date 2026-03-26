package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgSubmitJob defines a message to submit a new compute job
type MsgSubmitJob struct {
	Submitter    string `json:"submitter"`
	JobSpec      []byte `json:"job_spec"`
	Stake        sdk.Coin `json:"stake"`
	ComputeUnits uint64 `json:"compute_units"`
}

// Route implements sdk.Msg
func (m MsgSubmitJob) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgSubmitJob) Type() string {
	return "submit_job"
}

// GetSigners implements sdk.Msg
func (m MsgSubmitJob) GetSigners() []sdk.AccAddress {
	submitter, err := sdk.AccAddressFromBech32(m.Submitter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{submitter}
}

// GetSignBytes implements sdk.Msg
func (m MsgSubmitJob) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgSubmitJob) ValidateBasic() error {
	if m.Submitter == "" {
		return sdkerrors.Wrap(ErrInvalidSubmitter, "submitter cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(m.Submitter)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidSubmitter, err.Error())
	}

	if len(m.JobSpec) == 0 {
		return sdkerrors.Wrap(ErrInvalidJobID, "job spec cannot be empty")
	}

	if !m.Stake.IsValid() {
		return sdkerrors.Wrap(ErrInvalidStake, "invalid stake coin")
	}

	if m.ComputeUnits == 0 {
		return sdkerrors.Wrap(ErrInvalidComputeUnits, "compute units must be positive")
	}

	return nil
}

// MsgSubmitProof defines a message to submit compute proof
type MsgSubmitProof struct {
	JobID         string `json:"job_id"`
	NodeAddress   string `json:"node_address"`
	ComputeUnits  uint64 `json:"compute_units"`
	ExecutionTime int64  `json:"execution_time"`
	OutputHash    []byte `json:"output_hash"`
	Signature     []byte `json:"signature"`
}

// Route implements sdk.Msg
func (m MsgSubmitProof) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgSubmitProof) Type() string {
	return "submit_proof"
}

// GetSigners implements sdk.Msg
func (m MsgSubmitProof) GetSigners() []sdk.AccAddress {
	nodeAddr, err := sdk.AccAddressFromBech32(m.NodeAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{nodeAddr}
}

// GetSignBytes implements sdk.Msg
func (m MsgSubmitProof) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgSubmitProof) ValidateBasic() error {
	if m.JobID == "" {
		return sdkerrors.Wrap(ErrInvalidJobID, "job ID cannot be empty")
	}

	if m.NodeAddress == "" {
		return sdkerrors.Wrap(ErrInvalidComputeNode, "node address cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(m.NodeAddress)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidComputeNode, err.Error())
	}

	if m.ComputeUnits == 0 {
		return sdkerrors.Wrap(ErrInvalidComputeUnits, "compute units must be positive")
	}

	if m.ExecutionTime <= 0 {
		return sdkerrors.Wrap(ErrInvalidProof, "execution time must be positive")
	}

	if len(m.OutputHash) == 0 {
		return sdkerrors.Wrap(ErrInvalidProof, "output hash cannot be empty")
	}

	if len(m.Signature) == 0 {
		return sdkerrors.Wrap(ErrInvalidSignature, "signature cannot be empty")
	}

	return nil
}

// MsgCancelJob defines a message to cancel a job
type MsgCancelJob struct {
	JobID     string `json:"job_id"`
	Submitter string `json:"submitter"`
}

// Route implements sdk.Msg
func (m MsgCancelJob) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgCancelJob) Type() string {
	return "cancel_job"
}

// GetSigners implements sdk.Msg
func (m MsgCancelJob) GetSigners() []sdk.AccAddress {
	submitter, err := sdk.AccAddressFromBech32(m.Submitter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{submitter}
}

// GetSignBytes implements sdk.Msg
func (m MsgCancelJob) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgCancelJob) ValidateBasic() error {
	if m.JobID == "" {
		return sdkerrors.Wrap(ErrInvalidJobID, "job ID cannot be empty")
	}

	if m.Submitter == "" {
		return sdkerrors.Wrap(ErrInvalidSubmitter, "submitter cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(m.Submitter)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidSubmitter, err.Error())
	}

	return nil
}

// MsgClaimReward defines a message to claim pending rewards
type MsgClaimReward struct {
	NodeAddress string `json:"node_address"`
}

// Route implements sdk.Msg
func (m MsgClaimReward) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgClaimReward) Type() string {
	return "claim_reward"
}

// GetSigners implements sdk.Msg
func (m MsgClaimReward) GetSigners() []sdk.AccAddress {
	nodeAddr, err := sdk.AccAddressFromBech32(m.NodeAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{nodeAddr}
}

// GetSignBytes implements sdk.Msg
func (m MsgClaimReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgClaimReward) ValidateBasic() error {
	if m.NodeAddress == "" {
		return sdkerrors.Wrap(ErrInvalidComputeNode, "node address cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(m.NodeAddress)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidComputeNode, err.Error())
	}

	return nil
}
