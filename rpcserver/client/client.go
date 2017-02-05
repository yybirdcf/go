package main

import (
	"context"
	"flag"
	"fmt"
	"go/rpcserver/pb"
	"io"
	"log"

	"google.golang.org/grpc"
)

func main() {
	var (
		port = flag.Int("port", 3000, "example rpc server listen port")
	)
	flag.Parse()

	conn, err := grpc.Dial(fmt.Sprintf(":%d", *port), grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}

	client := pb.NewExampleClient(conn)
	word, err := client.Say(context.Background(), &pb.Hello{Word: "hello"})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(word.Word)

	//双向
	waitc := make(chan struct{})
	stream, err := client.SayStream(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a hello : %v", err)
			}
			log.Printf("Got message %s", in.Word)
		}
	}()
	hellos := []pb.Hello{
		pb.Hello{Word: "a"},
		pb.Hello{Word: "b"},
		pb.Hello{Word: "c"},
	}
	for _, hello := range hellos {
		if err := stream.Send(&hello); err != nil {
			log.Fatalf("Failed to send a hello: %v", err)
		}
	}
	stream.CloseSend()
	<-waitc
}
