package main

import (
	"flag"
	"fmt"
	"go/rpcserver/grpc_middleware"
	"go/rpcserver/pb"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go/rpcserver/gateway/subscribe"

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
		grpcAddr = flag.String("grpc.addr", ":3000", "example rgpc server listen address")
		etcdAddr = flag.String("etcd.addr", "", "Etcd agent address, like: http://192.168.0.1:2379,http://192.168.0.2:2379")
	)
	flag.Parse()

	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	errc := make(chan error)
	machines := strings.Split(*etcdAddr, ",")
	subClient, err := subscribe.NewEtcdClient(context.Background(), machines)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	subKey := fmt.Sprintf("/rpcserver/%s", *grpcAddr)

	go func() {
		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_middleware.LogUnary(), grpc_middleware.RecoveryUnary(), grpc_middleware.AuthUnary())),
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_middleware.LogStream(), grpc_middleware.RecoveryStream(), grpc_middleware.AuthStream())),
		)
		pb.RegisterExampleServer(grpcServer, &exampleServer{})
		err = grpcServer.Serve(lis)
		errc <- err
	}()

	go func() {
		//注册服务发现
		err = subClient.Register(subKey, *grpcAddr)
		errc <- err
	}()

	defer func() {
		subClient.Deregister(subKey)
	}()

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	fmt.Printf("exit: %v", <-errc)
}
