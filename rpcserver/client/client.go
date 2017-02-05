package main

import (
	"context"
	"flag"
	"fmt"
	"go/rpcserver/pb"
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
}
