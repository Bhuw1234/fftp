package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "cosmossdk.io/errors"
)

// MsgCreateWallet defines a message to create a new agent wallet
type MsgCreateWallet struct {
	DID        string   `json:"did"`
	Address    string   `json:"address"`
	InitialFunds sdk.Coin `json:"initial_funds"`
}

// Route implements sdk.Msg
func (m MsgCreateWallet) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgCreateWallet) Type() string {
	return "create_wallet"
}

// GetSigners implements sdk.Msg
func (m MsgCreateWallet) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgCreateWallet) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgCreateWallet) ValidateBasic() error {
	if m.DID == "" {
		return sdkerrors.Wrap(ErrInvalidDID, "DID cannot be empty")
	}
	if !IsValidDID(m.DID) {
		return sdkerrors.Wrap(ErrInvalidDID, "invalid DID format")
	}
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	if m.InitialFunds.IsNegative() {
		return sdkerrors.Wrap(ErrInvalidAmount, "initial funds cannot be negative")
	}
	return nil
}

// MsgDeleteWallet defines a message to delete an agent wallet
type MsgDeleteWallet struct {
	Address string `json:"address"`
}

// Route implements sdk.Msg
func (m MsgDeleteWallet) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgDeleteWallet) Type() string {
	return "delete_wallet"
}

// GetSigners implements sdk.Msg
func (m MsgDeleteWallet) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgDeleteWallet) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgDeleteWallet) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	return nil
}

// MsgDeposit defines a message to deposit funds into a wallet
type MsgDeposit struct {
	Address string   `json:"address"`
	Amount  sdk.Coin `json:"amount"`
}

// Route implements sdk.Msg
func (m MsgDeposit) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgDeposit) Type() string {
	return "deposit"
}

// GetSigners implements sdk.Msg
func (m MsgDeposit) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgDeposit) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgDeposit) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	if !m.Amount.IsValid() || m.Amount.IsZero() {
		return sdkerrors.Wrap(ErrInvalidAmount, "invalid deposit amount")
	}
	return nil
}

// MsgWithdraw defines a message to withdraw funds from a wallet
type MsgWithdraw struct {
	Address string   `json:"address"`
	Amount  sdk.Coin `json:"amount"`
}

// Route implements sdk.Msg
func (m MsgWithdraw) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgWithdraw) Type() string {
	return "withdraw"
}

// GetSigners implements sdk.Msg
func (m MsgWithdraw) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgWithdraw) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgWithdraw) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	if !m.Amount.IsValid() || m.Amount.IsZero() {
		return sdkerrors.Wrap(ErrInvalidAmount, "invalid withdrawal amount")
	}
	return nil
}

// MsgTransfer defines a message to transfer funds between wallets
type MsgTransfer struct {
	Sender    string   `json:"sender"`
	Recipient string   `json:"recipient"`
	Amount    sdk.Coin `json:"amount"`
	Operation string   `json:"operation"` // For spending rule check
}

// Route implements sdk.Msg
func (m MsgTransfer) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgTransfer) Type() string {
	return "transfer"
}

// GetSigners implements sdk.Msg
func (m MsgTransfer) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgTransfer) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgTransfer) ValidateBasic() error {
	if m.Sender == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "sender cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, "invalid sender address")
	}
	if m.Recipient == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "recipient cannot be empty")
	}
	_, err = sdk.AccAddressFromBech32(m.Recipient)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, "invalid recipient address")
	}
	if m.Sender == m.Recipient {
		return sdkerrors.Wrap(ErrSelfTransfer, "cannot transfer to self")
	}
	if !m.Amount.IsValid() || m.Amount.IsZero() {
		return sdkerrors.Wrap(ErrInvalidAmount, "invalid transfer amount")
	}
	return nil
}

// MsgAutonomousSpend defines a message for autonomous spending by AI agents
type MsgAutonomousSpend struct {
	DID       string   `json:"did"`
	Address   string   `json:"address"`
	Recipient string   `json:"recipient"`
	Amount    sdk.Coin `json:"amount"`
	Operation string   `json:"operation"`
	Signature []byte   `json:"signature"` // Autonomous signature
}

// Route implements sdk.Msg
func (m MsgAutonomousSpend) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgAutonomousSpend) Type() string {
	return "autonomous_spend"
}

// GetSigners implements sdk.Msg
func (m MsgAutonomousSpend) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgAutonomousSpend) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgAutonomousSpend) ValidateBasic() error {
	if !IsValidDID(m.DID) {
		return sdkerrors.Wrap(ErrInvalidDID, "invalid DID format")
	}
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, "invalid address")
	}
	if m.Recipient == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "recipient cannot be empty")
	}
	if !m.Amount.IsValid() || m.Amount.IsZero() {
		return sdkerrors.Wrap(ErrInvalidAmount, "invalid spend amount")
	}
	if len(m.Signature) == 0 {
		return sdkerrors.Wrap(ErrInvalidSignature, "signature cannot be empty")
	}
	return nil
}

// MsgAddSpendingRule defines a message to add a spending rule
type MsgAddSpendingRule struct {
	Address string        `json:"address"`
	Rule    SpendingRule  `json:"rule"`
}

// Route implements sdk.Msg
func (m MsgAddSpendingRule) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgAddSpendingRule) Type() string {
	return "add_spending_rule"
}

// GetSigners implements sdk.Msg
func (m MsgAddSpendingRule) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgAddSpendingRule) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgAddSpendingRule) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	return m.Rule.Validate()
}

// MsgUpdateSpendingRule defines a message to update a spending rule
type MsgUpdateSpendingRule struct {
	Address   string       `json:"address"`
	RuleIndex uint32       `json:"rule_index"`
	Rule      SpendingRule `json:"rule"`
}

// Route implements sdk.Msg
func (m MsgUpdateSpendingRule) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgUpdateSpendingRule) Type() string {
	return "update_spending_rule"
}

// GetSigners implements sdk.Msg
func (m MsgUpdateSpendingRule) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgUpdateSpendingRule) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgUpdateSpendingRule) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	return m.Rule.Validate()
}

// MsgRemoveSpendingRule defines a message to remove a spending rule
type MsgRemoveSpendingRule struct {
	Address   string `json:"address"`
	RuleIndex uint32 `json:"rule_index"`
}

// Route implements sdk.Msg
func (m MsgRemoveSpendingRule) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgRemoveSpendingRule) Type() string {
	return "remove_spending_rule"
}

// GetSigners implements sdk.Msg
func (m MsgRemoveSpendingRule) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgRemoveSpendingRule) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgRemoveSpendingRule) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	return nil
}

// MsgAddAutomationRule defines a message to add an automation rule
type MsgAddAutomationRule struct {
	Address string          `json:"address"`
	Rule    AutomationRule  `json:"rule"`
}

// Route implements sdk.Msg
func (m MsgAddAutomationRule) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgAddAutomationRule) Type() string {
	return "add_automation_rule"
}

// GetSigners implements sdk.Msg
func (m MsgAddAutomationRule) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgAddAutomationRule) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgAddAutomationRule) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	return m.Rule.Validate()
}

// MsgUpdateAutomationRule defines a message to update an automation rule
type MsgUpdateAutomationRule struct {
	Address   string         `json:"address"`
	RuleIndex uint32         `json:"rule_index"`
	Rule      AutomationRule `json:"rule"`
}

// Route implements sdk.Msg
func (m MsgUpdateAutomationRule) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgUpdateAutomationRule) Type() string {
	return "update_automation_rule"
}

// GetSigners implements sdk.Msg
func (m MsgUpdateAutomationRule) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgUpdateAutomationRule) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgUpdateAutomationRule) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	return m.Rule.Validate()
}

// MsgRemoveAutomationRule defines a message to remove an automation rule
type MsgRemoveAutomationRule struct {
	Address   string `json:"address"`
	RuleIndex uint32 `json:"rule_index"`
}

// Route implements sdk.Msg
func (m MsgRemoveAutomationRule) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgRemoveAutomationRule) Type() string {
	return "remove_automation_rule"
}

// GetSigners implements sdk.Msg
func (m MsgRemoveAutomationRule) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgRemoveAutomationRule) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgRemoveAutomationRule) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	return nil
}

// MsgSetEmergencyReserve defines a message to set the emergency reserve
type MsgSetEmergencyReserve struct {
	Address string   `json:"address"`
	Amount  sdk.Coin `json:"amount"`
}

// Route implements sdk.Msg
func (m MsgSetEmergencyReserve) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgSetEmergencyReserve) Type() string {
	return "set_emergency_reserve"
}

// GetSigners implements sdk.Msg
func (m MsgSetEmergencyReserve) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgSetEmergencyReserve) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgSetEmergencyReserve) ValidateBasic() error {
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	if m.Amount.IsNegative() {
		return sdkerrors.Wrap(ErrInvalidReserveAmount, "reserve cannot be negative")
	}
	return nil
}

// MsgRegisterAgent defines a message to register an AI agent
type MsgRegisterAgent struct {
	DID       string `json:"did"`
	Address   string `json:"address"`
	AgentType string `json:"agent_type"`
	Metadata  string `json:"metadata"`
}

// Route implements sdk.Msg
func (m MsgRegisterAgent) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgRegisterAgent) Type() string {
	return "register_agent"
}

// GetSigners implements sdk.Msg
func (m MsgRegisterAgent) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgRegisterAgent) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgRegisterAgent) ValidateBasic() error {
	if !IsValidDID(m.DID) {
		return sdkerrors.Wrap(ErrInvalidDID, "invalid DID format")
	}
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	return nil
}

// MsgUnregisterAgent defines a message to unregister an AI agent
type MsgUnregisterAgent struct {
	DID     string `json:"did"`
	Address string `json:"address"`
}

// Route implements sdk.Msg
func (m MsgUnregisterAgent) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgUnregisterAgent) Type() string {
	return "unregister_agent"
}

// GetSigners implements sdk.Msg
func (m MsgUnregisterAgent) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgUnregisterAgent) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgUnregisterAgent) ValidateBasic() error {
	if !IsValidDID(m.DID) {
		return sdkerrors.Wrap(ErrInvalidDID, "invalid DID format")
	}
	if m.Address == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	return nil
}

// MsgUpdateParams defines a message to update module parameters
type MsgUpdateParams struct {
	Authority string `json:"authority"`
	Params    Params `json:"params"`
}

// Route implements sdk.Msg
func (m MsgUpdateParams) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgUpdateParams) Type() string {
	return "update_params"
}

// GetSigners implements sdk.Msg
func (m MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgUpdateParams) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgUpdateParams) ValidateBasic() error {
	if m.Authority == "" {
		return sdkerrors.Wrap(ErrInvalidAddress, "authority cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}
	return m.Params.Validate()
}
