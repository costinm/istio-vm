// Copyright 2015 gRPC authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// An integration test service that covers all the method signature permutations
// of unary/streaming requests/responses.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: grpc/testing/benchmark_service.proto

package testing

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

const (
	BenchmarkService_UnaryCall_FullMethodName           = "/grpc.testing.BenchmarkService/UnaryCall"
	BenchmarkService_StreamingCall_FullMethodName       = "/grpc.testing.BenchmarkService/StreamingCall"
	BenchmarkService_StreamingFromClient_FullMethodName = "/grpc.testing.BenchmarkService/StreamingFromClient"
	BenchmarkService_StreamingFromServer_FullMethodName = "/grpc.testing.BenchmarkService/StreamingFromServer"
	BenchmarkService_StreamingBothWays_FullMethodName   = "/grpc.testing.BenchmarkService/StreamingBothWays"
)

// BenchmarkServiceClient is the client API for BenchmarkService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BenchmarkServiceClient interface {
	// One request followed by one response.
	// The server returns the client payload as-is.
	UnaryCall(ctx context.Context, in *SimpleRequest, opts ...grpc.CallOption) (*SimpleResponse, error)
	// Repeated sequence of one request followed by one response.
	// Should be called streaming ping-pong
	// The server returns the client payload as-is on each response
	StreamingCall(ctx context.Context, opts ...grpc.CallOption) (BenchmarkService_StreamingCallClient, error)
	// Single-sided unbounded streaming from client to server
	// The server returns the client payload as-is once the client does WritesDone
	StreamingFromClient(ctx context.Context, opts ...grpc.CallOption) (BenchmarkService_StreamingFromClientClient, error)
	// Single-sided unbounded streaming from server to client
	// The server repeatedly returns the client payload as-is
	StreamingFromServer(ctx context.Context, in *SimpleRequest, opts ...grpc.CallOption) (BenchmarkService_StreamingFromServerClient, error)
	// Two-sided unbounded streaming between server to client
	// Both sides send the content of their own choice to the other
	StreamingBothWays(ctx context.Context, opts ...grpc.CallOption) (BenchmarkService_StreamingBothWaysClient, error)
}

type benchmarkServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBenchmarkServiceClient(cc grpc.ClientConnInterface) BenchmarkServiceClient {
	return &benchmarkServiceClient{cc}
}

func (c *benchmarkServiceClient) UnaryCall(ctx context.Context, in *SimpleRequest, opts ...grpc.CallOption) (*SimpleResponse, error) {
	out := new(SimpleResponse)
	err := c.cc.Invoke(ctx, BenchmarkService_UnaryCall_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *benchmarkServiceClient) StreamingCall(ctx context.Context, opts ...grpc.CallOption) (BenchmarkService_StreamingCallClient, error) {
	stream, err := c.cc.NewStream(ctx, &BenchmarkService_ServiceDesc.Streams[0], BenchmarkService_StreamingCall_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &benchmarkServiceStreamingCallClient{stream}
	return x, nil
}

type BenchmarkService_StreamingCallClient interface {
	Send(*SimpleRequest) error
	Recv() (*SimpleResponse, error)
	grpc.ClientStream
}

type benchmarkServiceStreamingCallClient struct {
	grpc.ClientStream
}

func (x *benchmarkServiceStreamingCallClient) Send(m *SimpleRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *benchmarkServiceStreamingCallClient) Recv() (*SimpleResponse, error) {
	m := new(SimpleResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *benchmarkServiceClient) StreamingFromClient(ctx context.Context, opts ...grpc.CallOption) (BenchmarkService_StreamingFromClientClient, error) {
	stream, err := c.cc.NewStream(ctx, &BenchmarkService_ServiceDesc.Streams[1], BenchmarkService_StreamingFromClient_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &benchmarkServiceStreamingFromClientClient{stream}
	return x, nil
}

type BenchmarkService_StreamingFromClientClient interface {
	Send(*SimpleRequest) error
	CloseAndRecv() (*SimpleResponse, error)
	grpc.ClientStream
}

type benchmarkServiceStreamingFromClientClient struct {
	grpc.ClientStream
}

func (x *benchmarkServiceStreamingFromClientClient) Send(m *SimpleRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *benchmarkServiceStreamingFromClientClient) CloseAndRecv() (*SimpleResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(SimpleResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *benchmarkServiceClient) StreamingFromServer(ctx context.Context, in *SimpleRequest, opts ...grpc.CallOption) (BenchmarkService_StreamingFromServerClient, error) {
	stream, err := c.cc.NewStream(ctx, &BenchmarkService_ServiceDesc.Streams[2], BenchmarkService_StreamingFromServer_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &benchmarkServiceStreamingFromServerClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type BenchmarkService_StreamingFromServerClient interface {
	Recv() (*SimpleResponse, error)
	grpc.ClientStream
}

type benchmarkServiceStreamingFromServerClient struct {
	grpc.ClientStream
}

func (x *benchmarkServiceStreamingFromServerClient) Recv() (*SimpleResponse, error) {
	m := new(SimpleResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *benchmarkServiceClient) StreamingBothWays(ctx context.Context, opts ...grpc.CallOption) (BenchmarkService_StreamingBothWaysClient, error) {
	stream, err := c.cc.NewStream(ctx, &BenchmarkService_ServiceDesc.Streams[3], BenchmarkService_StreamingBothWays_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &benchmarkServiceStreamingBothWaysClient{stream}
	return x, nil
}

type BenchmarkService_StreamingBothWaysClient interface {
	Send(*SimpleRequest) error
	Recv() (*SimpleResponse, error)
	grpc.ClientStream
}

type benchmarkServiceStreamingBothWaysClient struct {
	grpc.ClientStream
}

func (x *benchmarkServiceStreamingBothWaysClient) Send(m *SimpleRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *benchmarkServiceStreamingBothWaysClient) Recv() (*SimpleResponse, error) {
	m := new(SimpleResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// BenchmarkServiceServer is the server API for BenchmarkService service.
// All implementations should embed UnimplementedBenchmarkServiceServer
// for forward compatibility
type BenchmarkServiceServer interface {
	// One request followed by one response.
	// The server returns the client payload as-is.
	UnaryCall(context.Context, *SimpleRequest) (*SimpleResponse, error)
	// Repeated sequence of one request followed by one response.
	// Should be called streaming ping-pong
	// The server returns the client payload as-is on each response
	StreamingCall(BenchmarkService_StreamingCallServer) error
	// Single-sided unbounded streaming from client to server
	// The server returns the client payload as-is once the client does WritesDone
	StreamingFromClient(BenchmarkService_StreamingFromClientServer) error
	// Single-sided unbounded streaming from server to client
	// The server repeatedly returns the client payload as-is
	StreamingFromServer(*SimpleRequest, BenchmarkService_StreamingFromServerServer) error
	// Two-sided unbounded streaming between server to client
	// Both sides send the content of their own choice to the other
	StreamingBothWays(BenchmarkService_StreamingBothWaysServer) error
}

// UnimplementedBenchmarkServiceServer should be embedded to have forward compatible implementations.
type UnimplementedBenchmarkServiceServer struct {
}

func (UnimplementedBenchmarkServiceServer) UnaryCall(context.Context, *SimpleRequest) (*SimpleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnaryCall not implemented")
}
func (UnimplementedBenchmarkServiceServer) StreamingCall(BenchmarkService_StreamingCallServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamingCall not implemented")
}
func (UnimplementedBenchmarkServiceServer) StreamingFromClient(BenchmarkService_StreamingFromClientServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamingFromClient not implemented")
}
func (UnimplementedBenchmarkServiceServer) StreamingFromServer(*SimpleRequest, BenchmarkService_StreamingFromServerServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamingFromServer not implemented")
}
func (UnimplementedBenchmarkServiceServer) StreamingBothWays(BenchmarkService_StreamingBothWaysServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamingBothWays not implemented")
}

// UnsafeBenchmarkServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BenchmarkServiceServer will
// result in compilation errors.
type UnsafeBenchmarkServiceServer interface {
	mustEmbedUnimplementedBenchmarkServiceServer()
}

func RegisterBenchmarkServiceServer(s grpc.ServiceRegistrar, srv BenchmarkServiceServer) {
	s.RegisterService(&BenchmarkService_ServiceDesc, srv)
}

func _BenchmarkService_UnaryCall_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SimpleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BenchmarkServiceServer).UnaryCall(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BenchmarkService_UnaryCall_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BenchmarkServiceServer).UnaryCall(ctx, req.(*SimpleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BenchmarkService_StreamingCall_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BenchmarkServiceServer).StreamingCall(&benchmarkServiceStreamingCallServer{stream})
}

type BenchmarkService_StreamingCallServer interface {
	Send(*SimpleResponse) error
	Recv() (*SimpleRequest, error)
	grpc.ServerStream
}

type benchmarkServiceStreamingCallServer struct {
	grpc.ServerStream
}

func (x *benchmarkServiceStreamingCallServer) Send(m *SimpleResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *benchmarkServiceStreamingCallServer) Recv() (*SimpleRequest, error) {
	m := new(SimpleRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _BenchmarkService_StreamingFromClient_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BenchmarkServiceServer).StreamingFromClient(&benchmarkServiceStreamingFromClientServer{stream})
}

type BenchmarkService_StreamingFromClientServer interface {
	SendAndClose(*SimpleResponse) error
	Recv() (*SimpleRequest, error)
	grpc.ServerStream
}

type benchmarkServiceStreamingFromClientServer struct {
	grpc.ServerStream
}

func (x *benchmarkServiceStreamingFromClientServer) SendAndClose(m *SimpleResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *benchmarkServiceStreamingFromClientServer) Recv() (*SimpleRequest, error) {
	m := new(SimpleRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _BenchmarkService_StreamingFromServer_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SimpleRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BenchmarkServiceServer).StreamingFromServer(m, &benchmarkServiceStreamingFromServerServer{stream})
}

type BenchmarkService_StreamingFromServerServer interface {
	Send(*SimpleResponse) error
	grpc.ServerStream
}

type benchmarkServiceStreamingFromServerServer struct {
	grpc.ServerStream
}

func (x *benchmarkServiceStreamingFromServerServer) Send(m *SimpleResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _BenchmarkService_StreamingBothWays_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BenchmarkServiceServer).StreamingBothWays(&benchmarkServiceStreamingBothWaysServer{stream})
}

type BenchmarkService_StreamingBothWaysServer interface {
	Send(*SimpleResponse) error
	Recv() (*SimpleRequest, error)
	grpc.ServerStream
}

type benchmarkServiceStreamingBothWaysServer struct {
	grpc.ServerStream
}

func (x *benchmarkServiceStreamingBothWaysServer) Send(m *SimpleResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *benchmarkServiceStreamingBothWaysServer) Recv() (*SimpleRequest, error) {
	m := new(SimpleRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// BenchmarkService_ServiceDesc is the grpc.ServiceDesc for BenchmarkService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BenchmarkService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.testing.BenchmarkService",
	HandlerType: (*BenchmarkServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UnaryCall",
			Handler:    _BenchmarkService_UnaryCall_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamingCall",
			Handler:       _BenchmarkService_StreamingCall_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "StreamingFromClient",
			Handler:       _BenchmarkService_StreamingFromClient_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "StreamingFromServer",
			Handler:       _BenchmarkService_StreamingFromServer_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "StreamingBothWays",
			Handler:       _BenchmarkService_StreamingBothWays_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "grpc/testing/benchmark_service.proto",
}
