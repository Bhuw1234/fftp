package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers the necessary x/computemarket interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterProvider{}, "dpc/RegisterProvider", nil)
	cdc.RegisterConcrete(&MsgUnregisterProvider{}, "dpc/UnregisterProvider", nil)
	cdc.RegisterConcrete(&MsgStakeProvider{}, "dpc/StakeProvider", nil)
	cdc.RegisterConcrete(&MsgUnstakeProvider{}, "dpc/UnstakeProvider", nil)
	cdc.RegisterConcrete(&MsgCreateEscrow{}, "dpc/CreateEscrow", nil)
	cdc.RegisterConcrete(&MsgReleaseEscrow{}, "dpc/ReleaseEscrow", nil)
	cdc.RegisterConcrete(&MsgRefundEscrow{}, "dpc/RefundEscrow", nil)
	cdc.RegisterConcrete(&MsgOpenDispute{}, "dpc/OpenDispute", nil)
	cdc.RegisterConcrete(&MsgResolveDispute{}, "dpc/ResolveDispute", nil)
}

// RegisterInterfaces registers the x/computemarket interfaces types with the interface registry
// Note: In production, proto-generated types would implement proto.Message
// For now, we skip interface registration as manual types don't implement proto.Message
func RegisterInterfaces(registry interface{}) {
	// Placeholder for proto-generated types registration
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/computemarket module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)
}
