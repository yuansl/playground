// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// CdnStatisticsClient is the client API for CdnStatistics service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CdnStatisticsClient interface {
	ListCdnDomains(ctx context.Context, in *CdnDomainsRequest, opts ...grpc.CallOption) (*CdnDomainsResponse, error)
	CdnDomainsBandwidth(ctx context.Context, in *CdnStatRequest, opts ...grpc.CallOption) (*CdnStatResponse, error)
}

type cdnStatisticsClient struct {
	cc grpc.ClientConnInterface
}

func NewCdnStatisticsClient(cc grpc.ClientConnInterface) CdnStatisticsClient {
	return &cdnStatisticsClient{cc}
}

func (c *cdnStatisticsClient) ListCdnDomains(ctx context.Context, in *CdnDomainsRequest, opts ...grpc.CallOption) (*CdnDomainsResponse, error) {
	out := new(CdnDomainsResponse)
	err := c.cc.Invoke(ctx, "/CdnStatistics/ListCdnDomains", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cdnStatisticsClient) CdnDomainsBandwidth(ctx context.Context, in *CdnStatRequest, opts ...grpc.CallOption) (*CdnStatResponse, error) {
	out := new(CdnStatResponse)
	err := c.cc.Invoke(ctx, "/CdnStatistics/CdnDomainsBandwidth", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CdnStatisticsServer is the server API for CdnStatistics service.
// All implementations must embed UnimplementedCdnStatisticsServer
// for forward compatibility
type CdnStatisticsServer interface {
	ListCdnDomains(context.Context, *CdnDomainsRequest) (*CdnDomainsResponse, error)
	CdnDomainsBandwidth(context.Context, *CdnStatRequest) (*CdnStatResponse, error)
	mustEmbedUnimplementedCdnStatisticsServer()
}

// UnimplementedCdnStatisticsServer must be embedded to have forward compatible implementations.
type UnimplementedCdnStatisticsServer struct {
}

func (UnimplementedCdnStatisticsServer) ListCdnDomains(context.Context, *CdnDomainsRequest) (*CdnDomainsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListCdnDomains not implemented")
}
func (UnimplementedCdnStatisticsServer) CdnDomainsBandwidth(context.Context, *CdnStatRequest) (*CdnStatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CdnDomainsBandwidth not implemented")
}
func (UnimplementedCdnStatisticsServer) mustEmbedUnimplementedCdnStatisticsServer() {}

// UnsafeCdnStatisticsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CdnStatisticsServer will
// result in compilation errors.
type UnsafeCdnStatisticsServer interface {
	mustEmbedUnimplementedCdnStatisticsServer()
}

func RegisterCdnStatisticsServer(s grpc.ServiceRegistrar, srv CdnStatisticsServer) {
	s.RegisterService(&CdnStatistics_ServiceDesc, srv)
}

func _CdnStatistics_ListCdnDomains_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CdnDomainsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CdnStatisticsServer).ListCdnDomains(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/CdnStatistics/ListCdnDomains",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CdnStatisticsServer).ListCdnDomains(ctx, req.(*CdnDomainsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CdnStatistics_CdnDomainsBandwidth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CdnStatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CdnStatisticsServer).CdnDomainsBandwidth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/CdnStatistics/CdnDomainsBandwidth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CdnStatisticsServer).CdnDomainsBandwidth(ctx, req.(*CdnStatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CdnStatistics_ServiceDesc is the grpc.ServiceDesc for CdnStatistics service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CdnStatistics_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "CdnStatistics",
	HandlerType: (*CdnStatisticsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListCdnDomains",
			Handler:    _CdnStatistics_ListCdnDomains_Handler,
		},
		{
			MethodName: "CdnDomainsBandwidth",
			Handler:    _CdnStatistics_CdnDomainsBandwidth_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "fscdn.proto",
}
