// Package grpcmiddleware provides a way to add functions that will be called for each gRPC, before that RPC executes.
package grpcmiddleware

import (
	"context"

	"google.golang.org/grpc"
)

type StreamMiddleware func(srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo) error

type UnaryMiddleware func(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo) (interface{}, error)

type Middleware struct {
	streamMiddleware []StreamMiddleware
	unaryMiddleware  []UnaryMiddleware
}

func (m *Middleware) AddStreamMiddleware(middleware StreamMiddleware) {
	m.streamMiddleware = append(m.streamMiddleware, middleware)
}

func (m *Middleware) AddUnaryMiddleware(middleware UnaryMiddleware) {
	m.unaryMiddleware = append(m.unaryMiddleware, middleware)
}

// StreamInterceptor method can be provided to gRPC server options as a grpc.StreamInterceptor
//
//	opts := []grpc.ServerOption{}
//	opts = append(opts, grpc.StreamInterceptor(middleware.StreamIntercept))
//	s := grpc.NewServer(opts...)
func (m *Middleware) StreamIntercept(srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {
	for _, m := range m.streamMiddleware {
		if err := m(srv, ss, info); err != nil {
			return err
		}
	}
	return handler(srv, ss)
}

// UnaryIntercept method can be provided to gRPC server options as a grpc.UnaryInterceptor
//
//	opts := []grpc.ServerOption{}
//	opts = append(opts, grpc.UnaryInterceptor(middleware.UnaryIntercept))
//	s := grpc.NewServer(opts...)
func (m *Middleware) UnaryIntercept(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	for _, m := range m.unaryMiddleware {
		if _, err := m(ctx, req, info); err != nil {
			return nil, err
		}
	}

	return handler(ctx, req)
}
