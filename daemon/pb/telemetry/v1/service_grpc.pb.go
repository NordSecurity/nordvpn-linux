// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.6
// source: service.proto

// Telemetry v1 defines service and message structure used to transmit client
// system metadata (e.g. OS, architecture, display protocol, etc.) to the daemon

package telemetrypb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	TelemetryService_SetDesktopEnvironment_FullMethodName = "/telemetry.v1.TelemetryService/SetDesktopEnvironment"
	TelemetryService_SetDisplayProtocol_FullMethodName    = "/telemetry.v1.TelemetryService/SetDisplayProtocol"
)

// TelemetryServiceClient is the client API for TelemetryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TelemetryServiceClient interface {
	// Reports the desktop environment name (e.g. gnome, kde, unity)
	SetDesktopEnvironment(ctx context.Context, in *DesktopEnvironmentRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Reports the active display protocol (e.g. x11, wayland)
	SetDisplayProtocol(ctx context.Context, in *DisplayProtocolRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type telemetryServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTelemetryServiceClient(cc grpc.ClientConnInterface) TelemetryServiceClient {
	return &telemetryServiceClient{cc}
}

func (c *telemetryServiceClient) SetDesktopEnvironment(ctx context.Context, in *DesktopEnvironmentRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, TelemetryService_SetDesktopEnvironment_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *telemetryServiceClient) SetDisplayProtocol(ctx context.Context, in *DisplayProtocolRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, TelemetryService_SetDisplayProtocol_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TelemetryServiceServer is the server API for TelemetryService service.
// All implementations must embed UnimplementedTelemetryServiceServer
// for forward compatibility.
type TelemetryServiceServer interface {
	// Reports the desktop environment name (e.g. gnome, kde, unity)
	SetDesktopEnvironment(context.Context, *DesktopEnvironmentRequest) (*emptypb.Empty, error)
	// Reports the active display protocol (e.g. x11, wayland)
	SetDisplayProtocol(context.Context, *DisplayProtocolRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedTelemetryServiceServer()
}

// UnimplementedTelemetryServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTelemetryServiceServer struct{}

func (UnimplementedTelemetryServiceServer) SetDesktopEnvironment(context.Context, *DesktopEnvironmentRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetDesktopEnvironment not implemented")
}
func (UnimplementedTelemetryServiceServer) SetDisplayProtocol(context.Context, *DisplayProtocolRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetDisplayProtocol not implemented")
}
func (UnimplementedTelemetryServiceServer) mustEmbedUnimplementedTelemetryServiceServer() {}
func (UnimplementedTelemetryServiceServer) testEmbeddedByValue()                          {}

// UnsafeTelemetryServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TelemetryServiceServer will
// result in compilation errors.
type UnsafeTelemetryServiceServer interface {
	mustEmbedUnimplementedTelemetryServiceServer()
}

func RegisterTelemetryServiceServer(s grpc.ServiceRegistrar, srv TelemetryServiceServer) {
	// If the following call pancis, it indicates UnimplementedTelemetryServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TelemetryService_ServiceDesc, srv)
}

func _TelemetryService_SetDesktopEnvironment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DesktopEnvironmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TelemetryServiceServer).SetDesktopEnvironment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TelemetryService_SetDesktopEnvironment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TelemetryServiceServer).SetDesktopEnvironment(ctx, req.(*DesktopEnvironmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TelemetryService_SetDisplayProtocol_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DisplayProtocolRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TelemetryServiceServer).SetDisplayProtocol(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TelemetryService_SetDisplayProtocol_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TelemetryServiceServer).SetDisplayProtocol(ctx, req.(*DisplayProtocolRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TelemetryService_ServiceDesc is the grpc.ServiceDesc for TelemetryService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TelemetryService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "telemetry.v1.TelemetryService",
	HandlerType: (*TelemetryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetDesktopEnvironment",
			Handler:    _TelemetryService_SetDesktopEnvironment_Handler,
		},
		{
			MethodName: "SetDisplayProtocol",
			Handler:    _TelemetryService_SetDisplayProtocol_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service.proto",
}
