package main

import (
	"encoding/json"
	"fmt"
	"go/pushserver"
	"go/worker"
	"os"
	"os/signal"
	"syscall"
)

type PushJob struct {
	service     *pushserver.Service
	deviceToken string
	headers     *pushserver.Headers
	payload     []byte
}

func (pj PushJob) Run() {
	pj.service.Push(pj.deviceToken, pj.headers, pj.payload)
}

func main() {
	jobQueue := make(chan worker.Job, 4028)
	dispatch := worker.NewDispatch(100, jobQueue)
	defer func() {
		dispatch.Stop()
	}()

	dispatch.Run()

	// Mechanical domain.
	errc := make(chan error)

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	const name = "./cert.p12"
	cert, err := pushserver.Load(name, "")
	if err != nil {
		fmt.Printf("error load certificate: %s\n", err.Error())
	}
	client, err := pushserver.NewClient(cert)
	if err != nil {
		fmt.Printf("error load new tls client: %s\n", err.Error())
	}

	service := pushserver.NewService(client, pushserver.Development2197)
	// service := pushserver.NewService(http.DefaultClient, pushserver.Development2197)
	headers := &pushserver.Headers{}
	aps := pushserver.APS{
		Alert: pushserver.Alert{
			Title:    "Message",
			Subtitle: "This is important",
			Body:     "Message received from Bob",
		},
	}
	apsmap := aps.Map()
	apsmap["ext"] = []string{"a", "b"}
	payload, err := json.Marshal(apsmap)

	deviceToken := "c2732227a1d8021cfaf781d71fb2f908c61f5861079a00954a5453f1d0281433"

	//生成任务
	go func() {
		for i := 0; i < 5; i++ {
			jobQueue <- PushJob{
				service:     service,
				headers:     headers,
				payload:     payload,
				deviceToken: deviceToken,
			}
		}
	}()

	fmt.Printf("exit: %v", <-errc)
}
