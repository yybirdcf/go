package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go/rpcserver/gateway/subscribe"

	context "golang.org/x/net/context"
)

//api gateway
func main() {
	var (
		etcdAddr = flag.String("etcd.addr", "", "Etcd agent address, like: http://192.168.0.1:2379,http://192.168.0.2:2379")
	)

	errc := make(chan error)

	prefix := "/rpcserver/"
	machines := strings.Split(*etcdAddr, ",")
	sub, err := subscribe.NewEtcdSub(context.Background(), machines, prefix)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	sub.GetEntries()

	defer func() {
		sub.Stop()
	}()

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	fmt.Printf("exit: %v", <-errc)
}
