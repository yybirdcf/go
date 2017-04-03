package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		path = flag.String("config", "./config.yml", "Config file path")
	)
	flag.Parse()

	cfg := GetConfig(*path)

	errc := make(chan error)

	s := NewSchedule(cfg)
	go s.Run()

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	log.Fatalln(<-errc)
}
