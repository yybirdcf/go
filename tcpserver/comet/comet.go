package main

import (
	"flag"
	"fmt"
	"go/tcpserver"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		listenAddr = flag.String("listen.addr", ":12000", "tcp listen address")
		nsqdAddr   = flag.String("nsqd.addr", ":4150", "nsqd address")
	)
	flag.Parse()

	server := tcpserver.NewTCPServer(*listenAddr, *nsqdAddr)
	defer func() {
		server.Close()
	}()
	errc := make(chan error)
	go func() {
		errc <- fmt.Errorf("%s", server.Serve())
	}()
	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	fmt.Printf("exit: %v", <-errc)
}
