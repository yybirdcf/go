package main

import (
	"fmt"
	"go/tcpserver"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config := tcpserver.NewCometConfig()
	server := tcpserver.NewTCPServer(config)
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
