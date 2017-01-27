package main

import (
	"fmt"
	"go/tcpserver"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config := tcpserver.NewStoreConfig()
	d := tcpserver.NewStoreSrv(config)
	defer func() {
		d.Close()
	}()

	errc := make(chan error)
	go func() {
		d.Run()
	}()

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	fmt.Printf("exit: %v", <-errc)
}
