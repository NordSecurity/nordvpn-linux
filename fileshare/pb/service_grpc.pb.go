// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.6
// source: protobuf/fileshare/service.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Fileshare_Ping_FullMethodName                = "/filesharepb.Fileshare/Ping"
	Fileshare_Stop_FullMethodName                = "/filesharepb.Fileshare/Stop"
	Fileshare_Send_FullMethodName                = "/filesharepb.Fileshare/Send"
	Fileshare_Accept_FullMethodName              = "/filesharepb.Fileshare/Accept"
	Fileshare_Cancel_FullMethodName              = "/filesharepb.Fileshare/Cancel"
	Fileshare_List_FullMethodName                = "/filesharepb.Fileshare/List"
	Fileshare_CancelFile_FullMethodName          = "/filesharepb.Fileshare/CancelFile"
	Fileshare_SetNotifications_FullMethodName    = "/filesharepb.Fileshare/SetNotifications"
	Fileshare_PurgeTransfersUntil_FullMethodName = "/filesharepb.Fileshare/PurgeTransfersUntil"
)

// FileshareClient is the client API for Fileshare service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FileshareClient interface {
	// Ping to test connection between CLI and Fileshare daemon
	Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
	// Stop
	Stop(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
	// Send a file to a peer
	Send(ctx context.Context, in *SendRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StatusResponse], error)
	// Accept a request from another peer to send you a file
	Accept(ctx context.Context, in *AcceptRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StatusResponse], error)
	// Reject a request from another peer to send you a file
	Cancel(ctx context.Context, in *CancelRequest, opts ...grpc.CallOption) (*Error, error)
	// List all transfers
	List(ctx context.Context, in *Empty, opts ...grpc.CallOption) (grpc.ServerStreamingClient[ListResponse], error)
	// Cancel file transfer to another peer
	CancelFile(ctx context.Context, in *CancelFileRequest, opts ...grpc.CallOption) (*Error, error)
	// SetNotifications about transfer status changes
	SetNotifications(ctx context.Context, in *SetNotificationsRequest, opts ...grpc.CallOption) (*SetNotificationsResponse, error)
	// PurgeTransfersUntil provided time from fileshare implementation storage
	PurgeTransfersUntil(ctx context.Context, in *PurgeTransfersUntilRequest, opts ...grpc.CallOption) (*Error, error)
}

type fileshareClient struct {
	cc grpc.ClientConnInterface
}

func NewFileshareClient(cc grpc.ClientConnInterface) FileshareClient {
	return &fileshareClient{cc}
}

func (c *fileshareClient) Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Empty)
	err := c.cc.Invoke(ctx, Fileshare_Ping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileshareClient) Stop(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Empty)
	err := c.cc.Invoke(ctx, Fileshare_Stop_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileshareClient) Send(ctx context.Context, in *SendRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StatusResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Fileshare_ServiceDesc.Streams[0], Fileshare_Send_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[SendRequest, StatusResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Fileshare_SendClient = grpc.ServerStreamingClient[StatusResponse]

func (c *fileshareClient) Accept(ctx context.Context, in *AcceptRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StatusResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Fileshare_ServiceDesc.Streams[1], Fileshare_Accept_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[AcceptRequest, StatusResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Fileshare_AcceptClient = grpc.ServerStreamingClient[StatusResponse]

func (c *fileshareClient) Cancel(ctx context.Context, in *CancelRequest, opts ...grpc.CallOption) (*Error, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Error)
	err := c.cc.Invoke(ctx, Fileshare_Cancel_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileshareClient) List(ctx context.Context, in *Empty, opts ...grpc.CallOption) (grpc.ServerStreamingClient[ListResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Fileshare_ServiceDesc.Streams[2], Fileshare_List_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[Empty, ListResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Fileshare_ListClient = grpc.ServerStreamingClient[ListResponse]

func (c *fileshareClient) CancelFile(ctx context.Context, in *CancelFileRequest, opts ...grpc.CallOption) (*Error, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Error)
	err := c.cc.Invoke(ctx, Fileshare_CancelFile_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileshareClient) SetNotifications(ctx context.Context, in *SetNotificationsRequest, opts ...grpc.CallOption) (*SetNotificationsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SetNotificationsResponse)
	err := c.cc.Invoke(ctx, Fileshare_SetNotifications_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileshareClient) PurgeTransfersUntil(ctx context.Context, in *PurgeTransfersUntilRequest, opts ...grpc.CallOption) (*Error, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Error)
	err := c.cc.Invoke(ctx, Fileshare_PurgeTransfersUntil_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FileshareServer is the server API for Fileshare service.
// All implementations must embed UnimplementedFileshareServer
// for forward compatibility.
type FileshareServer interface {
	// Ping to test connection between CLI and Fileshare daemon
	Ping(context.Context, *Empty) (*Empty, error)
	// Stop
	Stop(context.Context, *Empty) (*Empty, error)
	// Send a file to a peer
	Send(*SendRequest, grpc.ServerStreamingServer[StatusResponse]) error
	// Accept a request from another peer to send you a file
	Accept(*AcceptRequest, grpc.ServerStreamingServer[StatusResponse]) error
	// Reject a request from another peer to send you a file
	Cancel(context.Context, *CancelRequest) (*Error, error)
	// List all transfers
	List(*Empty, grpc.ServerStreamingServer[ListResponse]) error
	// Cancel file transfer to another peer
	CancelFile(context.Context, *CancelFileRequest) (*Error, error)
	// SetNotifications about transfer status changes
	SetNotifications(context.Context, *SetNotificationsRequest) (*SetNotificationsResponse, error)
	// PurgeTransfersUntil provided time from fileshare implementation storage
	PurgeTransfersUntil(context.Context, *PurgeTransfersUntilRequest) (*Error, error)
	mustEmbedUnimplementedFileshareServer()
}

// UnimplementedFileshareServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedFileshareServer struct{}

func (UnimplementedFileshareServer) Ping(context.Context, *Empty) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedFileshareServer) Stop(context.Context, *Empty) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stop not implemented")
}
func (UnimplementedFileshareServer) Send(*SendRequest, grpc.ServerStreamingServer[StatusResponse]) error {
	return status.Errorf(codes.Unimplemented, "method Send not implemented")
}
func (UnimplementedFileshareServer) Accept(*AcceptRequest, grpc.ServerStreamingServer[StatusResponse]) error {
	return status.Errorf(codes.Unimplemented, "method Accept not implemented")
}
func (UnimplementedFileshareServer) Cancel(context.Context, *CancelRequest) (*Error, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Cancel not implemented")
}
func (UnimplementedFileshareServer) List(*Empty, grpc.ServerStreamingServer[ListResponse]) error {
	return status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedFileshareServer) CancelFile(context.Context, *CancelFileRequest) (*Error, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelFile not implemented")
}
func (UnimplementedFileshareServer) SetNotifications(context.Context, *SetNotificationsRequest) (*SetNotificationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetNotifications not implemented")
}
func (UnimplementedFileshareServer) PurgeTransfersUntil(context.Context, *PurgeTransfersUntilRequest) (*Error, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PurgeTransfersUntil not implemented")
}
func (UnimplementedFileshareServer) mustEmbedUnimplementedFileshareServer() {}
func (UnimplementedFileshareServer) testEmbeddedByValue()                   {}

// UnsafeFileshareServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FileshareServer will
// result in compilation errors.
type UnsafeFileshareServer interface {
	mustEmbedUnimplementedFileshareServer()
}

func RegisterFileshareServer(s grpc.ServiceRegistrar, srv FileshareServer) {
	// If the following call pancis, it indicates UnimplementedFileshareServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Fileshare_ServiceDesc, srv)
}

func _Fileshare_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileshareServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Fileshare_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileshareServer).Ping(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Fileshare_Stop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileshareServer).Stop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Fileshare_Stop_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileshareServer).Stop(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Fileshare_Send_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SendRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FileshareServer).Send(m, &grpc.GenericServerStream[SendRequest, StatusResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Fileshare_SendServer = grpc.ServerStreamingServer[StatusResponse]

func _Fileshare_Accept_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(AcceptRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FileshareServer).Accept(m, &grpc.GenericServerStream[AcceptRequest, StatusResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Fileshare_AcceptServer = grpc.ServerStreamingServer[StatusResponse]

func _Fileshare_Cancel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileshareServer).Cancel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Fileshare_Cancel_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileshareServer).Cancel(ctx, req.(*CancelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Fileshare_List_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FileshareServer).List(m, &grpc.GenericServerStream[Empty, ListResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Fileshare_ListServer = grpc.ServerStreamingServer[ListResponse]

func _Fileshare_CancelFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileshareServer).CancelFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Fileshare_CancelFile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileshareServer).CancelFile(ctx, req.(*CancelFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Fileshare_SetNotifications_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetNotificationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileshareServer).SetNotifications(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Fileshare_SetNotifications_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileshareServer).SetNotifications(ctx, req.(*SetNotificationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Fileshare_PurgeTransfersUntil_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PurgeTransfersUntilRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileshareServer).PurgeTransfersUntil(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Fileshare_PurgeTransfersUntil_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileshareServer).PurgeTransfersUntil(ctx, req.(*PurgeTransfersUntilRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Fileshare_ServiceDesc is the grpc.ServiceDesc for Fileshare service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Fileshare_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "filesharepb.Fileshare",
	HandlerType: (*FileshareServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Fileshare_Ping_Handler,
		},
		{
			MethodName: "Stop",
			Handler:    _Fileshare_Stop_Handler,
		},
		{
			MethodName: "Cancel",
			Handler:    _Fileshare_Cancel_Handler,
		},
		{
			MethodName: "CancelFile",
			Handler:    _Fileshare_CancelFile_Handler,
		},
		{
			MethodName: "SetNotifications",
			Handler:    _Fileshare_SetNotifications_Handler,
		},
		{
			MethodName: "PurgeTransfersUntil",
			Handler:    _Fileshare_PurgeTransfersUntil_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Send",
			Handler:       _Fileshare_Send_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Accept",
			Handler:       _Fileshare_Accept_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "List",
			Handler:       _Fileshare_List_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "protobuf/fileshare/service.proto",
}
