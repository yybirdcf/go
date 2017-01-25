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
		workerId = flag.Int64("snowflake.workerId", 1, "0 < workerId < 1024")
		nsqdAddr = flag.String("nsqd.addr", "", "nsqd address")
	)
	flag.Parse()

	d := tcpserver.NewDispatch(*workerId, *nsqdAddr)
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
