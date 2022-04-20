// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.3
// source: file.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// FileClient is the client API for File service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FileClient interface {
	Write(ctx context.Context, opts ...grpc.CallOption) (File_WriteClient, error)
	Read(ctx context.Context, in *wrapperspb.StringValue, opts ...grpc.CallOption) (File_ReadClient, error)
	GetInfo(ctx context.Context, in *wrapperspb.StringValue, opts ...grpc.CallOption) (*Info, error)
	SetMode(ctx context.Context, in *Mode, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type fileClient struct {
	cc grpc.ClientConnInterface
}

func NewFileClient(cc grpc.ClientConnInterface) FileClient {
	return &fileClient{cc}
}

func (c *fileClient) Write(ctx context.Context, opts ...grpc.CallOption) (File_WriteClient, error) {
	stream, err := c.cc.NewStream(ctx, &File_ServiceDesc.Streams[0], "/omega.File/Write", opts...)
	if err != nil {
		return nil, err
	}
	x := &fileWriteClient{stream}
	return x, nil
}

type File_WriteClient interface {
	Send(*Buffer) error
	CloseAndRecv() (*emptypb.Empty, error)
	grpc.ClientStream
}

type fileWriteClient struct {
	grpc.ClientStream
}

func (x *fileWriteClient) Send(m *Buffer) error {
	return x.ClientStream.SendMsg(m)
}

func (x *fileWriteClient) CloseAndRecv() (*emptypb.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(emptypb.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *fileClient) Read(ctx context.Context, in *wrapperspb.StringValue, opts ...grpc.CallOption) (File_ReadClient, error) {
	stream, err := c.cc.NewStream(ctx, &File_ServiceDesc.Streams[1], "/omega.File/Read", opts...)
	if err != nil {
		return nil, err
	}
	x := &fileReadClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type File_ReadClient interface {
	Recv() (*Buffer, error)
	grpc.ClientStream
}

type fileReadClient struct {
	grpc.ClientStream
}

func (x *fileReadClient) Recv() (*Buffer, error) {
	m := new(Buffer)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *fileClient) GetInfo(ctx context.Context, in *wrapperspb.StringValue, opts ...grpc.CallOption) (*Info, error) {
	out := new(Info)
	err := c.cc.Invoke(ctx, "/omega.File/GetInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileClient) SetMode(ctx context.Context, in *Mode, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/omega.File/SetMode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FileServer is the server API for File service.
// All implementations must embed UnimplementedFileServer
// for forward compatibility
type FileServer interface {
	Write(File_WriteServer) error
	Read(*wrapperspb.StringValue, File_ReadServer) error
	GetInfo(context.Context, *wrapperspb.StringValue) (*Info, error)
	SetMode(context.Context, *Mode) (*emptypb.Empty, error)
	mustEmbedUnimplementedFileServer()
}

// UnimplementedFileServer must be embedded to have forward compatible implementations.
type UnimplementedFileServer struct {
}

func (UnimplementedFileServer) Write(File_WriteServer) error {
	return status.Errorf(codes.Unimplemented, "method Write not implemented")
}
func (UnimplementedFileServer) Read(*wrapperspb.StringValue, File_ReadServer) error {
	return status.Errorf(codes.Unimplemented, "method Read not implemented")
}
func (UnimplementedFileServer) GetInfo(context.Context, *wrapperspb.StringValue) (*Info, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInfo not implemented")
}
func (UnimplementedFileServer) SetMode(context.Context, *Mode) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetMode not implemented")
}
func (UnimplementedFileServer) mustEmbedUnimplementedFileServer() {}

// UnsafeFileServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FileServer will
// result in compilation errors.
type UnsafeFileServer interface {
	mustEmbedUnimplementedFileServer()
}

func RegisterFileServer(s grpc.ServiceRegistrar, srv FileServer) {
	s.RegisterService(&File_ServiceDesc, srv)
}

func _File_Write_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(FileServer).Write(&fileWriteServer{stream})
}

type File_WriteServer interface {
	SendAndClose(*emptypb.Empty) error
	Recv() (*Buffer, error)
	grpc.ServerStream
}

type fileWriteServer struct {
	grpc.ServerStream
}

func (x *fileWriteServer) SendAndClose(m *emptypb.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *fileWriteServer) Recv() (*Buffer, error) {
	m := new(Buffer)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _File_Read_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(wrapperspb.StringValue)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FileServer).Read(m, &fileReadServer{stream})
}

type File_ReadServer interface {
	Send(*Buffer) error
	grpc.ServerStream
}

type fileReadServer struct {
	grpc.ServerStream
}

func (x *fileReadServer) Send(m *Buffer) error {
	return x.ServerStream.SendMsg(m)
}

func _File_GetInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(wrapperspb.StringValue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServer).GetInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/omega.File/GetInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServer).GetInfo(ctx, req.(*wrapperspb.StringValue))
	}
	return interceptor(ctx, in, info, handler)
}

func _File_SetMode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Mode)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServer).SetMode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/omega.File/SetMode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServer).SetMode(ctx, req.(*Mode))
	}
	return interceptor(ctx, in, info, handler)
}

// File_ServiceDesc is the grpc.ServiceDesc for File service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var File_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "omega.File",
	HandlerType: (*FileServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetInfo",
			Handler:    _File_GetInfo_Handler,
		},
		{
			MethodName: "SetMode",
			Handler:    _File_SetMode_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Write",
			Handler:       _File_Write_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "Read",
			Handler:       _File_Read_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "file.proto",
}
