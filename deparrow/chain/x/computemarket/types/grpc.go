package types

import (
	context "context"
)

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// Provider management
	RegisterProvider(context.Context, *MsgRegisterProvider) (*MsgRegisterProviderResponse, error)
	UnregisterProvider(context.Context, *MsgUnregisterProvider) (*MsgUnregisterProviderResponse, error)
	StakeProvider(context.Context, *MsgStakeProvider) (*MsgStakeProviderResponse, error)
	UnstakeProvider(context.Context, *MsgUnstakeProvider) (*MsgUnstakeProviderResponse, error)
	
	// Escrow management
	CreateEscrow(context.Context, *MsgCreateEscrow) (*MsgCreateEscrowResponse, error)
	ReleaseEscrow(context.Context, *MsgReleaseEscrow) (*MsgReleaseEscrowResponse, error)
	RefundEscrow(context.Context, *MsgRefundEscrow) (*MsgRefundEscrowResponse, error)
	
	// Dispute management
	OpenDispute(context.Context, *MsgOpenDispute) (*MsgOpenDisputeResponse, error)
	ResolveDispute(context.Context, *MsgResolveDispute) (*MsgResolveDisputeResponse, error)
}

// Response types

// MsgRegisterProviderResponse is the response for MsgRegisterProvider
type MsgRegisterProviderResponse struct {
	Address         string `json:"address"`
	ReputationScore uint32 `json:"reputation_score"`
}

// MsgUnregisterProviderResponse is the response for MsgUnregisterProvider
type MsgUnregisterProviderResponse struct{}

// MsgStakeProviderResponse is the response for MsgStakeProvider
type MsgStakeProviderResponse struct {
	Address    string `json:"address"`
	TotalStake string `json:"total_stake"`
}

// MsgUnstakeProviderResponse is the response for MsgUnstakeProvider
type MsgUnstakeProviderResponse struct {
	Address        string `json:"address"`
	RemainingStake string `json:"remaining_stake"`
}

// MsgCreateEscrowResponse is the response for MsgCreateEscrow
type MsgCreateEscrowResponse struct {
	EscrowId string `json:"escrow_id"`
	Status   string `json:"status"`
}

// MsgReleaseEscrowResponse is the response for MsgReleaseEscrow
type MsgReleaseEscrowResponse struct {
	EscrowId string `json:"escrow_id"`
	Status   string `json:"status"`
}

// MsgRefundEscrowResponse is the response for MsgRefundEscrow
type MsgRefundEscrowResponse struct {
	EscrowId string `json:"escrow_id"`
	Status   string `json:"status"`
}

// MsgOpenDisputeResponse is the response for MsgOpenDispute
type MsgOpenDisputeResponse struct {
	DisputeId string `json:"dispute_id"`
	Status    string `json:"status"`
}

// MsgResolveDisputeResponse is the response for MsgResolveDispute
type MsgResolveDisputeResponse struct {
	DisputeId  string `json:"dispute_id"`
	Resolution string `json:"resolution"`
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct{}

func (UnimplementedMsgServer) RegisterProvider(ctx context.Context, req *MsgRegisterProvider) (*MsgRegisterProviderResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) UnregisterProvider(ctx context.Context, req *MsgUnregisterProvider) (*MsgUnregisterProviderResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) StakeProvider(ctx context.Context, req *MsgStakeProvider) (*MsgStakeProviderResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) UnstakeProvider(ctx context.Context, req *MsgUnstakeProvider) (*MsgUnstakeProviderResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) CreateEscrow(ctx context.Context, req *MsgCreateEscrow) (*MsgCreateEscrowResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) ReleaseEscrow(ctx context.Context, req *MsgReleaseEscrow) (*MsgReleaseEscrowResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) RefundEscrow(ctx context.Context, req *MsgRefundEscrow) (*MsgRefundEscrowResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) OpenDispute(ctx context.Context, req *MsgOpenDispute) (*MsgOpenDisputeResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) ResolveDispute(ctx context.Context, req *MsgResolveDispute) (*MsgResolveDisputeResponse, error) {
	return nil, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// Provider queries
	Provider(context.Context, *QueryProviderRequest) (*QueryProviderResponse, error)
	Providers(context.Context, *QueryProvidersRequest) (*QueryProvidersResponse, error)
	ActiveProviders(context.Context, *QueryActiveProvidersRequest) (*QueryActiveProvidersResponse, error)
	
	// Escrow queries
	Escrow(context.Context, *QueryEscrowRequest) (*QueryEscrowResponse, error)
	EscrowsByProvider(context.Context, *QueryEscrowsByProviderRequest) (*QueryEscrowsByProviderResponse, error)
	EscrowByJob(context.Context, *QueryEscrowByJobRequest) (*QueryEscrowResponse, error)
	
	// Dispute queries
	Dispute(context.Context, *QueryDisputeRequest) (*QueryDisputeResponse, error)
	DisputesByProvider(context.Context, *QueryDisputesByProviderRequest) (*QueryDisputesByProviderResponse, error)
	
	// Stats queries
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	Stats(context.Context, *QueryStatsRequest) (*QueryStatsResponse, error)
}

// Query request types

type QueryProviderRequest struct {
	Address string `json:"address"`
}

type QueryProvidersRequest struct{}

type QueryActiveProvidersRequest struct{}

type QueryEscrowRequest struct {
	EscrowId string `json:"escrow_id"`
}

type QueryEscrowsByProviderRequest struct {
	Provider string `json:"provider"`
}

type QueryEscrowByJobRequest struct {
	JobId string `json:"job_id"`
}

type QueryDisputeRequest struct {
	DisputeId string `json:"dispute_id"`
}

type QueryDisputesByProviderRequest struct {
	Provider string `json:"provider"`
}

type QueryParamsRequest struct{}

type QueryStatsRequest struct{}

// Query response types

type QueryProviderResponse struct {
	Provider ProviderExtended `json:"provider"`
}

type QueryProvidersResponse struct {
	Providers []ProviderExtended `json:"providers"`
}

type QueryActiveProvidersResponse struct {
	Providers []string `json:"providers"`
}

type QueryEscrowResponse struct {
	Escrow EscrowExtended `json:"escrow"`
}

type QueryEscrowsByProviderResponse struct {
	Escrows []EscrowExtended `json:"escrows"`
}

type QueryDisputeResponse struct {
	Dispute Dispute `json:"dispute"`
}

type QueryDisputesByProviderResponse struct {
	Disputes []Dispute `json:"disputes"`
}

type QueryParamsResponse struct {
	Params Params `json:"params"`
}

type QueryStatsResponse struct {
	TotalStaked     string `json:"total_staked"`
	TotalEscrowed   string `json:"total_escrowed"`
	ActiveProviders uint64 `json:"active_providers"`
	TotalProviders  uint64 `json:"total_providers"`
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct{}

func (UnimplementedQueryServer) Provider(ctx context.Context, req *QueryProviderRequest) (*QueryProviderResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Providers(ctx context.Context, req *QueryProvidersRequest) (*QueryProvidersResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) ActiveProviders(ctx context.Context, req *QueryActiveProvidersRequest) (*QueryActiveProvidersResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Escrow(ctx context.Context, req *QueryEscrowRequest) (*QueryEscrowResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) EscrowsByProvider(ctx context.Context, req *QueryEscrowsByProviderRequest) (*QueryEscrowsByProviderResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) EscrowByJob(ctx context.Context, req *QueryEscrowByJobRequest) (*QueryEscrowResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Dispute(ctx context.Context, req *QueryDisputeRequest) (*QueryDisputeResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) DisputesByProvider(ctx context.Context, req *QueryDisputesByProviderRequest) (*QueryDisputesByProviderResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Stats(ctx context.Context, req *QueryStatsRequest) (*QueryStatsResponse, error) {
	return nil, nil
}

// _Msg_serviceDesc is the gRPC service descriptor for Msg service.
var _Msg_serviceDesc = _Msg_serviceDesc_placeholder{}

type _Msg_serviceDesc_placeholder struct{}

// Query service descriptor placeholder
var _Query_serviceDesc = _Query_serviceDesc_placeholder{}

type _Query_serviceDesc_placeholder struct{}