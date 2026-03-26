package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers the necessary x/agentwallet interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateWallet{}, "dpc/CreateWallet", nil)
	cdc.RegisterConcrete(&MsgDeleteWallet{}, "dpc/DeleteWallet", nil)
	cdc.RegisterConcrete(&MsgDeposit{}, "dpc/Deposit", nil)
	cdc.RegisterConcrete(&MsgWithdraw{}, "dpc/Withdraw", nil)
	cdc.RegisterConcrete(&MsgTransfer{}, "dpc/Transfer", nil)
	cdc.RegisterConcrete(&MsgAutonomousSpend{}, "dpc/AutonomousSpend", nil)
	cdc.RegisterConcrete(&MsgAddSpendingRule{}, "dpc/AddSpendingRule", nil)
	cdc.RegisterConcrete(&MsgUpdateSpendingRule{}, "dpc/UpdateSpendingRule", nil)
	cdc.RegisterConcrete(&MsgRemoveSpendingRule{}, "dpc/RemoveSpendingRule", nil)
	cdc.RegisterConcrete(&MsgAddAutomationRule{}, "dpc/AddAutomationRule", nil)
	cdc.RegisterConcrete(&MsgUpdateAutomationRule{}, "dpc/UpdateAutomationRule", nil)
	cdc.RegisterConcrete(&MsgRemoveAutomationRule{}, "dpc/RemoveAutomationRule", nil)
	cdc.RegisterConcrete(&MsgSetEmergencyReserve{}, "dpc/SetEmergencyReserve", nil)
	cdc.RegisterConcrete(&MsgRegisterAgent{}, "dpc/RegisterAgent", nil)
	cdc.RegisterConcrete(&MsgUnregisterAgent{}, "dpc/UnregisterAgent", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "dpc/UpdateParams", nil)
}

// RegisterInterfaces registers the x/agentwallet interfaces types with the interface registry
// Note: In production, proto-generated types would implement proto.Message
// For now, we skip interface registration as manual types don't implement proto.Message
func RegisterInterfaces(registry interface{}) {
	// Placeholder for proto-generated types registration
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/agentwallet module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)
}
