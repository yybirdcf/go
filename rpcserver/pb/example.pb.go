// Code generated by protoc-gen-go.
// source: example.proto
// DO NOT EDIT!

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	example.proto

It has these top-level messages:
	Hello
	World
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Hello struct {
	Word string `protobuf:"bytes,1,opt,name=word" json:"word,omitempty"`
}

func (m *Hello) Reset()                    { *m = Hello{} }
func (m *Hello) String() string            { return proto.CompactTextString(m) }
func (*Hello) ProtoMessage()               {}
func (*Hello) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Hello) GetWord() string {
	if m != nil {
		return m.Word
	}
	return ""
}

type World struct {
	Word string `protobuf:"bytes,1,opt,name=word" json:"word,omitempty"`
}

func (m *World) Reset()                    { *m = World{} }
func (m *World) String() string            { return proto.CompactTextString(m) }
func (*World) ProtoMessage()               {}
func (*World) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *World) GetWord() string {
	if m != nil {
		return m.Word
	}
	return ""
}

func init() {
	proto.RegisterType((*Hello)(nil), "pb.Hello")
	proto.RegisterType((*World)(nil), "pb.World")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Example service

type ExampleClient interface {
	Say(ctx context.Context, in *Hello, opts ...grpc.CallOption) (*World, error)
}

type exampleClient struct {
	cc *grpc.ClientConn
}

func NewExampleClient(cc *grpc.ClientConn) ExampleClient {
	return &exampleClient{cc}
}

func (c *exampleClient) Say(ctx context.Context, in *Hello, opts ...grpc.CallOption) (*World, error) {
	out := new(World)
	err := grpc.Invoke(ctx, "/pb.Example/Say", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Example service

type ExampleServer interface {
	Say(context.Context, *Hello) (*World, error)
}

func RegisterExampleServer(s *grpc.Server, srv ExampleServer) {
	s.RegisterService(&_Example_serviceDesc, srv)
}

func _Example_Say_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Hello)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExampleServer).Say(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Example/Say",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExampleServer).Say(ctx, req.(*Hello))
	}
	return interceptor(ctx, in, info, handler)
}

var _Example_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Example",
	HandlerType: (*ExampleServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Say",
			Handler:    _Example_Say_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "example.proto",
}

func init() { proto.RegisterFile("example.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 112 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0x4d, 0xad, 0x48, 0xcc,
	0x2d, 0xc8, 0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2a, 0x48, 0x52, 0x92, 0xe6,
	0x62, 0xf5, 0x48, 0xcd, 0xc9, 0xc9, 0x17, 0x12, 0xe2, 0x62, 0x29, 0xcf, 0x2f, 0x4a, 0x91, 0x60,
	0x54, 0x60, 0xd4, 0xe0, 0x0c, 0x02, 0xb3, 0x41, 0x92, 0xe1, 0xf9, 0x45, 0x39, 0x29, 0xd8, 0x24,
	0x8d, 0x34, 0xb8, 0xd8, 0x5d, 0x21, 0xc6, 0x09, 0xc9, 0x72, 0x31, 0x07, 0x27, 0x56, 0x0a, 0x71,
	0xea, 0x15, 0x24, 0xe9, 0x81, 0x4d, 0x93, 0x02, 0x33, 0xc1, 0x7a, 0x95, 0x18, 0x92, 0xd8, 0xc0,
	0xd6, 0x19, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x86, 0xfb, 0x71, 0xf0, 0x7f, 0x00, 0x00, 0x00,
}
