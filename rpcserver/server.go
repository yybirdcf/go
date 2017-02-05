package main

import (
	"flag"
	"fmt"
	"go/rpcserver/grpc_middleware"
	"go/rpcserver/pb"
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
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_middleware.LogUnary(), grpc_middleware.RecoveryUnary())),
	)
	pb.RegisterExampleServer(grpcServer, &exampleServer{})
	grpcServer.Serve(lis)
}
