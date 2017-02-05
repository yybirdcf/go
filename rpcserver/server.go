package main

import (
	"flag"
	"fmt"
	"go/rpcserver/grpc_middleware"
	"go/rpcserver/pb"
	"io"
	"log"
	"net"

	context "golang.org/x/net/context"

	"google.golang.org/grpc"
)

type exampleServer struct {
}

func (server *exampleServer) Say(ctx context.Context, hello *pb.Hello) (*pb.World, error) {
	// panic("test panic")
	return &pb.World{
		Word: fmt.Sprintf("%s %s", hello.GetWord(), "world"),
	}, nil
}

func (server *exampleServer) SayStream(stream pb.Example_SayStreamServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		word := &pb.World{
			Word: fmt.Sprintf("%s %s", in.GetWord(), "world"),
		}

		if err := stream.Send(word); err != nil {
			return err
		}
	}
}

func main() {
	var (
		port = flag.Int("port", 3000, "example rpc server listen port")
	)
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_middleware.LogUnary(), grpc_middleware.RecoveryUnary(), grpc_middleware.AuthUnary())),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_middleware.LogStream(), grpc_middleware.RecoveryStream(), grpc_middleware.AuthStream())),
	)
	pb.RegisterExampleServer(grpcServer, &exampleServer{})
	grpcServer.Serve(lis)
}
